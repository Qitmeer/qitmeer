/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package discover

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/Qitmeer/qitmeer/crypto/ecc/secp256k1"
	"math/rand"

	"net"
	"reflect"
	"testing"
	"testing/quick"
	"time"

	"github.com/Qitmeer/qitmeer/crypto"
	"github.com/Qitmeer/qitmeer/p2p/netutil"
	"github.com/Qitmeer/qitmeer/p2p/qnode"
	"github.com/Qitmeer/qitmeer/p2p/qnr"
)

func TestTable_pingReplace(t *testing.T) {
	run := func(newNodeResponding, lastInBucketResponding bool) {
		name := fmt.Sprintf("newNodeResponding=%t/lastInBucketResponding=%t", newNodeResponding, lastInBucketResponding)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			testPingReplace(t, newNodeResponding, lastInBucketResponding)
		})
	}

	run(true, true)
	run(false, true)
	run(true, false)
	run(false, false)
}

func testPingReplace(t *testing.T, newNodeIsResponding, lastInBucketIsResponding bool) {
	transport := newPingRecorder()
	tab, db := newTestTable(transport)
	defer db.Close()
	defer tab.close()

	<-tab.initDone

	// Fill up the sender's bucket.
	pingKey, _ := crypto.HexToECDSA("45a915e4d060149eb4365960e6a7a45f334393093061116b197e3240065ff2d8")
	pingSender := wrapNode(qnode.NewV4(&pingKey.PublicKey, net.IP{}, 99, 99))
	last := fillBucket(tab, pingSender)

	// Add the sender as if it just pinged us. Revalidate should replace the last node in
	// its bucket if it is unresponsive. Revalidate again to ensure that
	transport.dead[last.ID()] = !lastInBucketIsResponding
	transport.dead[pingSender.ID()] = !newNodeIsResponding
	tab.addSeenNode(pingSender)
	tab.doRevalidate(make(chan struct{}, 1))
	tab.doRevalidate(make(chan struct{}, 1))

	if !transport.pinged[last.ID()] {
		// Oldest node in bucket is pinged to see whether it is still alive.
		t.Error("table did not ping last node in bucket")
	}

	tab.mutex.Lock()
	defer tab.mutex.Unlock()
	wantSize := bucketSize
	if !lastInBucketIsResponding && !newNodeIsResponding {
		wantSize--
	}
	if l := len(tab.bucket(pingSender.ID()).entries); l != wantSize {
		t.Errorf("wrong bucket size after bond: got %d, want %d", l, wantSize)
	}
	if found := contains(tab.bucket(pingSender.ID()).entries, last.ID()); found != lastInBucketIsResponding {
		t.Errorf("last entry found: %t, want: %t", found, lastInBucketIsResponding)
	}
	wantNewEntry := newNodeIsResponding && !lastInBucketIsResponding
	if found := contains(tab.bucket(pingSender.ID()).entries, pingSender.ID()); found != wantNewEntry {
		t.Errorf("new entry found: %t, want: %t", found, wantNewEntry)
	}
}

func TestBucket_bumpNoDuplicates(t *testing.T) {
	t.Parallel()
	cfg := &quick.Config{
		MaxCount: 1000,
		Rand:     rand.New(rand.NewSource(time.Now().Unix())),
		Values: func(args []reflect.Value, rand *rand.Rand) {
			// generate a random list of nodes. this will be the content of the bucket.
			n := rand.Intn(bucketSize-1) + 1
			nodes := make([]*node, n)
			for i := range nodes {
				nodes[i] = nodeAtDistance(qnode.ID{}, 200, intIP(200))
			}
			args[0] = reflect.ValueOf(nodes)
			// generate random bump positions.
			bumps := make([]int, rand.Intn(100))
			for i := range bumps {
				bumps[i] = rand.Intn(len(nodes))
			}
			args[1] = reflect.ValueOf(bumps)
		},
	}

	prop := func(nodes []*node, bumps []int) (ok bool) {
		tab, db := newTestTable(newPingRecorder())
		defer db.Close()
		defer tab.close()

		b := &bucket{entries: make([]*node, len(nodes))}
		copy(b.entries, nodes)
		for i, pos := range bumps {
			tab.bumpInBucket(b, b.entries[pos])
			if hasDuplicates(b.entries) {
				t.Logf("bucket has duplicates after %d/%d bumps:", i+1, len(bumps))
				for _, n := range b.entries {
					t.Logf("  %p", n)
				}
				return false
			}
		}
		checkIPLimitInvariant(t, tab)
		return true
	}
	if err := quick.Check(prop, cfg); err != nil {
		t.Error(err)
	}
}

