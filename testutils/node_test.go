package testutils

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"testing"
	"time"
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

func TestNodeStartStop(t *testing.T) {
	found, err := exec.LookPath("qitmeer")
	if err != nil {
		t.Skip(fmt.Sprintf("skip the test since: %v", err))
	} else {
		t.Logf("found qitmeer execuable at %v", found)
	}
	testDir, _ := ioutil.TempDir("", "test")
	defer os.RemoveAll(testDir)
	c := newNodeConfig(testDir, []string{"--privnet"})
	n, err := newNode(t, c)
	if err != nil {
		t.Errorf("new node failed :%v", err)
	}
	err = n.start()
	if err != nil {
		t.Errorf("new node start failed :%v", err)
	}
	time.Sleep(200 * time.Millisecond)
	err = n.stop()
	if err != nil {
		t.Errorf("new node stop failed :%v", err)
	}

}
