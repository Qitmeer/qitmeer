// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package testutils

import (
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/rpc"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
)

var (
	// process id
	pid = os.Getpid()

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
}

func (h *Harness) Id() string {
	return h.instanceDir
}

// Setup func initialize the test state.
// 1. start the qitmeer node according to the net params
// 2. setup the rpc clint so that the rpc call can be sent to the node
// 3. generate a test block dag by configuration (optional, may empty dag by config)
func (*Harness) Setup() error {
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
	if err := os.RemoveAll(h.instanceDir); err != nil {
		return err
	}
	delete(harnessInstances, h.instanceDir)
	return nil
}

// NewHarness func creates an new instance of test harness with provided network params.
// The args is the arguments list that are used when setup a qitmeer node. In the most
// case, it should be set to nil if no extra args need to add on the default starting up.
func NewHarness(t *testing.T, net *params.Params, args []string) (*Harness, error) {
	harnessStateMutex.Lock()
	defer harnessStateMutex.Unlock()
	// create temporary folder
	testDir, err := ioutil.TempDir("", "test-harness-"+strconv.Itoa(len(harnessInstances))+"-*")
	if err != nil {
		return nil, err
	}
	// create rpc cert and key
	certFile := filepath.Join(testDir, "rpc.cert")
	keyFile := filepath.Join(testDir, "rpc.key")
	if err := rpc.GenCertPair(certFile, keyFile); err != nil {
		return nil, err
	}
	// initialize the node process
	newNode := &node{}
	h := Harness{
		instanceDir: testDir,
		node:        newNode,
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
