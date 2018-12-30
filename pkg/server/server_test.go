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

package server

import (
	"context"
	"encoding/json"
	"net"
	"reflect"
	"testing"
	"time"

	"airman.com/airfk/pkg/codec"
)

type DemoServer struct{}

type Args struct {
	S string
}

func (s *DemoServer) NoArgsRets() {
}

type Result struct {
	String string
	Int    int
	Args   *Args
}

func (s *DemoServer) Echo(str string, i int, args *Args) Result {
	return Result{str, i, args}
}

func (s *DemoServer) EchoWithCtx(ctx context.Context, str string, i int, args *Args) Result {
	return Result{str, i, args}
}

func (s *DemoServer) Sleep(ctx context.Context, duration time.Duration) {
	select {
	case <-time.After(duration):
	case <-ctx.Done():
	}
}

func (s *DemoServer) Rets() (string, error) {
	return "", nil
}

func (s *DemoServer) InvalidRets1() (error, string) {
	return nil, ""
}

func (s *DemoServer) InvalidRets2() (string, string) {
	return "", ""
}

func (s *DemoServer) InvalidRets3() (string, string, error) {
	return "", "", nil
}

func (s *DemoServer) Subscription(ctx context.Context) (*Subscription, error) {
	return nil, nil
}

func TestServerRegisterName(t *testing.T) {
	server := NewServer()
	DemoServer := new(DemoServer)

	if err := server.RegisterName("calc", DemoServer); err != nil {
		t.Fatalf("%v", err)
	}

	if len(server.Services) != 2 {
		t.Fatalf("Expected 2 DemoServer entries, got %d", len(server.Services))
	}

	svc, ok := server.Services["calc"]
	if !ok {
		t.Fatalf("Expected DemoServer calc to be registered")
	}

	if len(svc.Callbacks) != 5 {
		t.Errorf("Expected 5 callbacks for DemoServer 'calc', got %d", len(svc.Callbacks))
	}

	if len(svc.Subscriptions) != 1 {
		t.Errorf("Expected 1 subscription for DemoServer 'calc', got %d", len(svc.Subscriptions))
	}
}

func testServerMethodExecution(t *testing.T, method string) {
	server := NewServer()
	DemoServer := new(DemoServer)

	if err := server.RegisterName("test", DemoServer); err != nil {
		t.Fatalf("%v", err)
	}

	stringArg := "string arg"
	intArg := 1122
	argsArg := &Args{"abcde"}
	params := []interface{}{stringArg, intArg, argsArg}

	request := map[string]interface{}{
		"id":      12345,
		"method":  "test_" + method,
		"version": "2.0",
		"params":  params,
	}

	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()

	go server.ServeCodec(codec.NewJSONCodec(serverConn), OptionMethodInvocation)

	out := json.NewEncoder(clientConn)
	in := json.NewDecoder(clientConn)

	if err := out.Encode(request); err != nil {
		t.Fatal(err)
	}

	response := codec.JsonSuccessResponse{Result: &Result{}}
	if err := in.Decode(&response); err != nil {
		t.Fatal(err)
	}

	if result, ok := response.Result.(*Result); ok {
		if result.String != stringArg {
			t.Errorf("expected %s, got : %s\n", stringArg, result.String)
		}
		if result.Int != intArg {
			t.Errorf("expected %d, got %d\n", intArg, result.Int)
		}
		if !reflect.DeepEqual(result.Args, argsArg) {
			t.Errorf("expected %v, got %v\n", argsArg, result)
		}
	} else {
		t.Fatalf("invalid response: expected *Result - got: %T", response.Result)
	}
}

func TestServerMethodExecution(t *testing.T) {
	testServerMethodExecution(t, "echo")
}

func TestServerMethodWithCtx(t *testing.T) {
	testServerMethodExecution(t, "echoWithCtx")
}
