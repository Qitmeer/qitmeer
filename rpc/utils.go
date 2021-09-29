package rpc

import (
	"bufio"
	"context"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/tls"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/marshal"
	"github.com/Qitmeer/qitmeer/common/math"
	"github.com/Qitmeer/qitmeer/common/network"
	"github.com/Qitmeer/qitmeer/common/roughtime"
	"github.com/Qitmeer/qitmeer/common/util"
	"github.com/Qitmeer/qitmeer/config"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/crypto/certgen"
	"github.com/Qitmeer/qitmeer/rpc/client/cmds"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

var (
	subscriptionIDGenMu sync.Mutex
	subscriptionIDGen   = idGenerator()
)

// Is this an exported - upper case - name?
func isExported(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(r)
}

// Is this type exported or a builtin?
func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return isExported(t.Name()) || t.PkgPath() == ""
}

var contextType = reflect.TypeOf((*context.Context)(nil)).Elem()

// isContextType returns an indication if the given t is of context.Context or *context.Context type
func isContextType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t == contextType
}

var errorType = reflect.TypeOf((*error)(nil)).Elem()

// Implements this type the error interface
func isErrorType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Implements(errorType)
}

var subscriptionType = reflect.TypeOf((*Subscription)(nil)).Elem()

// isSubscriptionType returns an indication if the given t is of Subscription or *Subscription type
func isSubscriptionType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t == subscriptionType
}

// isPubSub tests whether the given method has as as first argument a context.Context
// and returns the pair (Subscription, error)
func isPubSub(methodType reflect.Type) bool {
	// numIn(0) is the receiver type
	if methodType.NumIn() < 2 || methodType.NumOut() != 2 {
		return false
	}

	return isContextType(methodType.In(1)) &&
		isSubscriptionType(methodType.Out(0)) &&
		isErrorType(methodType.Out(1))
}

// formatName will convert to first character to lower case
func formatName(name string) string {
	ret := []rune(name)
	if len(ret) > 0 {
		ret[0] = unicode.ToLower(ret[0])
	}
	return string(ret)
}

// suitableCallbacks iterates over the methods of the given type. It will determine if a method satisfies the criteria
// for a RPC callback or a subscription callback and adds it to the collection of callbacks or subscriptions. See server
// documentation for a summary of these criteria.
func suitableCallbacks(rcvr reflect.Value, typ reflect.Type) (callbacks, subscriptions) {
	callbacks := make(callbacks)
	subscriptions := make(subscriptions)

METHODS:
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := formatName(method.Name)
		if method.PkgPath != "" { // method must be exported
			continue
		}

		var h callback
		h.isSubscribe = isPubSub(mtype)
		h.receiver = rcvr
		h.method = method
		h.errPos = -1

		firstArg := 1
		numIn := mtype.NumIn()
		if numIn >= 2 && mtype.In(1) == contextType {
			h.hasCtx = true
			firstArg = 2
		}

		if h.isSubscribe {
			h.argTypes = make([]reflect.Type, numIn-firstArg) // skip rcvr type
			for i := firstArg; i < numIn; i++ {
				argType := mtype.In(i)
				if isExportedOrBuiltinType(argType) {
					h.argTypes[i-firstArg] = argType
				} else {
					continue METHODS
				}
			}

			subscriptions[mname] = &h
			continue METHODS
		}

		// determine method arguments, ignore first arg since it's the receiver type
		// Arguments must be exported or builtin types
		h.argTypes = make([]reflect.Type, numIn-firstArg)
		for i := firstArg; i < numIn; i++ {
			argType := mtype.In(i)
			if !isExportedOrBuiltinType(argType) {
				continue METHODS
			}
			h.argTypes[i-firstArg] = argType
		}

		// check that all returned values are exported or builtin types
		for i := 0; i < mtype.NumOut(); i++ {
			if !isExportedOrBuiltinType(mtype.Out(i)) {
				continue METHODS
			}
		}

		// when a method returns an error it must be the last returned value
		h.errPos = -1
		for i := 0; i < mtype.NumOut(); i++ {
			if isErrorType(mtype.Out(i)) {
				h.errPos = i
				break
			}
		}

		if h.errPos >= 0 && h.errPos != mtype.NumOut()-1 {
			continue METHODS
		}

		switch mtype.NumOut() {
		case 0, 1, 2:
			if mtype.NumOut() == 2 && h.errPos == -1 { // method must one return value and 1 error
				continue METHODS
			}
			callbacks[mname] = &h
		}
	}

	return callbacks, subscriptions
}

