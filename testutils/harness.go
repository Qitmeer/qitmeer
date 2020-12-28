// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package testutils

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/core/protocol"
	"github.com/Qitmeer/qitmeer/params"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

const DefaultMaxRpcConnRetries = 10

var (
	// harness main-process id which shared for all harness instances
	harnessMainProcessId = os.Getpid()

	// the private harness instances map contains all initialized harnesses
	// which returned by the NewHarness func. and the instance will delete itself
	// from the map after the Teardown func has been called
	harnessInstances = make(map[string]*Harness)
	// protect the global harness state variables
	harnessStateMutex sync.RWMutex
)

// Harness manage an embedded qitmeer node process for running the rpc driven
// integration tests. The active qitmeer node will typically be run in privnet
// mode in order to allow for easy block generation. Harness handles the node
// start/shutdown and any temporary directories need to be created.
type Harness struct {
	// the temporary directory created when the Harness instance initialized
	// its also used as the unique id of the harness instance, its in the
	// format of `test-harness-<num>-*`
	instanceDir string
	// the qitmeer node process
	node *node
	// the rpc client to the qitmeer node in the Harness instance.
	Client *Client
	// the maximized attempts try to establish the rpc connection
	maxRpcConnRetries int
}

func (h *Harness) Id() string {
	return h.instanceDir
}

// Setup func initialize the test state.
// 1. start the qitmeer node according to the net params
// 2. setup the rpc clint so that the rpc call can be sent to the node
// 3. generate a test block dag by configuration (optional, may empty dag by config)
func (h *Harness) Setup() error {
	// start up the qitmeer node
	if err := h.node.start(); err != nil {
		return err
	}
	if err := h.connectRPCClient(); err != nil {
		return err
	}
	return nil
}

// connectRPCClient attempts to establish an RPC connection to the Harness instance.
// If the initial attempt fails, this function will retry h.maxRpcConnRetries times,
// this function returns with an error if all retries failed.
func (h *Harness) connectRPCClient() error {
	var client *Client
	var err error

	url, user, pass := h.node.config.rpclisten, h.node.config.rpcuser, h.node.config.rpcpass
	for i := 0; i < h.maxRpcConnRetries; i++ {
		if client, err = Dial("https://"+url, user, pass); err != nil {
			time.Sleep(time.Duration(i) * 50 * time.Millisecond)
			continue
		}
		break
	}
	if client == nil || err != nil {
		return fmt.Errorf("failed to establish rpc client connection: %v", err)
	}

	h.Client = client
	return nil
}

// Teardown func the concurrent safe wrapper of teardown func
func (h *Harness) Teardown() error {
	harnessStateMutex.Lock()
	defer harnessStateMutex.Unlock()
	return h.teardown()
}

// teardown func stop the running test, stop the rpc client shutdown the node,
// kill any related processes if need and clean up the temporary data folder
// NOTE: the func is NOT concurrent safe. see also the Teardown func
func (h *Harness) teardown() error {
	if err := h.node.stop(); err != nil {
		return err
	}
	if err := os.RemoveAll(h.instanceDir); err != nil {
		return err
	}

	delete(harnessInstances, h.instanceDir)
	return nil
}

// NewHarness func creates an new instance of test harness with provided network params.
// The args is the arguments list that are used when setup a qitmeer node. In the most
// case, it should be set to nil if no extra args need to add on the default starting up.
func NewHarness(t *testing.T, params *params.Params, args ...string) (*Harness, error) {
	harnessStateMutex.Lock()
	defer harnessStateMutex.Unlock()

	// create temporary folder
	testDir, err := ioutil.TempDir("", "test-harness-"+strconv.Itoa(len(harnessInstances))+"-*")
	if err != nil {
		return nil, err
	}

	// setup network type
	extraArgs := []string{}
	switch params.Net {
	case protocol.MainNet:
		//do nothing for mainnet which is by default
	case protocol.MixNet:
		extraArgs = append(extraArgs, "--mixnet")
	case protocol.TestNet:
		extraArgs = append(extraArgs, "--testnet")
	case protocol.PrivNet:
		extraArgs = append(extraArgs, "--privnet")
	default:
		return nil, fmt.Errorf("unknown network type %v", params.Net)
	}

	extraArgs = append(extraArgs, args...)

	// create node config & initialize the node process
	config := newNodeConfig(testDir, extraArgs)

	// use auto-genereated p2p/rpc port settings instead of default
	config.listen, config.rpclisten = genListenArgs()

	newNode, err := newNode(t, config)
	if err != nil {
		return nil, err
	}

	h := Harness{
		instanceDir:       testDir,
		node:              newNode,
		maxRpcConnRetries: DefaultMaxRpcConnRetries,
	}
	harnessInstances[h.instanceDir] = &h
	return &h, nil
}

// TearDownAll func teardown all Harness Instances
func TearDownAll() error {
	harnessStateMutex.Lock()
	defer harnessStateMutex.Unlock()
	for _, h := range harnessInstances {
		if err := h.teardown(); err != nil {
			return err
		}
	}
	return nil
}

// AllHarnesses func get all Harness instances
func AllHarnesses() []*Harness {
	harnessStateMutex.RLock()
	defer harnessStateMutex.RUnlock()
	all := make([]*Harness, 0, len(harnessInstances))
	for _, h := range harnessInstances {
		all = append(all, h)
	}
	return all
}

const (
	// the minimum and maximum p2p and rpc port numbers used by a test harness.
	minP2PPort = 38200              // 38200 The min is inclusive
	maxP2PPort = minP2PPort + 10000 // 48199 The max is exclusive
	minRPCPort = maxP2PPort         // 48200
	maxRPCPort = minRPCPort + 10000 // 58199
)

// GenListenArgs returns auto generated args for p2p listen and rpc listen in the format of
// ["--listen=127.0.0.1:12345", --rpclisten=127.0.0.1:12346"].
// in order to support multiple test node running at the same time.
func genListenArgs() (string, string) {
	localhost := "127.0.0.1"
	genPort := func(min, max int) string {
		port := min + len(harnessInstances) + (42 * harnessMainProcessId % (max - min))
		return strconv.Itoa(port)
	}
	p2p := net.JoinHostPort(localhost, genPort(minP2PPort, maxP2PPort))
	rpc := net.JoinHostPort(localhost, genPort(minRPCPort, maxRPCPort))
	return p2p, rpc
}
