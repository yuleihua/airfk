package common

import "airman.com/airfk/pkg/types"

// Service is an individual protocol that can be registered into a server.
type Service interface {
	// APIs retrieves the list of RPC descriptors the service provides
	APIs() []types.API

	// Start is called after all services have been constructed and the networking
	// layer was also initialized to spawn any goroutines required by the service.
	Start() error

	// Stop terminates all goroutines belonging to the service, blocking until they
	// are all terminated.
	Stop() error
}
