// Copyright (c) 2017-2018 The qitmeer developers

package rpc

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"github.com/Qitmeer/qitmeer/cmd/crawler/config"
	"github.com/Qitmeer/qitmeer/common/util"
	"github.com/Qitmeer/qitmeer/log"
	"github.com/deckarep/golang-set"
	"golang.org/x/net/context"
	"io"
	"net"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// API describes the set of methods offered over the RPC interface
type API struct {
	NameSpace string      // namespace under which the rpc methods of Service are exposed
	Service   interface{} // receiver instance which holds the methods
	Public    bool        // indication if the methods must be considered safe for public use
}

// RpcServer provides a concurrent safe RPC server to a chain server.
type RpcServer struct {
	run        int32
	wg         util.WaitGroupWrapper
	quit       chan int
	statusLock sync.RWMutex

	config *config.Config

	rpcSvcRegistry serviceRegistry

	codecsMu sync.Mutex
	codecs   mapset.Set

	authsha                [sha256.Size]byte
	numClients             int32
	statusLines            map[int]string
	requestProcessShutdown chan struct{}

	ReqStatus     map[string]*RequestStatus
	reqStatusLock sync.RWMutex
}

// service represents a registered object
type service struct {
	svcNamespace  string        //the name space for service
	svcType       reflect.Type  // receiver type
	callbacks     callbacks     // registered service method
	subscriptions subscriptions // available subscriptions/notifications
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

// serverRequest is an incoming request
type serverRequest struct {
	id            interface{}
	svcname       string
	callb         *callback
	args          []reflect.Value
	isUnsubscribe bool
	err           Error
	time          time.Time
}

// newRPCServer returns a new instance of the rpcServer struct.
func NewRPCServer(cfg *config.Config) (*RpcServer, error) {
	rpc := RpcServer{

		config: cfg,

		rpcSvcRegistry: make(serviceRegistry),
		codecs:         mapset.NewSet(),

		statusLines:            make(map[int]string),
		requestProcessShutdown: make(chan struct{}),
		quit:                   make(chan int),
		ReqStatus:              map[string]*RequestStatus{},
	}

	if cfg.RPCUser != "" && cfg.RPCPass != "" {
		login := cfg.RPCUser + ":" + cfg.RPCPass
		auth := "Basic " +
			base64.StdEncoding.EncodeToString([]byte(login))
		rpc.authsha = sha256.Sum256([]byte(auth))
	}
	return &rpc, nil
}

func (s *RpcServer) Start() error {
	//TODO control by config
	if err := s.startHTTP(s.config.RPCListeners); err != nil {
		return err
	}
	s.run = 1
	return nil
}

// Stop will stop reading new requests, wait for stopPendingRequestTimeout to allow pending requests to finish,
// close all codecs which will cancel pending requests/subscriptions.
func (s *RpcServer) Stop() {
	if atomic.CompareAndSwapInt32(&s.run, 1, 0) {
		log.Debug("RPC Server is stopping")
		s.codecsMu.Lock()
		defer s.codecsMu.Unlock()
		s.codecs.Each(func(c interface{}) bool {
			c.(ServerCodec).Close()
			return true
		})
	}
}

const (
	// rpcAuthTimeoutSeconds is the number of seconds a connection to the
	// RPC server is allowed to stay open without authenticating before it
	// is closed.
	rpcAuthTimeoutSeconds = 10
)

func (s *RpcServer) startHTTP(listenAddrs []string) error {

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
		_, err := s.checkAuth(r, true)
		if err != nil {
			jsonAuthFail(w)
			return
		}
		// Read and respond to the request.
		s.jsonRPCRead(w, r)
	})
	listeners, err := parseListeners(s.config, listenAddrs)
	if err != nil {
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
		log.Info("RPC clients exceeded", "max", s.config.RPCMaxClients,
			"client", remoteAddr)
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

// TODO, repalace Basic Authentication
// checkAuth checks the HTTP Basic authentication supplied by a wallet or RPC
// client in the HTTP request r.  If the supplied authentication does not match
// the username and password expected, a non-nil error is returned.
//
// This check is time-constant.
func (s *RpcServer) checkAuth(r *http.Request, require bool) (bool, error) {
	authhdr := r.Header["Authorization"]
	if len(authhdr) <= 0 {
		if require {
			log.Warn("RPC authentication failure", "from", r.RemoteAddr,
				"error", "no authorization header")
			return false, fmt.Errorf("auth failure")
		}

		return false, nil
	}

	authsha := sha256.Sum256([]byte(authhdr[0]))

	// Check for auth
	cmp := subtle.ConstantTimeCompare(authsha[:], s.authsha[:])
	if cmp == 1 {
		return true, nil
	}

	// Request's auth doesn't match either user
	log.Warn("RPC authentication failure", "from", r.RemoteAddr)
	return false, fmt.Errorf("auth failure")
}

// jsonAuthFail sends a message back to the client if the http auth is rejected.
func jsonAuthFail(w http.ResponseWriter) {
	w.Header().Add("WWW-Authenticate", `Basic realm="qitmeer RPC"`)
	http.Error(w, "401 Unauthorized.", http.StatusUnauthorized)
}

// CodecOption specifies which type of messages this codec supports
type CodecOption int

const (
	// OptionMethodInvocation is an indication that the codec supports RPC method calls
	OptionMethodInvocation CodecOption = 1 << iota

	// OptionSubscriptions is an indication that the codec suports RPC notifications
	OptionSubscriptions = 1 << iota // support pub sub
)

// jsonRPCRead handles reading and responding to RPC messages.
func (s *RpcServer) jsonRPCRead(w http.ResponseWriter, r *http.Request) {
	if atomic.LoadInt32(&s.run) != 1 { // server stopped
		return
	}
	// discard dumb empty requests
	if emptyRequest(r) {
		log.Trace("Discard empty request", "from", r.RemoteAddr, "request", r)
		return
	}
	// validate request
	if code, err := validateRequest(r); err != nil {
		http.Error(w, err.Error(), code)
		return
	}
	// All checks passed, create a codec that reads direct from the request body
	// untilEOF and writes the response to w and order the server to process a
	// single request.
	ctx := r.Context()
	ctx = context.WithValue(ctx, "remote", r.RemoteAddr)
	ctx = context.WithValue(ctx, "scheme", r.Proto)
	ctx = context.WithValue(ctx, "local", r.Host)

	// Read and close the JSON-RPC request body from the caller.
	body := io.LimitReader(r.Body, maxRequestContentLength)
	codec := NewJSONCodec(&httpReadWriteNopCloser{body, w})
	defer codec.Close()

	log.Trace("jsonRPCRead", "body", body, "codec", codec)

	s.ServeSingleRequest(ctx, codec, OptionMethodInvocation)
}

// ServeSingleRequest reads and processes a single RPC request from the given codec. It will not
// close the codec unless a non-recoverable error has occurred. Note, this method will return after
// a single request has been processed!
func (s *RpcServer) ServeSingleRequest(ctx context.Context, codec ServerCodec, options CodecOption) {
	s.serveRequest(ctx, codec, true, options)
}

// serveRequest will reads requests from the codec, calls the RPC callback and
// writes the response to the given codec.
//
// If singleShot is true it will process a single request, otherwise it will handle
// requests until the codec returns an error when reading a request (in most cases
// an EOF). It executes requests in parallel when singleShot is false.
func (s *RpcServer) serveRequest(ctx context.Context, codec ServerCodec, singleShot bool, options CodecOption) error {
	var pend sync.WaitGroup

	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Error(string(buf))
		}
		s.codecsMu.Lock()
		s.codecs.Remove(codec)
		s.codecsMu.Unlock()
	}()

	//	ctx, cancel := context.WithCancel(context.Background())
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// if the codec supports notification include a notifier that callbacks can use
	// to send notification to clients. It is tied to the codec/connection. If the
	// connection is closed the notifier will stop and cancels all active subscriptions.
	if options&OptionSubscriptions == OptionSubscriptions {
		ctx = context.WithValue(ctx, notifierKey{}, newNotifier(codec))
	}
	s.codecsMu.Lock()
	if atomic.LoadInt32(&s.run) != 1 { // server stopped
		s.codecsMu.Unlock()
		return &shutdownError{}
	}
	s.codecs.Add(codec)
	s.codecsMu.Unlock()

	// test if the server is ordered to stop
	for atomic.LoadInt32(&s.run) == 1 {
		reqs, batch, err := s.readRequest(codec)
		if err != nil {
			// If a parsing error occurred, send an error
			if err.Error() != "EOF" {
				log.Debug(fmt.Sprintf("read error %v\n", err))
				codec.Write(codec.CreateErrorResponse(nil, err))
			}
			// Error or end of stream, wait for requests and tear down
			pend.Wait()
			return nil
		}

		// check if server is ordered to shutdown and return an error
		// telling the client that his request failed.
		if atomic.LoadInt32(&s.run) != 1 {
			err = &shutdownError{}
			if batch {
				resps := make([]interface{}, len(reqs))
				for i, r := range reqs {
					resps[i] = codec.CreateErrorResponse(&r.id, err)
				}
				codec.Write(resps)
			} else {
				codec.Write(codec.CreateErrorResponse(&reqs[0].id, err))
			}
			return nil
		}
		// If a single shot request is executing, run and return immediately
		if singleShot {
			if batch {
				s.execBatch(ctx, codec, reqs)
			} else {
				s.exec(ctx, codec, reqs[0])
			}
			return nil
		}
		// For multi-shot connections, start a goroutine to serve and loop back
		pend.Add(1)

		go func(reqs []*serverRequest, batch bool) {
			defer pend.Done()
			if batch {
				s.execBatch(ctx, codec, reqs)
			} else {
				s.exec(ctx, codec, reqs[0])
			}
		}(reqs, batch)
	}
	return nil
}

