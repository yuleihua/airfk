// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
package types

import "fmt"

// Error wraps RPC errors, which contain an error code in addition to the message.
type Error interface {
	Error() string  // returns the message
	ErrorCode() int // returns the code
}

// request is for an unknown service
type MethodNotFoundError struct {
	Service string
	Method  string
}

func (e *MethodNotFoundError) ErrorCode() int { return -32601 }

func (e *MethodNotFoundError) Error() string {
	return fmt.Sprintf("The method %s%s%s does not exist/is not available", e.Service, "_", e.Method)
}

// received Message isn't a valid request
type InvalidRequestError struct{ Message string }

func (e *InvalidRequestError) ErrorCode() int { return -32600 }

func (e *InvalidRequestError) Error() string { return e.Message }

// received Message is invalid
type InvalidMessageError struct{ Message string }

func (e *InvalidMessageError) ErrorCode() int { return -32700 }

func (e *InvalidMessageError) Error() string { return e.Message }

// unable to decode supplied params, or an invalid number of parameters
type InvalidParamsError struct{ Message string }

func (e *InvalidParamsError) ErrorCode() int { return -32602 }

func (e *InvalidParamsError) Error() string { return e.Message }

// logic error, callback returned an error
type CallbackError struct{ Message string }

func (e *CallbackError) ErrorCode() int { return -32000 }

func (e *CallbackError) Error() string { return e.Message }

// issued when a request is received after the server is issued to stop.
type ShutdownError struct{}

func (e *ShutdownError) ErrorCode() int { return -32000 }

func (e *ShutdownError) Error() string { return "server is shutting down" }
