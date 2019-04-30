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
package subscribe

import (
	"context"
	"fmt"
	"testing"
	"time"

	"{{ .LibDir }}/pkg/event"

	cmn "{{ .RelDir }}/node/common"
)

const msgCycle = 200 * time.Millisecond

type TestBackend struct {
	ctx context.Context

	resultFeed  event.Feed              // Event feed to notify wallet additions/removals
	resultScope event.SubscriptionScope // Subscription scope tracking current live listeners
}

func (t *TestBackend) SubscribeResultEvent(ch chan<- []cmn.Result) event.Subscription {

	// Subscribe the caller and track the subscriber count
	sub := t.resultScope.Track(t.resultFeed.Subscribe(ch))

	// Subscribers require an active notification loop, start it
	go t.updater()

	return sub
}

func (t *TestBackend) updater() {
	for {
		// Wait for an account update or a refresh timeout
		select {
		case <-time.After(msgCycle):
		case <-t.ctx.Done():
			return
		}

		r := cmn.NewResultWithEnd(time.Now().String(), time.Now().Unix(), time.Now().Unix(), "ok", []byte("output is 2046"))
		t.resultFeed.Send([]cmn.Result{*r})
	}
}

func TestEventMsg(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	backend := &TestBackend{ctx: ctx}

	m, err := NewEventMsg(ctx, backend)
	if m == nil || err != nil {
		t.Fatal("no notify")
	}

	go func(ctx context.Context) {
		uuid := make(chan []cmn.Result)
		uuidSub := m.SubscribeResultTask(uuid)

		for {
			select {
			case r := <-uuid:
				fmt.Printf("len(r):%d, %#v\n", len(r), r)
			case <-ctx.Done():
				uuidSub.Unsubscribe()
				return
			}
		}
	}(ctx)

	time.Sleep(11 * time.Second)
}
