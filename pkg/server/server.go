package server

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	log "github.com/sirupsen/logrus"
	"gopkg.in/fatih/set.v0"

	cc "airman.com/airfk/pkg/codec"
	ts "airman.com/airfk/pkg/types"
)

const MetadataApi = "rpc"

const (
	serviceMethodSeparator   = "_"
	subscribeMethodSuffix    = "_subscribe"
	unsubscribeMethodSuffix  = "_unsubscribe"
	notificationMethodSuffix = "_subscription"
)

// CodecOption specifies which type of messages this cc supports
type CodecOption int

const (
	// OptionMethodInvocation is an indication that the cc supports RPC method calls
	OptionMethodInvocation CodecOption = 1 << iota

	// OptionSubscriptions is an indication that the cc suports RPC notifications
	OptionSubscriptions = 1 << iota // support pub sub
)

// callback is a method callback which was registered in the server
type Callback struct {
	Rcvr        reflect.Value  // receiver of method
	Method      reflect.Method // callback
	ArgTypes    []reflect.Type // input argument types
	HasCtx      bool           // method's first argument is a context (not included in argTypes)
	ErrPos      int            // err return idx, of -1 when method cannot return error
	IsSubscribe bool           // indication if the callback is a subscription
}

type Callbacks map[string]*Callback     // collection of RPC Callbacks
type Subscriptions map[string]*Callback // collection of subscription Callbacks

// service represents a registered object
type Service struct {
	Name          string        // name for service
	Typ           reflect.Type  // receiver type
	Callbacks     Callbacks     // registered handlers
	Subscriptions Subscriptions // available Subscriptions/notifications
}

// ServerRequest is an incoming request
type ServerRequest struct {
	Id            interface{}
	Svcname       string
	Callb         *Callback
	Args          []reflect.Value
	IsUnsubscribe bool
	Err           ts.Error
}

type ServiceRegistry map[string]*Service // collection of services

// Server represents a RPC server
type Server struct {
	Services ServiceRegistry

	Run      int32
	CodecsMu sync.Mutex
	Codecs   *set.Set
}

// NewServer will create a new server instance with no registered handlers.
func NewServer() *Server {
	server := &Server{
		Services: make(ServiceRegistry),
		Codecs:   set.New(),
		Run:      1,
	}

	// register a default service which will provide meta information about the RPC service such as the services and
	// methods it offers.
	rpcService := &RPCService{server}
	server.RegisterName(MetadataApi, rpcService)

	return server
}

// RPCService gives meta information about the server.
// e.g. gives information about the loaded modules.
type RPCService struct {
	server *Server
}

// Modules returns the list of RPC services with their version number
func (s *RPCService) Modules() map[string]string {
	modules := make(map[string]string)
	for name := range s.server.Services {
		modules[name] = "1.0"
	}
	return modules
}

// RegisterName will create a service for the given rcvr type under the given name. When no methods on the given rcvr
// match the criteria to be either a RPC method or a subscription an error is returned. Otherwise a new service is
// created and added to the service collection this server instance serves.
func (s *Server) RegisterName(name string, rcvr interface{}) error {
	if s.Services == nil {
		s.Services = make(ServiceRegistry)
	}

	svc := new(Service)
	svc.Typ = reflect.TypeOf(rcvr)
	rcvrVal := reflect.ValueOf(rcvr)

	if name == "" {
		return fmt.Errorf("no service name for type %s", svc.Typ.String())
	}
	if !isExported(reflect.Indirect(rcvrVal).Type().Name()) {
		return fmt.Errorf("%s is not exported", reflect.Indirect(rcvrVal).Type().Name())
	}

	methods, subscriptions := suitableCallbacks(rcvrVal, svc.Typ)

	// already a previous service register under given sname, merge methods/Subscriptions
	if regsvc, present := s.Services[name]; present {
		if len(methods) == 0 && len(subscriptions) == 0 {
			return fmt.Errorf("Service %T doesn't have any suitable methods/Subscriptions to expose", rcvr)
		}
		for _, m := range methods {
			regsvc.Callbacks[formatName(m.Method.Name)] = m
		}
		for _, s := range subscriptions {
			regsvc.Subscriptions[formatName(s.Method.Name)] = s
		}
		return nil
	}

	svc.Name = name
	svc.Callbacks, svc.Subscriptions = methods, subscriptions

	if len(svc.Callbacks) == 0 && len(svc.Subscriptions) == 0 {
		return fmt.Errorf("Service %T doesn't have any suitable methods/Subscriptions to expose", rcvr)
	}

	s.Services[svc.Name] = svc
	return nil
}