// httpReadWriteNopCloser wraps a io.Reader and io.Writer with a NOP Close method.
type httpReadWriteNopCloser struct {
	io.Reader
	io.Writer
}

// Close does nothing and returns always nil
func (t *httpReadWriteNopCloser) Close() error {
	return nil
}

// readRequest requests the next (batch) request from the codec. It will return the collection
// of requests, an indication if the request was a batch, the invalid request identifier and an
// error when the request could not be read/parsed.
func (s *RpcServer) readRequest(codec ServerCodec) ([]*serverRequest, bool, Error) {
	reqs, batch, err := codec.ReadRequestHeaders()
	if err != nil {
		return nil, batch, err
	}

	requests := make([]*serverRequest, len(reqs))

	// verify requests
	for i, r := range reqs {
		var ok bool
		var svc *service

		if r.err != nil {
			requests[i] = &serverRequest{id: r.id, err: r.err}
			continue
		}

		if r.isPubSub && strings.HasSuffix(r.method, unsubscribeMethodSuffix) {
			requests[i] = &serverRequest{id: r.id, isUnsubscribe: true}
			argTypes := []reflect.Type{reflect.TypeOf("")} // expect subscription id as first arg
			if args, err := codec.ParseRequestArguments(argTypes, r.params); err == nil {
				requests[i].args = args
			} else {
				requests[i].err = &invalidParamsError{err.Error()}
			}
			continue
		}

		if svc, ok = s.rpcSvcRegistry[r.service]; !ok { // rpc method isn't available
			requests[i] = &serverRequest{id: r.id, err: &methodNotFoundError{r.service, r.method}}
			continue
		}

		if r.isPubSub { // eth_subscribe, r.method contains the subscription method name
			if callb, ok := svc.subscriptions[r.method]; ok {
				requests[i] = &serverRequest{id: r.id, svcname: svc.svcNamespace, callb: callb}
				if r.params != nil && len(callb.argTypes) > 0 {
					argTypes := []reflect.Type{reflect.TypeOf("")}
					argTypes = append(argTypes, callb.argTypes...)
					if args, err := codec.ParseRequestArguments(argTypes, r.params); err == nil {
						requests[i].args = args[1:] // first one is service.method name which isn't an actual argument
					} else {
						requests[i].err = &invalidParamsError{err.Error()}
					}
				}
			} else {
				requests[i] = &serverRequest{id: r.id, err: &methodNotFoundError{r.service, r.method}}
			}
			continue
		}

		if callb, ok := svc.callbacks[r.method]; ok { // lookup RPC method
			requests[i] = &serverRequest{id: r.id, svcname: svc.svcNamespace, callb: callb}
			if r.params != nil && len(callb.argTypes) > 0 {
				if args, err := codec.ParseRequestArguments(callb.argTypes, r.params); err == nil {
					requests[i].args = args
				} else {
					requests[i].err = &invalidParamsError{err.Error()}
				}
			}
			continue
		}

		requests[i] = &serverRequest{id: r.id, err: &methodNotFoundError{r.service, r.method}}
	}

	return requests, batch, nil
}

