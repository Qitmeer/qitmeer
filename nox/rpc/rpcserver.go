// Copyright (c) 2017-2018 The nox developers

package rpc

import (
	"github.com/noxproject/nox/config"
	"github.com/noxproject/nox/common/util"
	"sync"
	"reflect"
	"fmt"
)

// API describes the set of methods offered over the RPC interface
type API struct {
	NameSpace string      // namespace under which the rpc methods of Service are exposed
	Service   interface{} // receiver instance which holds the methods
}

// RpcServer provides a concurrent safe RPC server to a chain server.
type RpcServer struct {

	started                int32
	shutdown               int32
	wg                     util.WaitGroupWrapper
	quit                   chan int
	statusLock             sync.RWMutex

	config                  *config.Config

	rpcSvcRegistry         serviceRegistry

	numClients             int32
	statusLines            map[int]string
	requestProcessShutdown chan struct{}

}

// service represents a registered object
type service struct {
	svcNamespace     string        //the name space for service
	svcType          reflect.Type  // receiver type
	callbacks        callbacks     // registered service method
	subscriptions    subscriptions // available subscriptions/notifications
}

// callback is a method callback which was registered in the server
type callback struct {
	receiver    reflect.Value  // receiver of method
	method      reflect.Method // callback
	argTypes    []reflect.Type // input argument types
	hasCtx      bool           // method's first argument is a context (not included in argTypes)
	errPos      int            // err return idx, of -1 when method cannot return error
	isSubscribe bool           // indication if the callback is a subscription
}

// serviceRegistry is the collection of services by namespace
type serviceRegistry map[string]*service
// callbacks is the collection of RPC callbacks
type callbacks map[string]*callback
// subscriptions is the collection of subscription callbacks
type subscriptions map[string]*callback

// rpcRequest represents a raw incoming RPC request
type rpcRequest struct {
	service  string
	method   string
	id       interface{}
	isPubSub bool
	params   interface{}
	err      Error // invalid batch element
}


// newRPCServer returns a new instance of the rpcServer struct.
func NewRPCServer(cfg *config.Config) (*RpcServer, error) {
	rpc := RpcServer{

		config:                 cfg,

		rpcSvcRegistry:         make(serviceRegistry),

		statusLines:            make(map[int]string),
		requestProcessShutdown: make(chan struct{}),
		quit: make(chan int),
	}
	return &rpc, nil
}

func (s *RpcServer) Start() error {
	return nil
}

func (s *RpcServer) RegisterService(namespace string, regSvc interface{}) error {
	typ := reflect.TypeOf(regSvc)
	if namespace == "" {
		return fmt.Errorf("no service namespace for type %s", typ.String())
	}

	svc := service{
		svcNamespace:namespace,
		svcType:typ,
	}

	s.rpcSvcRegistry[svc.svcNamespace] = &svc
	return nil
}