// serveRequest will reads requests from the cc, calls the RPC callback and
// writes the response to the given cc.
//
// If singleShot is true it will process a single request, otherwise it will handle
// requests until the cc returns an error when reading a request (in most cases
// an EOF). It executes requests in parallel when singleShot is false.
func (s *Server) serveRequest(ctx context.Context, cc cc.ServerCodec, singleShot bool, options CodecOption) error {
	var pend sync.WaitGroup

	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Error(string(buf))
		}
		s.CodecsMu.Lock()
		s.Codecs.Remove(cc)
		s.CodecsMu.Unlock()
	}()

	//	ctx, cancel := context.WithCancel(context.Background())
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// if the cc supports notification include a notifier that Callbacks can use
	// to send notification to clients. It is tied to the cc/connection. If the
	// connection is closed the notifier will stop and cancels all active Subscriptions.
	if options&OptionSubscriptions == OptionSubscriptions {
		ctx = context.WithValue(ctx, notifierKey{}, newNotifier(cc))
	}
	s.CodecsMu.Lock()
	if atomic.LoadInt32(&s.Run) != 1 { // server stopped
		s.CodecsMu.Unlock()
		return &ts.ShutdownError{}
	}
	s.Codecs.Add(cc)
	s.CodecsMu.Unlock()

	// test if the server is ordered to stop
	for atomic.LoadInt32(&s.Run) == 1 {
		reqs, batch, err := s.readRequest(cc)
		if err != nil {
			// If a parsing error occurred, send an error
			if err.Error() != "EOF" {
				log.Debug(fmt.Sprintf("read error %v\n", err))
				cc.Write(cc.CreateErrorResponse(nil, err))
			}
			// ts.Error or end of stream, wait for requests and tear down
			pend.Wait()
			return nil
		}

		// check if server is ordered to shutdown and return an error
		// telling the client that his request failed.
		if atomic.LoadInt32(&s.Run) != 1 {
			err = &ts.ShutdownError{}
			if batch {
				resps := make([]interface{}, len(reqs))
				for i, r := range reqs {
					resps[i] = cc.CreateErrorResponse(&r.Id, err)
				}
				cc.Write(resps)
			} else {
				cc.Write(cc.CreateErrorResponse(&reqs[0].Id, err))
			}
			return nil
		}
		// If a single shot request is executing, run and return immediately
		if singleShot {
			if batch {
				s.execBatch(ctx, cc, reqs)
			} else {
				s.exec(ctx, cc, reqs[0])
			}
			return nil
		}
		// For multi-shot connections, start a goroutine to serve and loop back
		pend.Add(1)

		go func(reqs []*ServerRequest, batch bool) {
			defer pend.Done()
			if batch {
				s.execBatch(ctx, cc, reqs)
			} else {
				s.exec(ctx, cc, reqs[0])
			}
		}(reqs, batch)
	}
	return nil
}

// ServeCodec reads incoming requests from cc, calls the appropriate callback and writes the
// response back using the given cc. It will block until the cc is closed or the server is
// stopped. In either case the cc is closed.
func (s *Server) ServeCodec(cc cc.ServerCodec, options CodecOption) {
	defer cc.Close()
	s.serveRequest(context.Background(), cc, false, options)
}

// ServeSingleRequest reads and processes a single RPC request from the given cc. It will not
// close the cc unless a non-recoverable error has occurred. Note, this method will return after
// a single request has been processed!
func (s *Server) ServeSingleRequest(ctx context.Context, cc cc.ServerCodec, options CodecOption) {
	s.serveRequest(ctx, cc, true, options)
}

