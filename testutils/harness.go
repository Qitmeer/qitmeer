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
	// the global variables for a protected harness state
	// the number of initialized harness instance
	numOfHarnessInstances = 0

	// process id
	pid = os.Getpid()

	// protect the global harness state variables
	harnessStateMutex sync.RWMutex
)

// Harness manage an embedded qitmeer node process for running the rpc driven
// integration tests. The active qitmeer node will typically be run in privnet
// mode in order to allow for easy block generation. Harness handles the node
// start/shutdown and any temporary directories need to be created.
type Harness struct {
}

// setup func initialize the test state.
// 1. start the qitmeer node according to the net params
// 2. setup the rpc clint so that the rpc call can be sent to the node
// 3. generate a test block dag by configuration (optional, may empty dag by config)
func (*Harness) Setup() error {
	return nil
}

// teardown func stop the running test, stop the rpc client shutdown the node,
// kill any related processes if need and clean up the temporary data folder
func (*Harness) Teardown() error {
	return nil
}

// NewHarness func creates an new instance of test harness with provided network params.
// The args is the arguments list that are used when setup a qitmeer node. In the most
// case, it should be set to nil if no extra args need to add on the default starting up.
func NewHarness(t *testing.T, net *params.Params, args []string) (*Harness, error) {
	harnessStateMutex.Lock()
	defer harnessStateMutex.Unlock()
	h := Harness{}
	// create temporary folder
	testDir, err := ioutil.TempDir("", "test-harness-"+strconv.Itoa(numOfHarnessInstances))
	if err != nil {
		return nil, err
	}
	// create rpc cert and key
	certFile := filepath.Join(testDir, "rpc.cert")
	keyFile := filepath.Join(testDir, "rpc.key")
	if err := rpc.GenCertPair(certFile, keyFile); err != nil {
		return nil, err
	}
	numOfHarnessInstances++
	return &h, nil
}
