package node

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/rpc"
	"reflect"
)

// Service is a service can be registered into & running in a Node
type Service interface {

	// APIs retrieves the list of RPC descriptors the service provides
	APIs() []rpc.API

	// Start is called after all services have been constructed and the networking
	// layer was also initialized to spawn any goroutines required by the service.
	Start() error

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
	service  string
	initFunc func(ctx *ServiceContext) (Service, error)
}

func NewServiceConstructor(name string, constructor func(ctx *ServiceContext) (Service, error)) ServiceConstructor {
	sc := ServiceConstructor{
		initFunc: constructor,
		service:  name,
	}
	return sc
}
