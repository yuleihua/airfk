package codec

import (
	"reflect"

	ts "airman.com/airfk/pkg/types"
)

// ServerCodec implements reading, parsing and writing RPC messages for the server side of
// a RPC session. Implementations must be go-routine safe since the codec can be called in
// multiple go-routines concurrently.
type ServerCodec interface {
	// Read next request
	ReadRequestHeaders() ([]ts.RpcRequest, bool, ts.Error)
	// Parse request argument to the given types
	ParseRequestArguments(argTypes []reflect.Type, params interface{}) ([]reflect.Value, ts.Error)
	// Assemble success response, expects response id and payload
	CreateResponse(id interface{}, reply interface{}) interface{}
	// Assemble error response, expects response id and error
	CreateErrorResponse(id interface{}, err ts.Error) interface{}
	// Assemble error response with extra information about the error through info
	CreateErrorResponseWithInfo(id interface{}, err ts.Error, info interface{}) interface{}
	// Create notification response
	CreateNotification(id, namespace string, event interface{}) interface{}
	// Write msg to client.
	Write(msg interface{}) error
	// Close underlying data stream
	Close()
	// Closed when underlying connection is closed
	Closed() <-chan interface{}
}
