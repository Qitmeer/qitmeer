// Copyright (c) 2017-2018 The nox developers

package rpc

import (

	"sync"
	"reflect"
	"fmt"
	"net/http"
	"net"
	"time"
	"sync/atomic"
	"github.com/noxproject/nox/config"
	"github.com/noxproject/nox/common/util"
	"github.com/noxproject/nox/log"
	"crypto/sha256"
	"crypto/subtle"
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

	authsha                [sha256.Size]byte
	limitauthsha           [sha256.Size]byte
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
	//TODO control by config
	if err := s.startHTTP(s.config.Listeners); err!=nil {
		return err
	}
	return nil
}

const (
	// rpcAuthTimeoutSeconds is the number of seconds a connection to the
	// RPC server is allowed to stay open without authenticating before it
	// is closed.
	rpcAuthTimeoutSeconds = 10
)

func (s *RpcServer) startHTTP(listenAddrs []string) error{

	rpcServeMux := http.NewServeMux()
	httpServer := &http.Server{
		Handler: rpcServeMux,

		// Timeout connections which don't complete the initial
		// handshake within the allowed timeframe.
		ReadTimeout: time.Second * rpcAuthTimeoutSeconds,
	}
	rpcServeMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Connection", "close")
		w.Header().Set("Content-Type", "application/json")
		r.Close = true

		// Limit the number of connections to max allowed.
		if s.limitConnections(w, r.RemoteAddr) {
			return
		}

		// Keep track of the number of connected clients.
		s.incrementClients()
		defer s.decrementClients()
		_, isAdmin, err := s.checkAuth(r, true)
		if err != nil {
			jsonAuthFail(w)
			return
		}
		// Read and respond to the request.
		s.jsonRPCRead(w, r, isAdmin)
	})
	listeners, err := ParseListeners(s.config,listenAddrs);
	if err!=nil {
		return err
	}
	for _, listener := range listeners {
		s.wg.Add(1)
		go func(listener net.Listener) {
			log.Info("RPC server listening on ", "addr", listener.Addr())
			httpServer.Serve(listener)
			log.Trace("RPC listener done for %s", listener.Addr())
			s.wg.Done()
		}(listener)
	}
	return nil
}

// limitConnections responds with a 503 service unavailable and returns true if
// adding another client would exceed the maximum allow RPC clients.
//
// This function is safe for concurrent access.
func (s *RpcServer) limitConnections(w http.ResponseWriter, remoteAddr string) bool {
	if int(atomic.LoadInt32(&s.numClients)+1) > s.config.RPCMaxClients {
		log.Info("Max RPC clients exceeded [%d] - "+
			"disconnecting client %s", s.config.RPCMaxClients,
			remoteAddr)
		http.Error(w, "503 Too busy.  Try again later.",
			http.StatusServiceUnavailable)
		return true
	}
	return false
}

// incrementClients adds one to the number of connected RPC clients.  Note this
// only applies to standard clients.  Websocket clients have their own limits
// and are tracked separately.
//
// This function is safe for concurrent access.
func (s *RpcServer) incrementClients() {
	atomic.AddInt32(&s.numClients, 1)
}

// decrementClients subtracts one from the number of connected RPC clients.
// Note this only applies to standard clients.  Websocket clients have their
// own limits and are tracked separately.
//
// This function is safe for concurrent access.
func (s *RpcServer) decrementClients() {
	atomic.AddInt32(&s.numClients, -1)
}

// checkAuth checks the HTTP Basic authentication supplied by a wallet or RPC
// client in the HTTP request r.  If the supplied authentication does not match
// the username and password expected, a non-nil error is returned.
//
// This check is time-constant.
//
// The first bool return value signifies auth success (true if successful) and
// the second bool return value specifies whether the user can change the state
// of the server (true) or whether the user is limited (false). The second is
// always false if the first is.
func (s *RpcServer) checkAuth(r *http.Request, require bool) (bool, bool, error) {
	authhdr := r.Header["Authorization"]
	if len(authhdr) <= 0 {
		if require {
			log.Warn("RPC authentication failure from %s", r.RemoteAddr)
			return false, false, fmt.Errorf("auth failure")
		}

		return false, false, nil
	}

	authsha := sha256.Sum256([]byte(authhdr[0]))

	// Check for limited auth first as in environments with limited users,
	// those are probably expected to have a higher volume of calls
	limitcmp := subtle.ConstantTimeCompare(authsha[:], s.limitauthsha[:])
	if limitcmp == 1 {
		return true, false, nil
	}

	// Check for admin-level auth
	cmp := subtle.ConstantTimeCompare(authsha[:], s.authsha[:])
	if cmp == 1 {
		return true, true, nil
	}

	// Request's auth doesn't match either user
	log.Warn("RPC authentication failure from %s", r.RemoteAddr)
	return false, false, fmt.Errorf("auth failure")
}

// jsonAuthFail sends a message back to the client if the http auth is rejected.
func jsonAuthFail(w http.ResponseWriter) {
	w.Header().Add("WWW-Authenticate", `Basic realm="nox RPC"`)
	http.Error(w, "401 Unauthorized.", http.StatusUnauthorized)
}

// jsonRPCRead handles reading and responding to RPC messages.
func (s *RpcServer) jsonRPCRead(w http.ResponseWriter, r *http.Request, isAdmin bool) {
	if atomic.LoadInt32(&s.shutdown) != 0 {
		return
	}
	log.Debug("jsonRPC ", "request",r)
}

// RegisterService will create a service for the given type under the given namespace.
// When no methods on the given type match the criteria to be either a RPC method or
// a subscription an error is returned. Otherwise a new service is created and added
// to the service registry.
func (s *RpcServer) RegisterService(namespace string, regSvc interface{}) error {

	typ := reflect.TypeOf(regSvc)
	if namespace == "" {
		return fmt.Errorf("no service namespace for type %s", typ.String())
	}

	// parse & build callbacks/subscriptions
	value := reflect.ValueOf(regSvc)
	calls, subs := suitableCallbacks(value, typ)

	// if the namespace already registered, add callback/subscriptions & return
	if foundSrv, nsExist := s.rpcSvcRegistry[namespace]; nsExist {
		if len(calls) == 0 && len(subs) == 0 {
			return fmt.Errorf("Service %T doesn't have any suitable methods/subscriptions to expose", regSvc)
		}
		for _, m := range calls {
			foundSrv.callbacks[formatName(m.method.Name)] = m
		}
		for _, s := range subs {
			foundSrv.subscriptions[formatName(s.method.Name)] = s
		}
		return nil
	}

	// create new service with callbacks/subscriptions & add to registry
	svc := service{
		svcNamespace:namespace,
		svcType:typ,
		callbacks:calls,
		subscriptions:subs,
	}
	s.rpcSvcRegistry[svc.svcNamespace] = &svc
	return nil
}


