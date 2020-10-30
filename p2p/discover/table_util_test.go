/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package discover

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"sort"
	"sync"

	"github.com/Qitmeer/qitmeer/crypto"
	"github.com/Qitmeer/qitmeer/p2p/qnode"
	"github.com/Qitmeer/qitmeer/p2p/qnr"
)

var nullNode *qnode.Node

func init() {
	var r qnr.Record
	r.Set(qnr.IP{0, 0, 0, 0})
	nullNode = qnode.SignNull(&r, qnode.ID{})
}

func newTestTable(t transport) (*Table, *qnode.DB) {
	db, _ := qnode.OpenDB("")
	tab, _ := newTable(t, db, nil)
	go tab.loop()
	return tab, db
}

// nodeAtDistance creates a node for which enode.LogDist(base, n.id) == ld.
func nodeAtDistance(base qnode.ID, ld int, ip net.IP) *node {
	var r qnr.Record
	r.Set(qnr.IP(ip))
	return wrapNode(qnode.SignNull(&r, idAtDistance(base, ld)))
}

// nodesAtDistance creates n nodes for which enode.LogDist(base, node.ID()) == ld.
func nodesAtDistance(base qnode.ID, ld int, n int) []*qnode.Node {
	results := make([]*qnode.Node, n)
	for i := range results {
		results[i] = unwrapNode(nodeAtDistance(base, ld, intIP(i)))
	}
	return results
}

func nodesToRecords(nodes []*qnode.Node) []*qnr.Record {
	records := make([]*qnr.Record, len(nodes))
	for i := range nodes {
		records[i] = nodes[i].Record()
	}
	return records
}

// idAtDistance returns a random hash such that enode.LogDist(a, b) == n
func idAtDistance(a qnode.ID, n int) (b qnode.ID) {
	if n == 0 {
		return a
	}
	// flip bit at position n, fill the rest with random bits
	b = a
	pos := len(a) - n/8 - 1
	bit := byte(0x01) << (byte(n%8) - 1)
	if bit == 0 {
		pos++
		bit = 0x80
	}
	b[pos] = a[pos]&^bit | ^a[pos]&bit // TODO: randomize end bits
	for i := pos + 1; i < len(a); i++ {
		b[i] = byte(rand.Intn(255))
	}
	return b
}

func intIP(i int) net.IP {
	return net.IP{byte(i), 0, 2, byte(i)}
}

// fillBucket inserts nodes into the given bucket until it is full.
func fillBucket(tab *Table, n *node) (last *node) {
	ld := qnode.LogDist(tab.self().ID(), n.ID())
	b := tab.bucket(n.ID())
	for len(b.entries) < bucketSize {
		b.entries = append(b.entries, nodeAtDistance(tab.self().ID(), ld, intIP(ld)))
	}
	return b.entries[bucketSize-1]
}

// fillTable adds nodes the table to the end of their corresponding bucket
// if the bucket is not full. The caller must not hold tab.mutex.
func fillTable(tab *Table, nodes []*node) {
	for _, n := range nodes {
		tab.addSeenNode(n)
	}
}

type pingRecorder struct {
	mu           sync.Mutex
	dead, pinged map[qnode.ID]bool
	records      map[qnode.ID]*qnode.Node
	n            *qnode.Node
}

func newPingRecorder() *pingRecorder {
	var r qnr.Record
	r.Set(qnr.IP{0, 0, 0, 0})
	n := qnode.SignNull(&r, qnode.ID{})

	return &pingRecorder{
		dead:    make(map[qnode.ID]bool),
		pinged:  make(map[qnode.ID]bool),
		records: make(map[qnode.ID]*qnode.Node),
		n:       n,
	}
}

// setRecord updates a node record. Future calls to ping and
// requestQNR will return this record.
func (t *pingRecorder) updateRecord(n *qnode.Node) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.records[n.ID()] = n
}

// Stubs to satisfy the transport interface.
func (t *pingRecorder) Self() *qnode.Node           { return nullNode }
func (t *pingRecorder) lookupSelf() []*qnode.Node   { return nil }
func (t *pingRecorder) lookupRandom() []*qnode.Node { return nil }
func (t *pingRecorder) close()                      {}

// ping simulates a ping request.
func (t *pingRecorder) ping(n *qnode.Node) (seq uint64, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.pinged[n.ID()] = true
	if t.dead[n.ID()] {
		return 0, errTimeout
	}
	if t.records[n.ID()] != nil {
		seq = t.records[n.ID()].Seq()
	}
	return seq, nil
}

// requestQNR simulates an QNR request.
func (t *pingRecorder) RequestQNR(n *qnode.Node) (*qnode.Node, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.dead[n.ID()] || t.records[n.ID()] == nil {
		return nil, errTimeout
	}
	return t.records[n.ID()], nil
}

func hasDuplicates(slice []*node) bool {
	seen := make(map[qnode.ID]bool)
	for i, e := range slice {
		if e == nil {
			panic(fmt.Sprintf("nil *Node at %d", i))
		}
		if seen[e.ID()] {
			return true
		}
		seen[e.ID()] = true
	}
	return false
}

func checkNodesEqual(got, want []*qnode.Node) error {
	if len(got) == len(want) {
		for i := range got {
			if !nodeEqual(got[i], want[i]) {
				goto NotEqual
			}
			return nil
		}
	}

NotEqual:
	output := new(bytes.Buffer)
	fmt.Fprintf(output, "got %d nodes:\n", len(got))
	for _, n := range got {
		fmt.Fprintf(output, "  %v %v\n", n.ID(), n)
	}
	fmt.Fprintf(output, "want %d:\n", len(want))
	for _, n := range want {
		fmt.Fprintf(output, "  %v %v\n", n.ID(), n)
	}
	return errors.New(output.String())
}

func nodeEqual(n1 *qnode.Node, n2 *qnode.Node) bool {
	return n1.ID() == n2.ID() && n1.IP().Equal(n2.IP())
}

func sortByID(nodes []*qnode.Node) {
	sort.Slice(nodes, func(i, j int) bool {
		return string(nodes[i].ID().Bytes()) < string(nodes[j].ID().Bytes())
	})
}

func sortedByDistanceTo(distbase qnode.ID, slice []*node) bool {
	return sort.SliceIsSorted(slice, func(i, j int) bool {
		return qnode.DistCmp(distbase, slice[i].ID(), slice[j].ID()) < 0
	})
}

func hexEncPrivkey(h string) *ecdsa.PrivateKey {
	b, err := hex.DecodeString(h)
	if err != nil {
		panic(err)
	}
	key, err := crypto.ToECDSA(b)
	if err != nil {
		panic(err)
	}
	return key
}

func hexEncPubkey(h string) (ret encPubkey) {
	b, err := hex.DecodeString(h)
	if err != nil {
		panic(err)
	}
	if len(b) != len(ret) {
		panic("invalid length")
	}
	copy(ret[:], b)
	return ret
}