// Stop will stop reading new requests, wait for stopPendingRequestTimeout to allow pending requests to finish,
// close all Codecs which will cancel pending requests/Subscriptions.
func (s *Server) Stop() {
	if atomic.CompareAndSwapInt32(&s.Run, 1, 0) {
		log.Debug("RPC Server shutdown initiatied")
		s.CodecsMu.Lock()
		defer s.CodecsMu.Unlock()
		s.Codecs.Each(func(c interface{}) bool {
			c.(cc.ServerCodec).Close()
			return true
		})
	}
}

// createSubscription will call the subscription callback and returns the subscription id or error.
func (s *Server) createSubscription(ctx context.Context, c cc.ServerCodec, req *ServerRequest) (ID, error) {
	// subscription have as first argument the context following optional arguments
	args := []reflect.Value{req.Callb.Rcvr, reflect.ValueOf(ctx)}
	args = append(args, req.Args...)
	reply := req.Callb.Method.Func.Call(args)

	if !reply[1].IsNil() { // subscription creation failed
		return "", reply[1].Interface().(error)
	}

	return reply[0].Interface().(*Subscription).ID, nil
}

// handle executes a request and returns the response from the callback.
func (s *Server) handle(ctx context.Context, cc cc.ServerCodec, req *ServerRequest) (interface{}, func()) {
	if req.Err != nil {
		return cc.CreateErrorResponse(&req.Id, req.Err), nil
	}

	if req.IsUnsubscribe { // cancel subscription, first param must be the subscription id
		if len(req.Args) >= 1 && req.Args[0].Kind() == reflect.String {
			notifier, supported := NotifierFromContext(ctx)
			if !supported { // interface doesn't support Subscriptions (e.g. http)
				return cc.CreateErrorResponse(&req.Id, &ts.CallbackError{ErrNotificationsUnsupported.Error()}), nil
			}

			subid := ID(req.Args[0].String())
			if err := notifier.unsubscribe(subid); err != nil {
				return cc.CreateErrorResponse(&req.Id, &ts.CallbackError{err.Error()}), nil
			}

			return cc.CreateResponse(req.Id, true), nil
		}
		return cc.CreateErrorResponse(&req.Id, &ts.InvalidParamsError{"Expected subscription id as first argument"}), nil
	}

	if req.Callb.IsSubscribe {
		subid, err := s.createSubscription(ctx, cc, req)
		if err != nil {
			return cc.CreateErrorResponse(&req.Id, &ts.CallbackError{err.Error()}), nil
		}

		// active the subscription after the sub id was successfully sent to the client
		activateSub := func() {
			notifier, _ := NotifierFromContext(ctx)
			notifier.activate(subid, req.Svcname)
		}

		return cc.CreateResponse(req.Id, subid), activateSub
	}

	// regular RPC call, prepare arguments
	if len(req.Args) != len(req.Callb.ArgTypes) {
		rpcErr := &ts.InvalidParamsError{fmt.Sprintf("%s%s%s expects %d parameters, got %d",
			req.Svcname, serviceMethodSeparator, req.Callb.Method.Name,
			len(req.Callb.ArgTypes), len(req.Args))}
		return cc.CreateErrorResponse(&req.Id, rpcErr), nil
	}

	arguments := []reflect.Value{req.Callb.Rcvr}
	if req.Callb.HasCtx {
		arguments = append(arguments, reflect.ValueOf(ctx))
	}
	if len(req.Args) > 0 {
		arguments = append(arguments, req.Args...)
	}

	// execute RPC method and return result
	reply := req.Callb.Method.Func.Call(arguments)
	if len(reply) == 0 {
		return cc.CreateResponse(req.Id, nil), nil
	}
	if req.Callb.ErrPos >= 0 { // test if method returned an error
		if !reply[req.Callb.ErrPos].IsNil() {
			e := reply[req.Callb.ErrPos].Interface().(error)
			res := cc.CreateErrorResponse(&req.Id, &ts.CallbackError{e.Error()})
			return res, nil
		}
	}
	return cc.CreateResponse(req.Id, reply[0].Interface()), nil
}