// idGenerator helper utility that generates a (pseudo) random sequence of
// bytes that are used to generate identifiers.
func idGenerator() *rand.Rand {
	if seed, err := binary.ReadVarint(bufio.NewReader(crand.Reader)); err == nil {
		return rand.New(rand.NewSource(seed))
	}
	return rand.New(rand.NewSource(int64(roughtime.Now().Nanosecond())))
}

// NewID generates a identifier that can be used as an identifier in the RPC interface.
// e.g. filter and subscription identifier.
func NewID() ID {
	subscriptionIDGenMu.Lock()
	defer subscriptionIDGenMu.Unlock()

	id := make([]byte, 16)
	for i := 0; i < len(id); i += 7 {
		val := subscriptionIDGen.Int63()
		for j := 0; i+j < len(id) && j < 7; j++ {
			id[i+j] = byte(val)
			val >>= 8
		}
	}

	rpcId := hex.EncodeToString(id)
	// rpc ID's are RPC quantities, no leading zero's and 0 is 0x0
	rpcId = strings.TrimLeft(rpcId, "0")
	if rpcId == "" {
		rpcId = "0"
	}

	return ID("0x" + rpcId)
}

// parseListeners splits the list of listen addresses passed in addrs into
// IPv4 and IPv6 slices and returns them.  This allows easy creation of the
// listeners on the correct interface "tcp4" and "tcp6".  It also properly
// detects addresses which apply to "all interfaces" and adds the address to
// both slices.
func parseListeners(cfg *config.Config, addrs []string) ([]net.Listener, error) {

	ipListenAddrs, err := network.ParseListeners(addrs)
	if err != nil {
		return nil, err
	}
	listenFunc := net.Listen
	if !cfg.DisableRPC && !cfg.DisableTLS {
		// Generate the TLS cert and key file if both don't already
		// exist.
		if !util.FileExists(cfg.RPCKey) && !util.FileExists(cfg.RPCCert) {
			err := GenCertPair(cfg.RPCCert, cfg.RPCKey)
			if err != nil {
				return nil, err
			}
		}
		keypair, err := tls.LoadX509KeyPair(cfg.RPCCert, cfg.RPCKey)
		if err != nil {
			return nil, err
		}

		tlsConfig := tls.Config{
			Certificates: []tls.Certificate{keypair},
			MinVersion:   tls.VersionTLS12,
		}

		// Change the standard net.Listen function to the tls one.
		listenFunc = func(net string, laddr string) (net.Listener, error) {
			return tls.Listen(net, laddr, &tlsConfig)
		}
	}
	listeners := make([]net.Listener, 0, len(ipListenAddrs))

	for _, addr := range ipListenAddrs {
		listener, err := listenFunc(addr.Network(), addr.String())
		if err != nil {
			log.Warn("Can't listen on", "addr", addr, "error", err)
			continue
		}
		listeners = append(listeners, listener)
	}

	if len(listeners) == 0 {
		return nil, fmt.Errorf("No valid listen address")
	}
	return listeners, nil
}

// genCertPair generates a key/cert pair to the paths provided.
func GenCertPair(certFile, keyFile string) error {
	log.Info("Generating TLS certificates...")

	org := "Qitmeer autogenerated cert"
	validUntil := roughtime.Now().Add(10 * 365 * 24 * time.Hour)
	cert, key, err := certgen.NewTLSCertPair(elliptic.P521(), org,
		validUntil, nil)
	if err != nil {
		return err
	}

	// Write cert and key files.
	if err = ioutil.WriteFile(certFile, cert, 0644); err != nil {
		return err
	}
	if err = ioutil.WriteFile(keyFile, key, 0600); err != nil {
		os.Remove(certFile)
		return err
	}

	log.Info("Done generating TLS certificates")
	return nil
}

type RequestStatus struct {
	Service      string
	Method       string
	TotalCalls   uint
	TotalTime    time.Duration
	MaxTime      time.Duration
	MinTime      time.Duration
	MaxTimeReqID string
	MinTimeReqID string

	Requests []*serverRequest
}

