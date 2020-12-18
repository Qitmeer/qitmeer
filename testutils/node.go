package testutils

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"
)

// node is the wrapper of a qitmeer node process. which contains all necessary
// configure information required to manage a qitmeer full node process.
type node struct {
	t *testing.T
	config *nodeConfig
	cmd *exec.Cmd
}

// the configuration of the node
type nodeConfig struct {
	program   string
	rpclisten string
	rpcuser  string
	rpcpass  string
	homeDir   string
	dataDir   string
	logDir    string
	extraArgs []string
}
// args return the arguments list build from the nodeConfig
// which be used to launch the qitmeer node
func (n * nodeConfig) args() []string {
	args := []string{}
	if n.rpclisten != "" {
		args = append(args, fmt.Sprintf("--rpclisten=%s", n.rpclisten))
	}
	if n.rpcuser != "" {
		args = append(args, fmt.Sprintf("--rpcuser=%s", n.rpcuser))
	}
	if n.rpcpass != "" {
		args = append(args, fmt.Sprintf("--rpcpass=%s", n.rpcpass))
	}
	if n.dataDir != "" {
		args = append(args, fmt.Sprintf("--datadir=%s", n.dataDir))
	}
	if n.logDir != "" {
		args = append(args, fmt.Sprintf("--logdir=%s", n.logDir))
	}
	args = append(args, n.extraArgs...)
	return args
}

func newNode(t *testing.T, config *nodeConfig) *node {
	return &node{
		t,
		config,
		exec.Command(config.program, config.args()...),
	}
}

func newNodeConfig(homeDir string, extraArgs []string) *nodeConfig {
	c := &nodeConfig{
		program: "qitmeer",
		rpclisten: "127.0.0.1:12345",
		rpcuser: "testuser",
		rpcpass: "testpass",
		homeDir: homeDir,
		dataDir: filepath.Join(homeDir, "data"),
		logDir: filepath.Join(homeDir, "log"),
		extraArgs: extraArgs,
	}
	return c
}
