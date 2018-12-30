package codec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"sync"

	ts "airman.com/airfk/pkg/types"
	log "github.com/sirupsen/logrus"
)

const (
	jsonrpcVersion           = "2.0"
	serviceMethodSeparator   = "_"
	subscribeMethodSuffix    = "_subscribe"
	unsubscribeMethodSuffix  = "_unsubscribe"
	notificationMethodSuffix = "_subscription"
)

type JsonRequest struct {
	Method  string          `json:"method"`
	Version string          `json:"jsonrpc"`
	Id      json.RawMessage `json:"id,omitempty"`
	Payload json.RawMessage `json:"params,omitempty"`
}

type JsonSuccessResponse struct {
	Version string      `json:"jsonrpc"`
	Id      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result"`
}

type JsonError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type JsonErrResponse struct {
	Version string      `json:"jsonrpc"`
	Id      interface{} `json:"id,omitempty"`
	Error   JsonError   `json:"error"`
}

type JsonSubscription struct {
	Subscription string      `json:"subscription"`
	Result       interface{} `json:"result,omitempty"`
}

type JsonNotification struct {
	Version string           `json:"jsonrpc"`
	Method  string           `json:"method"`
	Params  JsonSubscription `json:"params"`
}

// JsonCodec reads and writes JSON-RPC messages to the underlying connection. It
// also has support for parsing arguments and serializing (result) objects.
type JsonCodec struct {
	closer sync.Once                 // close closed channel once
	closed chan interface{}          // closed on Close
	decMu  sync.Mutex                // guards the decoder
	decode func(v interface{}) error // decoder to allow multiple transports
	encMu  sync.Mutex                // guards the encoder
	encode func(v interface{}) error // encoder to allow multiple transports
	rw     io.ReadWriteCloser        // connection
}

func (err *JsonError) Error() string {
	if err.Message == "" {
		return fmt.Sprintf("json-rpc error %d", err.Code)
	}
	return err.Message
}

func (err *JsonError) ErrorCode() int {
	return err.Code
}

// NewCodec creates a new RPC server codec with support for JSON-RPC 2.0 based
// on explicitly given encoding and decoding methods.
func NewCodec(rwc io.ReadWriteCloser, encode, decode func(v interface{}) error) ServerCodec {
	return &JsonCodec{
		closed: make(chan interface{}),
		encode: encode,
		decode: decode,
		rw:     rwc,
	}
}

// NewJSONCodec creates a new RPC server codec with support for JSON-RPC 2.0.
func NewJSONCodec(rwc io.ReadWriteCloser) ServerCodec {
	enc := json.NewEncoder(rwc)
	dec := json.NewDecoder(rwc)
	dec.UseNumber()

	return &JsonCodec{
		closed: make(chan interface{}),
		encode: enc.Encode,
		decode: dec.Decode,
		rw:     rwc,
	}
}

// isBatch returns true when the first non-whitespace characters is '['
func isBatch(msg json.RawMessage) bool {
	for _, c := range msg {
		// skip insignificant whitespace (http://www.ietf.org/rfc/rfc4627.txt)
		if c == 0x20 || c == 0x09 || c == 0x0a || c == 0x0d {
			continue
		}
		return c == '['
	}
	return false
}

// ReadRequestHeaders will read new requests without parsing the arguments. It will
// return a collection of requests, an indication if these requests are in batch
// form or an error when the incoming message could not be read/parsed.
func (c *JsonCodec) ReadRequestHeaders() ([]ts.RpcRequest, bool, ts.Error) {
	c.decMu.Lock()
	defer c.decMu.Unlock()

	var incomingMsg json.RawMessage
	if err := c.decode(&incomingMsg); err != nil {
		return nil, false, &ts.InvalidRequestError{Message:err.Error()}
	}
	if isBatch(incomingMsg) {
		return parseBatchRequest(incomingMsg)
	}
	return parseRequest(incomingMsg)
}

// checkReqId returns an error when the given reqId isn't valid for RPC method calls.
// valid id's are strings, numbers or null
func checkReqId(reqId json.RawMessage) error {
	if len(reqId) == 0 {
		return fmt.Errorf("missing request id")
	}
	if _, err := strconv.ParseFloat(string(reqId), 64); err == nil {
		return nil
	}
	var str string
	if err := json.Unmarshal(reqId, &str); err == nil {
		return nil
	}
	return fmt.Errorf("invalid request id")
}