// exec executes the given request and writes the result back using the cc.
func (s *Server) exec(ctx context.Context, cc cc.ServerCodec, req *ServerRequest) {
	var response interface{}
	var callback func()
	if req.Err != nil {
		response = cc.CreateErrorResponse(&req.Id, req.Err)
	} else {
		response, callback = s.handle(ctx, cc, req)
	}

	if err := cc.Write(response); err != nil {
		log.Error(fmt.Sprintf("%v\n", err))
		cc.Close()
	}

	// when request was a subscribe request this allows these Subscriptions to be actived
	if callback != nil {
		callback()
	}
}

// execBatch executes the given requests and writes the result back using the cc.
// It will only write the response back when the last request is processed.
func (s *Server) execBatch(ctx context.Context, cc cc.ServerCodec, requests []*ServerRequest) {
	responses := make([]interface{}, len(requests))
	var Callbacks []func()
	for i, req := range requests {
		if req.Err != nil {
			responses[i] = cc.CreateErrorResponse(&req.Id, req.Err)
		} else {
			var callback func()
			if responses[i], callback = s.handle(ctx, cc, req); callback != nil {
				Callbacks = append(Callbacks, callback)
			}
		}
	}

	if err := cc.Write(responses); err != nil {
		log.Error(fmt.Sprintf("%v\n", err))
		cc.Close()
	}

	// when request holds one of more subscribe requests this allows these Subscriptions to be activated
	for _, c := range Callbacks {
		c()
	}
}

// readRequest requests the next (batch) request from the cc. It will return the collection
// of requests, an indication if the request was a batch, the invalid request identifier and an
// error when the request could not be read/parsed.
func (s *Server) readRequest(cc cc.ServerCodec) ([]*ServerRequest, bool, ts.Error) {
	reqs, batch, err := cc.ReadRequestHeaders()
	if err != nil {
		return nil, batch, err
	}

	requests := make([]*ServerRequest, len(reqs))

	// verify requests
	for i, r := range reqs {
		var ok bool
		var svc *Service

		if r.Err != nil {
			requests[i] = &ServerRequest{Id: r.Id, Err: r.Err}
			continue
		}

		if r.IsPubSub && strings.HasSuffix(r.Method, unsubscribeMethodSuffix) {
			requests[i] = &ServerRequest{Id: r.Id, IsUnsubscribe: true}
			argTypes := []reflect.Type{reflect.TypeOf("")} // expect subscription id as first arg
			if args, err := cc.ParseRequestArguments(argTypes, r.Params); err == nil {
				requests[i].Args = args
			} else {
				requests[i].Err = &ts.InvalidParamsError{err.Error()}
			}
			continue
		}

		if svc, ok = s.Services[r.Service]; !ok { // rpc method isn't available
			requests[i] = &ServerRequest{Id: r.Id, Err: &ts.MethodNotFoundError{r.Service, r.Method}}
			continue
		}

		if r.IsPubSub { // eth_subscribe, r.Method contains the subscription method name
			if callb, ok := svc.Subscriptions[r.Method]; ok {
				requests[i] = &ServerRequest{Id: r.Id, Svcname: svc.Name, Callb: callb}
				if r.Params != nil && len(callb.ArgTypes) > 0 {
					argTypes := []reflect.Type{reflect.TypeOf("")}
					argTypes = append(argTypes, callb.ArgTypes...)
					if args, err := cc.ParseRequestArguments(argTypes, r.Params); err == nil {
						requests[i].Args = args[1:] // first one is service.method name which isn't an actual argument
					} else {
						requests[i].Err = &ts.InvalidParamsError{err.Error()}
					}
				}
			} else {
				requests[i] = &ServerRequest{Id: r.Id, Err: &ts.MethodNotFoundError{r.Service, r.Method}}
			}
			continue
		}

		if callb, ok := svc.Callbacks[r.Method]; ok { // lookup RPC method
			requests[i] = &ServerRequest{Id: r.Id, Svcname: svc.Name, Callb: callb}
			if r.Params != nil && len(callb.ArgTypes) > 0 {
				if args, err := cc.ParseRequestArguments(callb.ArgTypes, r.Params); err == nil {
					requests[i].Args = args
				} else {
					requests[i].Err = &ts.InvalidParamsError{err.Error()}
				}
			}
			continue
		}

		requests[i] = &ServerRequest{Id: r.Id, Err: &ts.MethodNotFoundError{r.Service, r.Method}}
	}

	return requests, batch, nil
}
