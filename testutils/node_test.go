package testutils

import (
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"testing"
)

func TestNewNodeCmdArgs(t *testing.T) {
	testDir, _ := ioutil.TempDir("", "test")
	defer os.RemoveAll(testDir)
	c := newNodeConfig(testDir, []string{"--k1=v1", "--k2=v2"})
	node, err := newNode(t, c)
	if err != nil {
		t.Errorf("failed to create new node : %v", err)
	}
	if node.cmd.Dir != "" {
		t.Errorf("failed to create node, expect %v but got %v", "", node.cmd.Dir)
	}
	args := []string{
		"qitmeer",
		"--rpclisten=127.0.0.1:12345",
		"--rpcuser=testuser",
		"--rpcpass=testpass",
		"--datadir=.*/test.*/data$",
		"--logdir=.*/test.*/log$",
		"--k1=v1",
		"--k2=v2",
	}
	data1 := args[4]
	data2 := node.cmd.Args[4]
	if !regexp.MustCompile(data1).MatchString(data2) {
		t.Errorf("failed to create node, expect %v but got %v", data1, data2)
	}
	log1 := args[5]
	log2 := node.cmd.Args[5]
	if !regexp.MustCompile(log1).MatchString(log2) {
		t.Errorf("failed to create node, expect %v but got %v", log1, log2)
	}
	//Must after data adn log test, because the slice has been cut off
	expect := append(args[:4], args[6:]...)
	got := append(node.cmd.Args[:4], node.cmd.Args[6:]...)
	if !reflect.DeepEqual(expect, got) {
		t.Errorf("failed to create node, expect %v but got %v", expect, got)
	}

}
