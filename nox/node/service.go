package node

import (
	"reflect"
	"fmt"
	"github.com/noxproject/nox/rpc"
	"github.com/noxproject/nox/core/protocol"
	"github.com/noxproject/nox/p2p/peerserver"
)

const (
	// the default services supported by the node
	defaultServices = protocol.Full| protocol.CF

	// the default services that are required to be supported
	defaultRequiredServices = protocol.Full & protocol.Light
)

// Service is a service can be registered into & running in a Node
type Service interface {

	// APIs retrieves the list of RPC descriptors the service provides
	APIs() []rpc.API

	// Start is called after all services have been constructed and the networking
	// layer was also initialized to spawn any goroutines required by the service.
	Start(server *peerserver.PeerServer) error

	// Stop terminates all goroutines belonging to the service, blocking until they
	// are all terminated.
	Stop() error
}

// ServiceStopError is returned if a Node fails to stop either any of its registered
// services or itself.
type ServiceStopError struct {
	Services map[reflect.Type]error
}

// Error generates a textual representation of the stop error.
func (e *ServiceStopError) Error() string {
	return fmt.Sprintf("services: %v", e.Services)
}

// ServiceContext is a collection of service independent options inherited from
// the protocol stack, that is passed to all constructors to be optionally used;
// as well as utility methods to operate on the service environment.
type ServiceContext struct {

}

// ServiceConstructor is the function signature of the constructors needed to be
// registered for service instantiation.
type ServiceConstructor struct {
	service string
	initFunc func(ctx *ServiceContext) (Service, error)
}

func NewServiceConstructor(name string, constructor func(ctx *ServiceContext) (Service, error)) ServiceConstructor{
	sc := ServiceConstructor{
		initFunc:constructor,
		service:name,
	}
	return sc
}