// execBatch executes the given requests and writes the result back using the codec.
// It will only write the response back when the last request is processed.
func (s *RpcServer) execBatch(ctx context.Context, codec ServerCodec, requests []*serverRequest) {
	responses := make([]interface{}, len(requests))
	var callbacks []func()
	for i, req := range requests {
		if req.err != nil {
			responses[i] = codec.CreateErrorResponse(&req.id, req.err)
		} else {
			var callback func()
			if responses[i], callback = s.handle(ctx, codec, req); callback != nil {
				callbacks = append(callbacks, callback)
			}
		}
	}

	if err := codec.Write(responses); err != nil {
		log.Error(fmt.Sprintf("%v\n", err))
		codec.Close()
	}

	// when request holds one of more subscribe requests this allows these subscriptions to be activated
	for _, c := range callbacks {
		c()
	}
}

// exec executes the given request and writes the result back using the codec.
func (s *RpcServer) exec(ctx context.Context, codec ServerCodec, req *serverRequest) {
	var response interface{}
	var callback func()
	if req.err != nil {
		response = codec.CreateErrorResponse(&req.id, req.err)
	} else {
		response, callback = s.handle(ctx, codec, req)
	}

	if err := codec.Write(response); err != nil {
		log.Error(fmt.Sprintf("%v\n", err))
		codec.Close()
	}

	// when request was a subscribe request this allows these subscriptions to be actived
	if callback != nil {
		callback()
	}
}