func (rs *RequestStatus) GetName() string {
	return rs.Service + "_" + rs.Method
}

func (rs *RequestStatus) AddRequst(sReq *serverRequest) {
	for _, v := range rs.Requests {
		if v == sReq {
			return
		}
	}
	rs.Requests = append(rs.Requests, sReq)
	rs.TotalCalls++
	sReq.time = roughtime.Now()
	log.Debug(fmt.Sprintf("Start RPC Call (id:%s method:%s)", sReq.id, rs.GetName()))
}

func (rs *RequestStatus) RemoveRequst(sReq *serverRequest) {
	for i := 0; i < len(rs.Requests); i++ {
		if rs.Requests[i] == sReq {
			cost := roughtime.Since(sReq.time)
			rs.TotalTime += cost
			rs.Requests = append(rs.Requests[:i], rs.Requests[i+1:]...)

			if cost > rs.MaxTime {
				rs.MaxTime = cost
				rs.MaxTimeReqID = fmt.Sprintf("%s", sReq.id)
			}
			if cost < rs.MinTime {
				rs.MinTime = cost
				rs.MinTimeReqID = fmt.Sprintf("%s", sReq.id)
			}
			log.Debug(fmt.Sprintf("End RPC Call (id:%s method:%s)", sReq.id, rs.GetName()))
			return
		}
	}
}

func (rs *RequestStatus) ToJson() *cmds.JsonRequestStatus {
	rsj := cmds.JsonRequestStatus{Name: rs.GetName(), TotalCalls: int(rs.TotalCalls),
		TotalTime: rs.TotalTime.String(), AverageTime: "", RunningNum: len(rs.Requests)}
	aTime := rs.TotalTime / time.Duration(rs.TotalCalls)
	rsj.AverageTime = aTime.String()
	rsj.MaxTime = rs.MaxTime.String()
	rsj.MinTime = rs.MinTime.String()
	rsj.MaxTimeReqID = rs.MaxTimeReqID
	rsj.MinTimeReqID = rs.MinTimeReqID
	return &rsj
}

func NewRequestStatus(sReq *serverRequest) (*RequestStatus, error) {
	rs := RequestStatus{sReq.svcname, sReq.callb.method.Name, 0,
		0, time.Duration(0), time.Duration(math.MaxInt64), "", "", []*serverRequest{}}
	rs.AddRequst(sReq)
	return &rs, nil
}

func createMarshalledReply(id, result interface{}, replyErr error) ([]byte, error) {
	var jsonErr *cmds.RPCError
	if replyErr != nil {
		if jErr, ok := replyErr.(*cmds.RPCError); ok {
			jsonErr = jErr
		} else {
			jsonErr = cmds.InternalRPCError(replyErr.Error(), "")
		}
	}

	return cmds.MarshalResponse(id, result, jsonErr)
}

func parseCmd(request *cmds.Request) *parsedRPCCmd {
	var parsedCmd parsedRPCCmd
	parsedCmd.id = request.ID
	parsedCmd.method = request.Method

	cmd, err := cmds.UnmarshalCmd(request)
	if err != nil {
		if jerr, ok := err.(cmds.Error); ok &&
			jerr.ErrorCode == cmds.ErrUnregisteredMethod {
			parsedCmd.err = cmds.ErrRPCMethodNotFound
			return &parsedCmd
		}

		parsedCmd.err = cmds.NewRPCError(
			cmds.ErrRPCInvalidParams.Code, err.Error())
		return &parsedCmd
	}

	parsedCmd.cmd = cmd
	return &parsedCmd
}

type parsedRPCCmd struct {
	id     interface{}
	method string
	cmd    interface{}
	err    *cmds.RPCError
}

func GetTxsHexFromBlock(block *types.SerializedBlock, duplicate bool) ([]string, error) {
	txs := []string{}
	for _, tx := range block.Transactions() {
		if duplicate {
			if tx.IsDuplicate {
				continue
			}
		}

		txhex, err := marshal.MessageToHex(tx.Tx)
		if err != nil {
			return nil, fmt.Errorf("Failed to marshal transaction:%v", tx.Hash().String())
		}
		txs = append(txs, txhex)
	}
	return txs, nil
}
