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
package {{.Name}}

import (
	"context"
	"sync"

	"{{ .LibDir }}/pkg/event"
	"{{ .LibDir }}/pkg/types"

	cmn "{{ .RelDir }}/node/common"
	sub "{{ .RelDir }}/node/subscribe"
)

// Manager workers.
type Manager struct {
	backend     Backend
	es          *sub.EventMsg
	resultsFeed event.Feed // Wallet feed notifying of arrivals/departures
	scope       event.SubscriptionScope

	cancel context.CancelFunc
	mu     sync.RWMutex
}

// NewManager create manager object.
func NewManager(backend Backend) *Manager {
	return &Manager{backend: backend}
}

// Start manager object.
func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	feed, err := sub.NewEventMsg(ctx, m)
	if err != nil {
		return err
	}
	m.cancel = cancel
	m.es = feed
	return nil
}

// Stop manager object.
func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cancel()
	return nil
}

// APIs returns the collection of RPC descriptors this node offers.
func (m *Manager) APIs() []types.API {
	return []types.API{
		{
			Namespace: "{{.Name}}",
			Version:   "1.0",
			Service:   NewPrivate{{.CamelCaseName}}API(m),
		}, {
			Namespace: "{{.Name}}",
			Version:   "1.0",
			Service:   NewPublic{{.CamelCaseName}}API(m),
			Public:    true,
		},
	}
}

// SubscribeResultEvent registers a subscription of task results.
func (m *Manager) SubscribeResultEvent(ch chan<- []cmn.Result) event.Subscription {
	return m.scope.Track(m.resultsFeed.Subscribe(ch))
}