// parseRequest will parse a single request from the given RawMessage. It will return
// the parsed request, an indication if the request was a batch or an error when
// the request could not be parsed.
func parseRequest(incomingMsg json.RawMessage) ([]ts.RpcRequest, bool, ts.Error) {
	var in JsonRequest
	if err := json.Unmarshal(incomingMsg, &in); err != nil {
		return nil, false, &ts.InvalidMessageError{Message:err.Error()}
	}

	if err := checkReqId(in.Id); err != nil {
		return nil, false, &ts.InvalidMessageError{Message:err.Error()}
	}

	// subscribe are special, they will always use `subscribeMethod` as first param in the payload
	if strings.HasSuffix(in.Method, subscribeMethodSuffix) {
		reqs := []ts.RpcRequest{{Id: &in.Id, IsPubSub: true}}
		if len(in.Payload) > 0 {
			// first param must be subscription name
			var subscribeMethod [1]string
			if err := json.Unmarshal(in.Payload, &subscribeMethod); err != nil {
				log.Debug(fmt.Sprintf("Unable to parse subscription method: %v\n", err))
				return nil, false, &ts.InvalidRequestError{Message:"Unable to parse subscription request"}
			}

			reqs[0].Service, reqs[0].Method = strings.TrimSuffix(in.Method, subscribeMethodSuffix), subscribeMethod[0]
			reqs[0].Params = in.Payload
			return reqs, false, nil
		}
		return nil, false, &ts.InvalidRequestError{Message:"Unable to parse subscription request"}
	}

	if strings.HasSuffix(in.Method, unsubscribeMethodSuffix) {
		return []ts.RpcRequest{{Id: &in.Id, IsPubSub: true,
			Method: in.Method, Params: in.Payload}}, false, nil
	}

	elems := strings.Split(in.Method, serviceMethodSeparator)
	if len(elems) != 2 {
		return nil, false, &ts.MethodNotFoundError{Service:in.Method, Method:""}
	}

	// regular RPC call
	if len(in.Payload) == 0 {
		return []ts.RpcRequest{{Service: elems[0], Method: elems[1], Id: &in.Id}}, false, nil
	}

	return []ts.RpcRequest{{Service: elems[0], Method: elems[1], Id: &in.Id, Params: in.Payload}}, false, nil
}

// parseBatchRequest will parse a batch request into a collection of requests from the given RawMessage, an indication
// if the request was a batch or an error when the request could not be read.
func parseBatchRequest(incomingMsg json.RawMessage) ([]ts.RpcRequest, bool, ts.Error) {
	var in []JsonRequest
	if err := json.Unmarshal(incomingMsg, &in); err != nil {
		return nil, false, &ts.InvalidMessageError{Message:err.Error()}
	}

	requests := make([]ts.RpcRequest, len(in))
	for i, r := range in {
		if err := checkReqId(r.Id); err != nil {
			return nil, false, &ts.InvalidMessageError{Message:err.Error()}
		}

		id := &in[i].Id

		// subscribe are special, they will always use `subscriptionMethod` as first param in the payload
		if strings.HasSuffix(r.Method, subscribeMethodSuffix) {
			requests[i] = ts.RpcRequest{Id: id, IsPubSub: true}
			if len(r.Payload) > 0 {
				// first param must be subscription name
				var subscribeMethod [1]string
				if err := json.Unmarshal(r.Payload, &subscribeMethod); err != nil {
					log.Debug(fmt.Sprintf("Unable to parse subscription method: %v\n", err))
					return nil, false, &ts.InvalidRequestError{Message:"Unable to parse subscription request"}
				}

				requests[i].Service, requests[i].Method = strings.TrimSuffix(r.Method, subscribeMethodSuffix), subscribeMethod[0]
				requests[i].Params = r.Payload
				continue
			}

			return nil, true, &ts.InvalidRequestError{Message:"Unable to parse (un)subscribe request arguments"}
		}

		if strings.HasSuffix(r.Method, unsubscribeMethodSuffix) {
			requests[i] = ts.RpcRequest{Id: id, IsPubSub: true, Method: r.Method, Params: r.Payload}
			continue
		}

		if len(r.Payload) == 0 {
			requests[i] = ts.RpcRequest{Id: id, Params: nil}
		} else {
			requests[i] = ts.RpcRequest{Id: id, Params: r.Payload}
		}
		if elem := strings.Split(r.Method, serviceMethodSeparator); len(elem) == 2 {
			requests[i].Service, requests[i].Method = elem[0], elem[1]
		} else {
			requests[i].Err = &ts.MethodNotFoundError{Service:r.Method, Method:""}
		}
	}

	return requests, true, nil
}