// This checks that the table-wide IP limit is applied correctly.
func TestTable_IPLimit(t *testing.T) {
	transport := newPingRecorder()
	tab, db := newTestTable(transport)
	defer db.Close()
	defer tab.close()

	for i := 0; i < tableIPLimit+1; i++ {
		n := nodeAtDistance(tab.self().ID(), i, net.IP{172, 0, 1, byte(i)})
		tab.addSeenNode(n)
	}
	if tab.len() > tableIPLimit {
		t.Errorf("too many nodes in table")
	}
	checkIPLimitInvariant(t, tab)
}

// This checks that the per-bucket IP limit is applied correctly.
func TestTable_BucketIPLimit(t *testing.T) {
	transport := newPingRecorder()
	tab, db := newTestTable(transport)
	defer db.Close()
	defer tab.close()

	d := 3
	for i := 0; i < bucketIPLimit+1; i++ {
		n := nodeAtDistance(tab.self().ID(), d, net.IP{172, 0, 1, byte(i)})
		tab.addSeenNode(n)
	}
	if tab.len() > bucketIPLimit {
		t.Errorf("too many nodes in table")
	}
	checkIPLimitInvariant(t, tab)
}

// checkIPLimitInvariant checks that ip limit sets contain an entry for every
// node in the table and no extra entries.
func checkIPLimitInvariant(t *testing.T, tab *Table) {
	t.Helper()

	tabset := netutil.DistinctNetSet{Subnet: tableSubnet, Limit: tableIPLimit}
	for _, b := range tab.buckets {
		for _, n := range b.entries {
			tabset.Add(n.IP())
		}
	}
	if tabset.String() != tab.ips.String() {
		t.Errorf("table IP set is incorrect:\nhave: %v\nwant: %v", tab.ips, tabset)
	}
}

func TestTable_closest(t *testing.T) {
	t.Parallel()

	test := func(test *closeTest) bool {
		// for any node table, Target and N
		transport := newPingRecorder()
		tab, db := newTestTable(transport)
		defer db.Close()
		defer tab.close()
		fillTable(tab, test.All)

		// check that closest(Target, N) returns nodes
		result := tab.closest(test.Target, test.N, false).entries
		if hasDuplicates(result) {
			t.Errorf("result contains duplicates")
			return false
		}
		if !sortedByDistanceTo(test.Target, result) {
			t.Errorf("result is not sorted by distance to target")
			return false
		}

		// check that the number of results is min(N, tablen)
		wantN := test.N
		if tlen := tab.len(); tlen < test.N {
			wantN = tlen
		}
		if len(result) != wantN {
			t.Errorf("wrong number of nodes: got %d, want %d", len(result), wantN)
			return false
		} else if len(result) == 0 {
			return true // no need to check distance
		}

		// check that the result nodes have minimum distance to target.
		for _, b := range tab.buckets {
			for _, n := range b.entries {
				if contains(result, n.ID()) {
					continue // don't run the check below for nodes in result
				}
				farthestResult := result[len(result)-1].ID()
				if qnode.DistCmp(test.Target, n.ID(), farthestResult) < 0 {
					t.Errorf("table contains node that is closer to target but it's not in result")
					t.Logf("  Target:          %v", test.Target)
					t.Logf("  Farthest Result: %v", farthestResult)
					t.Logf("  ID:              %v", n.ID())
					return false
				}
			}
		}
		return true
	}
	if err := quick.Check(test, quickcfg()); err != nil {
		t.Error(err)
	}
}

func TestTable_ReadRandomNodesGetAll(t *testing.T) {
	cfg := &quick.Config{
		MaxCount: 200,
		Rand:     rand.New(rand.NewSource(time.Now().Unix())),
		Values: func(args []reflect.Value, rand *rand.Rand) {
			args[0] = reflect.ValueOf(make([]*qnode.Node, rand.Intn(1000)))
		},
	}
	test := func(buf []*qnode.Node) bool {
		transport := newPingRecorder()
		tab, db := newTestTable(transport)
		defer db.Close()
		defer tab.close()
		<-tab.initDone

		for i := 0; i < len(buf); i++ {
			ld := cfg.Rand.Intn(len(tab.buckets))
			fillTable(tab, []*node{nodeAtDistance(tab.self().ID(), ld, intIP(ld))})
		}
		gotN := tab.ReadRandomNodes(buf)
		if gotN != tab.len() {
			t.Errorf("wrong number of nodes, got %d, want %d", gotN, tab.len())
			return false
		}
		if hasDuplicates(wrapNodes(buf[:gotN])) {
			t.Errorf("result contains duplicates")
			return false
		}
		return true
	}
	if err := quick.Check(test, cfg); err != nil {
		t.Error(err)
	}
}

type closeTest struct {
	Self   qnode.ID
	Target qnode.ID
	All    []*node
	N      int
}

