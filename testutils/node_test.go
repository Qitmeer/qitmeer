package testutils

import (
	"reflect"
	"testing"
)

func TestNewNodeCmdArgs(t *testing.T) {
	c := newNodeConfig("test",[]string{"--k1=v1","--k2=v2"})
 	node := newNode(t, c)
 	if node.cmd.Dir != "" {
		t.Errorf("failed to create node, expect %v but got %v", "", node.cmd.Dir)
	}
	args := []string{
		"qitmeer",
		"--rpclisten=127.0.0.1:12345",
		"--rpcuser=testuser",
		"--rpcpass=testpass",
		"--datadir=test/data",
		"--logdir=test/log",
		"--k1=v1",
		"--k2=v2",
		}
	if !reflect.DeepEqual(args, node.cmd.Args) {
		t.Errorf("failed to create node, expect %v but got %v", args, node.cmd.Args)
	}
}
