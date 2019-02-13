// Copyright 2018 The huayulei_2003@hotmail.com Authors
// This file is part of the airfk library.
//
// The airfk library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The airfk library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the airfk library. If not, see <http://www.gnu.org/licenses/>.
package context

import (
	"context"
	"time"
)

type Context struct {
	context context.Context
	cancel  context.CancelFunc
}

// NewContext returns a non-nil, empty Context.
func NewContext() *Context {
	return &Context{
		context: context.Background(),
	}
}

// WithCancel returns a copy of the original context with cancellation mechanism
// included.
func (c *Context) WithCancel() *Context {
	child, cancel := context.WithCancel(c.context)
	return &Context{
		context: child,
		cancel:  cancel,
	}
}

// WithTimeout returns a copy of the original context with the deadline adjusted
// to be no later than now + the duration specified.
func (c *Context) WithTimeout(timeout time.Duration) *Context {
	child, cancel := context.WithTimeout(c.context, timeout)
	return &Context{
		context: child,
		cancel:  cancel,
	}
}

type (
	logIDKey struct{}
)

// SetID
func SetLogID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, logIDKey{}, id)
}

// GetID
func GetLogID(ctx context.Context) string {
	v, ok := ctx.Value(logIDKey{}).(string)
	if ok {
		return v
	}
	return ""
}
