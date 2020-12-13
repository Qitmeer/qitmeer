// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package testutils

// Harness manage an embedded qitmeer node process for running the rpc driven
// integration tests. The active qitmeer node will typically be run in privnet
// mode in order to allow for easy block generation. Harness handles the node
// start/shutdown and any temporary directories need to be created.
type Harness struct {
}
