package server

import (
	"context"
	"reflect"
	"sync"
	"unicode"
	"unicode/utf8"
)

var (
	subscriptionIDGenMu sync.Mutex
	subscriptionIDGen   = idGenerator()
)

// Is this an exported - upper case - name?
func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

// Is this type exported or a builtin?
func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return isExported(t.Name()) || t.PkgPath() == ""
}

var contextType = reflect.TypeOf((*context.Context)(nil)).Elem()

// isContextType returns an indication if the given t is of context.Context or *context.Context type
func isContextType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t == contextType
}

var errorType = reflect.TypeOf((*error)(nil)).Elem()

// Implements this type the error interface
func isErrorType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Implements(errorType)
}

var subscriptionType = reflect.TypeOf((*Subscription)(nil)).Elem()

// isSubscriptionType returns an indication if the given t is of Subscription or *Subscription type
func isSubscriptionType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t == subscriptionType
}

// isPubSub tests whether the given method has as as first argument a context.Context
// and returns the pair (Subscription, error)
func isPubSub(methodType reflect.Type) bool {
	// numIn(0) is the receiver type
	if methodType.NumIn() < 2 || methodType.NumOut() != 2 {
		return false
	}

	return isContextType(methodType.In(1)) &&
		isSubscriptionType(methodType.Out(0)) &&
		isErrorType(methodType.Out(1))
}

// formatName will convert to first character to lower case
func formatName(name string) string {
	ret := []rune(name)
	if len(ret) > 0 {
		ret[0] = unicode.ToLower(ret[0])
	}
	return string(ret)
}

// suitableCallbacks iterates over the methods of the given type. It will determine if a method satisfies the criteria
// for a RPC callback or a subscription callback and adds it to the collection of callbacks or subscriptions. See server
// documentation for a summary of these criteria.
func suitableCallbacks(rcvr reflect.Value, typ reflect.Type) (Callbacks, Subscriptions) {
	callbacks := make(Callbacks)
	subscriptions := make(Subscriptions)

METHODS:
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := formatName(method.Name)
		if method.PkgPath != "" { // method must be exported
			continue
		}

		var h Callback
		h.IsSubscribe = isPubSub(mtype)
		h.Rcvr = rcvr
		h.Method = method
		h.ErrPos = -1

		firstArg := 1
		numIn := mtype.NumIn()
		if numIn >= 2 && mtype.In(1) == contextType {
			h.HasCtx = true
			firstArg = 2
		}

		if h.IsSubscribe {
			h.ArgTypes = make([]reflect.Type, numIn-firstArg) // skip rcvr type
			for i := firstArg; i < numIn; i++ {
				argType := mtype.In(i)
				if isExportedOrBuiltinType(argType) {
					h.ArgTypes[i-firstArg] = argType
				} else {
					continue METHODS
				}
			}

			subscriptions[mname] = &h
			continue METHODS
		}

		// determine method arguments, ignore first arg since it's the receiver type
		// Arguments must be exported or builtin types
		h.ArgTypes = make([]reflect.Type, numIn-firstArg)
		for i := firstArg; i < numIn; i++ {
			argType := mtype.In(i)
			if !isExportedOrBuiltinType(argType) {
				continue METHODS
			}
			h.ArgTypes[i-firstArg] = argType
		}

		// check that all returned values are exported or builtin types
		for i := 0; i < mtype.NumOut(); i++ {
			if !isExportedOrBuiltinType(mtype.Out(i)) {
				continue METHODS
			}
		}

		// when a method returns an error it must be the last returned value
		h.ErrPos = -1
		for i := 0; i < mtype.NumOut(); i++ {
			if isErrorType(mtype.Out(i)) {
				h.ErrPos = i
				break
			}
		}

		if h.ErrPos >= 0 && h.ErrPos != mtype.NumOut()-1 {
			continue METHODS
		}

		switch mtype.NumOut() {
		case 0, 1, 2:
			if mtype.NumOut() == 2 && h.ErrPos == -1 { // method must one return value and 1 error
				continue METHODS
			}
			callbacks[mname] = &h
		}
	}

	return callbacks, subscriptions
}