// ParseRequestArguments tries to parse the given params (json.RawMessage) with the given
// types. It returns the parsed values or an error when the parsing failed.
func (c *JsonCodec) ParseRequestArguments(argTypes []reflect.Type, params interface{}) ([]reflect.Value, ts.Error) {
	if args, ok := params.(json.RawMessage); !ok {
		return nil, &ts.InvalidParamsError{Message:"Invalid params supplied"}
	} else {
		return parsePositionalArguments(args, argTypes)
	}
}

// parsePositionalArguments tries to parse the given args to an array of values with the
// given types. It returns the parsed values or an error when the args could not be
// parsed. Missing optional arguments are returned as reflect.Zero values.
func parsePositionalArguments(rawArgs json.RawMessage, types []reflect.Type) ([]reflect.Value, ts.Error) {
	// Read beginning of the args array.
	dec := json.NewDecoder(bytes.NewReader(rawArgs))
	if tok, _ := dec.Token(); tok != json.Delim('[') {
		return nil, &ts.InvalidParamsError{Message:"non-array args"}
	}
	// Read args.
	args := make([]reflect.Value, 0, len(types))
	for i := 0; dec.More(); i++ {
		if i >= len(types) {
			return nil, &ts.InvalidParamsError{Message:fmt.Sprintf("too many arguments, want at most %d", len(types))}
		}
		argval := reflect.New(types[i])
		if err := dec.Decode(argval.Interface()); err != nil {
			return nil, &ts.InvalidParamsError{Message:fmt.Sprintf("invalid argument %d: %v", i, err)}
		}
		if argval.IsNil() && types[i].Kind() != reflect.Ptr {
			return nil, &ts.InvalidParamsError{Message:fmt.Sprintf("missing value for required argument %d", i)}
		}
		args = append(args, argval.Elem())
	}
	// Read end of args array.
	if _, err := dec.Token(); err != nil {
		return nil, &ts.InvalidParamsError{Message:err.Error()}
	}
	// Set any missing args to nil.
	for i := len(args); i < len(types); i++ {
		if types[i].Kind() != reflect.Ptr {
			return nil, &ts.InvalidParamsError{Message:fmt.Sprintf("missing value for required argument %d", i)}
		}
		args = append(args, reflect.Zero(types[i]))
	}
	return args, nil
}

// CreateResponse will create a JSON-RPC success response with the given id and reply as result.
func (c *JsonCodec) CreateResponse(id interface{}, reply interface{}) interface{} {
	return &JsonSuccessResponse{Version: jsonrpcVersion, Id: id, Result: reply}
}

// CreateErrorResponse will create a JSON-RPC error response with the given id and error.
func (c *JsonCodec) CreateErrorResponse(id interface{}, err ts.Error) interface{} {
	return &JsonErrResponse{Version: jsonrpcVersion, Id: id, Error: JsonError{Code: err.ErrorCode(), Message: err.Error()}}
}

// CreateErrorResponseWithInfo will create a JSON-RPC error response with the given id and error.
// info is optional and contains additional information about the error. When an empty string is passed it is ignored.
func (c *JsonCodec) CreateErrorResponseWithInfo(id interface{}, err ts.Error, info interface{}) interface{} {
	return &JsonErrResponse{Version: jsonrpcVersion, Id: id,
		Error: JsonError{Code: err.ErrorCode(), Message: err.Error(), Data: info}}
}

// CreateNotification will create a JSON-RPC notification with the given subscription id and event as params.
func (c *JsonCodec) CreateNotification(subid, namespace string, event interface{}) interface{} {
	return &JsonNotification{Version: jsonrpcVersion, Method: namespace + notificationMethodSuffix,
		Params: JsonSubscription{Subscription: subid, Result: event}}
}

// Write message to client
func (c *JsonCodec) Write(res interface{}) error {
	c.encMu.Lock()
	defer c.encMu.Unlock()

	return c.encode(res)
}

// Close the underlying connection
func (c *JsonCodec) Close() {
	c.closer.Do(func() {
		close(c.closed)
		//c.rw.Close()
	})
}

// Closed returns a channel which will be closed when Close is called
func (c *JsonCodec) Closed() <-chan interface{} {
	return c.closed
}