func (*closeTest) Generate(rand *rand.Rand, size int) reflect.Value {
	t := &closeTest{
		Self:   gen(qnode.ID{}, rand).(qnode.ID),
		Target: gen(qnode.ID{}, rand).(qnode.ID),
		N:      rand.Intn(bucketSize),
	}
	for _, id := range gen([]qnode.ID{}, rand).([]qnode.ID) {
		r := new(qnr.Record)
		r.Set(qnr.IP(genIP(rand)))
		n := wrapNode(qnode.SignNull(r, id))
		n.livenessChecks = 1
		t.All = append(t.All, n)
	}
	return reflect.ValueOf(t)
}

func TestTable_addVerifiedNode(t *testing.T) {
	tab, db := newTestTable(newPingRecorder())
	<-tab.initDone
	defer db.Close()
	defer tab.close()

	// Insert two nodes.
	n1 := nodeAtDistance(tab.self().ID(), 256, net.IP{88, 77, 66, 1})
	n2 := nodeAtDistance(tab.self().ID(), 256, net.IP{88, 77, 66, 2})
	tab.addSeenNode(n1)
	tab.addSeenNode(n2)

	// Verify bucket content:
	bcontent := []*node{n1, n2}
	if !reflect.DeepEqual(tab.bucket(n1.ID()).entries, bcontent) {
		t.Fatalf("wrong bucket content: %v", tab.bucket(n1.ID()).entries)
	}

	// Add a changed version of n2.
	newrec := n2.Record()
	newrec.Set(qnr.IP{99, 99, 99, 99})
	newn2 := wrapNode(qnode.SignNull(newrec, n2.ID()))
	tab.addVerifiedNode(newn2)

	// Check that bucket is updated correctly.
	newBcontent := []*node{newn2, n1}
	if !reflect.DeepEqual(tab.bucket(n1.ID()).entries, newBcontent) {
		t.Fatalf("wrong bucket content after update: %v", tab.bucket(n1.ID()).entries)
	}
	checkIPLimitInvariant(t, tab)
}

func TestTable_addSeenNode(t *testing.T) {
	tab, db := newTestTable(newPingRecorder())
	<-tab.initDone
	defer db.Close()
	defer tab.close()

	// Insert two nodes.
	n1 := nodeAtDistance(tab.self().ID(), 256, net.IP{88, 77, 66, 1})
	n2 := nodeAtDistance(tab.self().ID(), 256, net.IP{88, 77, 66, 2})
	tab.addSeenNode(n1)
	tab.addSeenNode(n2)

	// Verify bucket content:
	bcontent := []*node{n1, n2}
	if !reflect.DeepEqual(tab.bucket(n1.ID()).entries, bcontent) {
		t.Fatalf("wrong bucket content: %v", tab.bucket(n1.ID()).entries)
	}

	// Add a changed version of n2.
	newrec := n2.Record()
	newrec.Set(qnr.IP{99, 99, 99, 99})
	newn2 := wrapNode(qnode.SignNull(newrec, n2.ID()))
	tab.addSeenNode(newn2)

	// Check that bucket content is unchanged.
	if !reflect.DeepEqual(tab.bucket(n1.ID()).entries, bcontent) {
		t.Fatalf("wrong bucket content after update: %v", tab.bucket(n1.ID()).entries)
	}
	checkIPLimitInvariant(t, tab)
}

// This test checks that QNR updates happen during revalidation. If a node in the table
// announces a new sequence number, the new record should be pulled.
func TestTable_revalidateSyncRecord(t *testing.T) {
	transport := newPingRecorder()
	tab, db := newTestTable(transport)
	<-tab.initDone
	defer db.Close()
	defer tab.close()

	// Insert a node.
	var r qnr.Record
	r.Set(qnr.IP(net.IP{127, 0, 0, 1}))
	id := qnode.ID{1}
	n1 := wrapNode(qnode.SignNull(&r, id))
	tab.addSeenNode(n1)

	// Update the node record.
	r.Set(qnr.WithEntry("foo", "bar"))
	n2 := qnode.SignNull(&r, id)
	transport.updateRecord(n2)

	tab.doRevalidate(make(chan struct{}, 1))
	intable := tab.getNode(id)
	if !reflect.DeepEqual(intable, n2) {
		t.Fatalf("table contains old record with seq %d, want seq %d", intable.Seq(), n2.Seq())
	}
}

// gen wraps quick.Value so it's easier to use.
// it generates a random value of the given value's type.
func gen(typ interface{}, rand *rand.Rand) interface{} {
	v, ok := quick.Value(reflect.TypeOf(typ), rand)
	if !ok {
		panic(fmt.Sprintf("couldn't generate random value of type %T", typ))
	}
	return v.Interface()
}

func genIP(rand *rand.Rand) net.IP {
	ip := make(net.IP, 4)
	rand.Read(ip)
	return ip
}

func quickcfg() *quick.Config {
	return &quick.Config{
		MaxCount: 5000,
		Rand:     rand.New(rand.NewSource(time.Now().Unix())),
	}
}

func newkey() *ecdsa.PrivateKey {
	key, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		panic("couldn't generate key: " + err.Error())
	}
	return key.ToECDSA()
}