// handle executes a request and returns the response from the callback.
func (s *RpcServer) handle(ctx context.Context, codec ServerCodec, req *serverRequest) (interface{}, func()) {
	if req.err != nil {
		return codec.CreateErrorResponse(&req.id, req.err), nil
	}

	if req.isUnsubscribe { // cancel subscription, first param must be the subscription id
		if len(req.args) >= 1 && req.args[0].Kind() == reflect.String {
			notifier, supported := NotifierFromContext(ctx)
			if !supported { // interface doesn't support subscriptions (e.g. http)
				return codec.CreateErrorResponse(&req.id, &callbackError{ErrNotificationsUnsupported.Error()}), nil
			}

			subid := ID(req.args[0].String())
			if err := notifier.unsubscribe(subid); err != nil {
				return codec.CreateErrorResponse(&req.id, &callbackError{err.Error()}), nil
			}

			return codec.CreateResponse(req.id, true), nil
		}
		return codec.CreateErrorResponse(&req.id, &invalidParamsError{"Expected subscription id as first argument"}), nil
	}

	if req.callb.isSubscribe {
		subid, err := s.createSubscription(ctx, codec, req)
		if err != nil {
			return codec.CreateErrorResponse(&req.id, &callbackError{err.Error()}), nil
		}

		// active the subscription after the sub id was successfully sent to the client
		activateSub := func() {
			notifier, _ := NotifierFromContext(ctx)
			notifier.activate(subid, req.svcname)
		}

		return codec.CreateResponse(req.id, subid), activateSub
	}

	// regular RPC call, prepare arguments
	if len(req.args) != len(req.callb.argTypes) {
		rpcErr := &invalidParamsError{fmt.Sprintf("%s%s%s expects %d parameters, got %d",
			req.svcname, serviceMethodSeparator, req.callb.method.Name,
			len(req.callb.argTypes), len(req.args))}
		return codec.CreateErrorResponse(&req.id, rpcErr), nil
	}

	arguments := []reflect.Value{req.callb.receiver}
	if req.callb.hasCtx {
		arguments = append(arguments, reflect.ValueOf(ctx))
	}
	if len(req.args) > 0 {
		arguments = append(arguments, req.args...)
	}

	s.AddRequstStatus(req)
	// execute RPC method and return result
	reply := req.callb.method.Func.Call(arguments)
	s.RemoveRequstStatus(req)
	if len(reply) == 0 {
		return codec.CreateResponse(req.id, nil), nil
	}
	if req.callb.errPos >= 0 { // test if method returned an error
		if !reply[req.callb.errPos].IsNil() {
			e := reply[req.callb.errPos].Interface().(error)
			res := codec.CreateErrorResponse(&req.id, &callbackError{e.Error()})
			return res, nil
		}
	}
	return codec.CreateResponse(req.id, reply[0].Interface()), nil
}

// createSubscription will call the subscription callback and returns the subscription id or error.
func (s *RpcServer) createSubscription(ctx context.Context, c ServerCodec, req *serverRequest) (ID, error) {
	// subscription have as first argument the context following optional arguments
	args := []reflect.Value{req.callb.receiver, reflect.ValueOf(ctx)}
	args = append(args, req.args...)
	reply := req.callb.method.Func.Call(args)

	if !reply[1].IsNil() { // subscription creation failed
		return "", reply[1].Interface().(error)
	}

	return reply[0].Interface().(*Subscription).ID, nil
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
		svcNamespace:  namespace,
		svcType:       typ,
		callbacks:     calls,
		subscriptions: subs,
	}
	s.rpcSvcRegistry[svc.svcNamespace] = &svc
	return nil
}

func (s *RpcServer) RequestedProcessShutdown() chan struct{} {
	return s.requestProcessShutdown
}

func (s *RpcServer) AddRequstStatus(sReq *serverRequest) {
	s.reqStatusLock.Lock()
	defer s.reqStatusLock.Unlock()
	key := fmt.Sprintf("%s_%s", sReq.svcname, sReq.callb.method.Name)
	rs, ok := s.ReqStatus[key]
	if !ok {
		rs, _ = NewRequestStatus(sReq)
		s.ReqStatus[rs.GetName()] = rs
	} else {
		rs.AddRequst(sReq)
	}
}

func (s *RpcServer) RemoveRequstStatus(sReq *serverRequest) {
	s.reqStatusLock.Lock()
	defer s.reqStatusLock.Unlock()
	key := fmt.Sprintf("%s_%s", sReq.svcname, sReq.callb.method.Name)
	rs, ok := s.ReqStatus[key]
	if !ok {
		return
	} else {
		rs.RemoveRequst(sReq)
	}
}
