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
package admin

import (
	"fmt"
	"strings"
	"sync"

	"airman.com/airfk/node/conf"
)

// PrivateAdminAPI is the collection of administrative API methods exposed only
// over a secure RPC channel.
type PrivateAdminAPI struct {
	node Backend // Node interfaced by this API

	mu sync.Mutex
}

// NewPrivateAdminAPI creates a new API definition for the private admin methods
// of the node itself.
func NewPrivateAdminAPI(node Backend) *PrivateAdminAPI {
	return &PrivateAdminAPI{node: node}
}

// StartWS starts the websocket RPC API server.
func (api *PrivateAdminAPI) StartWS(host *string, port *int, allowedOrigins *string, apis *string) (bool, error) {
	api.mu.Lock()
	defer api.mu.Unlock()

	if api.node.WSHandle() != nil {
		return false, fmt.Errorf("WebSocket RPC already running on %s", api.node.WSEndpoint())
	}

	config := conf.DefaultConfig

	if host != nil {
		config.WSHost = *host
	}
	if port != nil {
		config.WSPort = *port
	}

	if allowedOrigins != nil {
		var origins []string
		for _, origin := range strings.Split(*allowedOrigins, ",") {
			origins = append(origins, strings.TrimSpace(origin))
		}
		config.WSOrigins = origins
	}

	if apis != nil {
		var modules []string
		for _, m := range strings.Split(*apis, ",") {
			modules = append(modules, strings.TrimSpace(m))
		}
		config.WSModules = modules
	}

	if err := api.node.StartWS(config, api.node.RpcAPIs()); err != nil {
		return false, err
	}
	return true, nil
}

// StopWS terminates an already running websocket RPC API endpoint.
func (api *PrivateAdminAPI) StopWS() (bool, error) {
	api.mu.Lock()
	defer api.mu.Unlock()

	if api.node.WSHandle() == nil {
		return false, fmt.Errorf("WebSocket RPC not running")
	}
	api.node.StopWS()
	return true, nil
}

// PublicAdminAPI is the collection of administrative API methods exposed over
// both secure and unsecure RPC channels.
type PublicAdminAPI struct {
	n Backend // Node interfaced by this API
}

// NewPublicAdminAPI creates a new API definition for the public admin methods
// of the node itself.
func NewPublicAdminAPI(node Backend) *PublicAdminAPI {
	return &PublicAdminAPI{n: node}
}

// NodeInfo retrieves all the information we know about the host node at the
// protocol granularity.
func (api *PublicAdminAPI) NodeInfo() (map[string]interface{}, error) {
	return map[string]interface{}{
		"config":    api.n.Config(),
		"apis":      api.n.RpcAPIs(),
		"services":  api.n.Services(),
		"isRunning": api.n.IsRunning(),
	}, nil
}

// Datadir retrieves the current data directory the node is using.
func (api *PublicAdminAPI) DataDir() string {
	return api.n.DataDir()
}

// Datadir retrieves the current data directory the node is using.
func (api *PublicAdminAPI) Version() string {
	return fmt.Sprintf("%s %s", api.n.Name(), api.n.Version())
}

// Ping retrieves the current data directory the node is using.
func (api *PublicAdminAPI) Ping() string {
	return "pong"
}
