package testutils

import (
	"bufio"
	"fmt"
	"github.com/Qitmeer/qitmeer/rpc"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
)

// the configuration of the node
type nodeConfig struct {
	program   string
	rpclisten string
	rpcuser   string
	rpcpass   string
	homeDir   string
	dataDir   string
	logDir    string
	keyFile   string
	certFile  string
	extraArgs []string
}

func newNodeConfig(homeDir string, extraArgs []string) *nodeConfig {
	c := &nodeConfig{
		program:   "qitmeer",
		rpclisten: "127.0.0.1:12345",
		rpcuser:   "testuser",
		rpcpass:   "testpass",
		homeDir:   homeDir,
		dataDir:   filepath.Join(homeDir, "data"),
		logDir:    filepath.Join(homeDir, "log"),
		keyFile:   filepath.Join(homeDir, "rpc.key"),
		certFile:  filepath.Join(homeDir, "rpc.cert"),
		extraArgs: extraArgs,
	}
	return c
}

// args return the arguments list build from the nodeConfig
// which be used to launch the qitmeer node
func (n *nodeConfig) args() []string {
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

// node is the wrapper of a qitmeer node process. which contains all necessary
// configure information required to manage a qitmeer full node process.
type node struct {
	t      *testing.T
	config *nodeConfig
	cmd    *exec.Cmd
	wg     sync.WaitGroup
}

// create an new node instance
func newNode(t *testing.T, config *nodeConfig) (*node, error) {
	// test if home directory exist
	if _, err := os.Stat(config.homeDir); os.IsNotExist(err) {
		return nil, err
	}
	// create rpc cert and key file
	if err := rpc.GenCertPair(config.certFile, config.keyFile); err != nil {
		return nil, err
	}
	return &node{
		t:      t,
		config: config,
		cmd:    exec.Command(config.program, config.args()...),
	}, nil
}

func (n *node) redirectOutput(reader io.ReadCloser) error {
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		r := bufio.NewReader(reader)
		for {
			l, err := r.ReadString('\n')
			if err != nil || err == io.EOF {
				return
			}
			n.t.Logf("qitmeer-harness-node: %s", l)
		}
	}()
	return nil
}

// start up the node instance which works as the wrapped qitmeer process
func (n *node) start() error {
	// redirect stdout
	if stdout, err := n.cmd.StdoutPipe(); err != nil {
		return err
	} else {
		n.redirectOutput(stdout)
	}
	// redirect stderr
	if stderr, err := n.cmd.StderrPipe(); err != nil {
		return err
	} else {
		n.redirectOutput(stderr)
	}

	// Launch command
	n.t.Logf("start node from %v", n.config.homeDir)
	if err := n.cmd.Start(); err != nil {
		return err
	}
	return nil
}

func (n *node) stop() error {
	n.t.Logf("stop node from %v", n.config.homeDir)
	if err := n.cmd.Process.Signal(os.Interrupt); err != nil {
		return err
	}
	if err := n.cmd.Wait(); err != nil {
		return err
	}
	n.wg.Wait()
	return nil
}
