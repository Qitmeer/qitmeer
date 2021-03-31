// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package testutils_test

import (
	"github.com/Qitmeer/qitmeer/config"
	"github.com/Qitmeer/qitmeer/core/event"
	"github.com/Qitmeer/qitmeer/rpc"
	"github.com/Qitmeer/qitmeer/services/mempool"
	"github.com/Qitmeer/qitmeer/testutils"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

const (
	defaultConfigFilename         = "qitmeer.conf"
	defaultDataDirname            = "data"
	defaultLogLevel               = "info"
	defaultDebugPrintOrigins      = false
	defaultLogDirname             = "logs"
	defaultLogFilename            = "qitmeer.log"
	defaultGenerate               = false
	defaultBlockMinSize           = 0
	defaultBlockMaxSize           = 375000
	defaultMaxRPCClients          = 10
	defaultMaxRPCWebsockets       = 25
	defaultMaxRPCConcurrentReqs   = 20
	defaultMaxPeers               = 30
	defaultMiningStateSync        = false
	defaultMaxInboundPeersPerHost = 10 // The default max total of inbound peer for host
	defaultTrickleInterval        = 10 * time.Second
	defaultCacheInvalidTx         = false
	defaultSigCacheMaxSize        = 100000
)

var (
	defaultHomeDir, _  = ioutil.TempDir("", "qitmeer-test-rpc-server")
	defaultConfigFile  = filepath.Join(defaultHomeDir, defaultConfigFilename)
	defaultDataDir     = filepath.Join(defaultHomeDir, defaultDataDirname)
	defaultDbType      = "ffldb"
	defaultLogDir      = filepath.Join(defaultHomeDir, defaultLogDirname)
	defaultRPCKeyFile  = filepath.Join(defaultHomeDir, "rpc.key")
	defaultRPCCertFile = filepath.Join(defaultHomeDir, "rpc.cert")
	defaultDAGType     = "phantom"
)

var (
	// Default config.
	cfg = config.Config{
		HomeDir:              defaultHomeDir,
		ConfigFile:           defaultConfigFile,
		DebugLevel:           defaultLogLevel,
		DebugPrintOrigins:    defaultDebugPrintOrigins,
		DataDir:              defaultDataDir,
		LogDir:               defaultLogDir,
		DbType:               defaultDbType,
		RPCKey:               defaultRPCKeyFile,
		RPCCert:              defaultRPCCertFile,
		RPCMaxClients:        defaultMaxRPCClients,
		RPCMaxWebsockets:     defaultMaxRPCWebsockets,
		RPCMaxConcurrentReqs: defaultMaxRPCConcurrentReqs,
		Generate:             defaultGenerate,
		MaxPeers:             defaultMaxPeers,
		MinTxFee:             mempool.DefaultMinRelayTxFee,
		BlockMinSize:         defaultBlockMinSize,
		BlockMaxSize:         defaultBlockMaxSize,
		SigCacheMaxSize:      defaultSigCacheMaxSize,
		MiningStateSync:      defaultMiningStateSync,
		DAGType:              defaultDAGType,
		Banning:              false,
		MaxInbound:           defaultMaxInboundPeersPerHost,
		CacheInvalidTx:       defaultCacheInvalidTx,
		NTP:                  false,
		RPCListeners:         []string{"127.0.0.1:5555"},
		RPCUser:              "test",
		RPCPass:              "pass",
	}
)

func newTestServer(t *testing.T) *rpc.RpcServer {
	server, err := rpc.NewRPCServer(&cfg, &event.Feed{})
	if err != nil {
		t.Errorf("failed to initialize rpc server: %v", err)
	}
	return server
}

func TestRpcClient(t *testing.T) {
	server := newTestServer(t)
	defer server.Stop()
	if err := server.Start(); err != nil {
		t.Errorf("start rpc server error : %v", err)
	}
	if err := server.RegisterService("test", new(testutils.TestService)); err != nil {
		t.Errorf("register test service err : %v", err)
	}

	client, err := testutils.Dial("https://"+cfg.RPCListeners[0], cfg.RPCUser, cfg.RPCPass, cfg.RPCCert)
	if err != nil {
		t.Errorf("Dial client error: %v", err)
	}
	var result testutils.EchoResult
	comp := &testutils.Complex{0, 0, "zero"}
	if err := client.Call(&result, "test_echo", "test", 1, comp); err != nil {
		t.Errorf("client call execute error: %v", err)
	}
	expect := testutils.EchoResult{
		"TEST", 1, &testutils.Complex{0, 0, "ZERO"},
	}
	// now deep equal should ok for every field
	if !reflect.DeepEqual(expect, result) {
		t.Errorf("call echo failed, expect %v but got %v", expect, result)
	}

}
