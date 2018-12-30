// Copyright 2016 The go-ethereum Authors
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

package event

import (
	"sync"
)

// SubscriptionScope provides a facility to unsubscribe multiple subscriptions at once.
//
// For code that handle more than one subscription, a scope can be used to conveniently
// unsubscribe all of them with a single call. The example demonstrates a typical use in a
// larger program.
//
// The zero value is ready to use.
type SubscriptionScope struct {
	mu     sync.Mutex
	subs   map[*scopeSub]struct{}
	closed bool
}

type scopeSub struct {
	sc *SubscriptionScope
	s  Subscription
}

// Track starts tracking a subscription. If the scope is closed, Track returns nil. The
// returned subscription is a wrapper. Unsubscribing the wrapper removes it from the
// scope.
func (sc *SubscriptionScope) Track(s Subscription) Subscription {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if sc.closed {
		return nil
	}
	if sc.subs == nil {
		sc.subs = make(map[*scopeSub]struct{})
	}
	ss := &scopeSub{sc, s}
	sc.subs[ss] = struct{}{}
	return ss
}

// Close calls Unsubscribe on all tracked subscriptions and prevents further additions to
// the tracked set. Calls to Track after Close return nil.
func (sc *SubscriptionScope) Close() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if sc.closed {
		return
	}
	sc.closed = true
	for s := range sc.subs {
		s.s.Unsubscribe()
	}
	sc.subs = nil
}

// Count returns the number of tracked subscriptions.
// It is meant to be used for debugging.
func (sc *SubscriptionScope) Count() int {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return len(sc.subs)
}

func (s *scopeSub) Unsubscribe() {
	s.s.Unsubscribe()
	s.sc.mu.Lock()
	defer s.sc.mu.Unlock()
	delete(s.sc.subs, s)
}

func (s *scopeSub) Err() <-chan error {
	return s.s.Err()
}
