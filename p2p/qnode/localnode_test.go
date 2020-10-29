/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package qnode

import (
	"github.com/Qitmeer/qitmeer/crypto/ecc/secp256k1"
	"math/rand"
	"net"
	"testing"

	"github.com/Qitmeer/qitmeer/p2p/qnr"
	"github.com/stretchr/testify/assert"
)

func newLocalNodeForTesting() (*LocalNode, *DB) {
	db, _ := OpenDB("")
	key, _ := secp256k1.GeneratePrivateKey()
	return NewLocalNode(db, key.ToECDSA()), db
}

func TestLocalNode(t *testing.T) {
	ln, db := newLocalNodeForTesting()
	defer db.Close()

	if ln.Node().ID() != ln.ID() {
		t.Fatal("inconsistent ID")
	}

	ln.Set(qnr.WithEntry("x", uint(3)))
	var x uint
	if err := ln.Node().Load(qnr.WithEntry("x", &x)); err != nil {
		t.Fatal("can't load entry 'x':", err)
	} else if x != 3 {
		t.Fatal("wrong value for entry 'x':", x)
	}
}

func TestLocalNodeSeqPersist(t *testing.T) {
	ln, db := newLocalNodeForTesting()
	defer db.Close()

	if s := ln.Node().Seq(); s != 1 {
		t.Fatalf("wrong initial seq %d, want 1", s)
	}
	ln.Set(qnr.WithEntry("x", uint(1)))
	if s := ln.Node().Seq(); s != 2 {
		t.Fatalf("wrong seq %d after set, want 2", s)
	}

	// Create a new instance, it should reload the sequence number.
	// The number increases just after that because a new record is
	// created without the "x" entry.
	ln2 := NewLocalNode(db, ln.key)
	if s := ln2.Node().Seq(); s != 3 {
		t.Fatalf("wrong seq %d on new instance, want 3", s)
	}

	// Create a new instance with a different node key on the same database.
	// This should reset the sequence number.
	key, _ := secp256k1.GeneratePrivateKey()
	ln3 := NewLocalNode(db, key.ToECDSA())
	if s := ln3.Node().Seq(); s != 1 {
		t.Fatalf("wrong seq %d on instance with changed key, want 1", s)
	}
}

// This test checks behavior of the endpoint predictor.
func TestLocalNodeEndpoint(t *testing.T) {
	var (
		fallback  = &net.UDPAddr{IP: net.IP{127, 0, 0, 1}, Port: 80}
		predicted = &net.UDPAddr{IP: net.IP{127, 0, 1, 2}, Port: 81}
		staticIP  = net.IP{127, 0, 1, 2}
	)
	ln, db := newLocalNodeForTesting()
	defer db.Close()

	// Nothing is set initially.
	assert.Equal(t, net.IP(nil), ln.Node().IP())
	assert.Equal(t, 0, ln.Node().UDP())
	assert.Equal(t, uint64(1), ln.Node().Seq())

	// Set up fallback address.
	ln.SetFallbackIP(fallback.IP)
	ln.SetFallbackUDP(fallback.Port)
	assert.Equal(t, fallback.IP, ln.Node().IP())
	assert.Equal(t, fallback.Port, ln.Node().UDP())
	assert.Equal(t, uint64(2), ln.Node().Seq())

	// Add endpoint statements from random hosts.
	for i := 0; i < iptrackMinStatements; i++ {
		assert.Equal(t, fallback.IP, ln.Node().IP())
		assert.Equal(t, fallback.Port, ln.Node().UDP())
		assert.Equal(t, uint64(2), ln.Node().Seq())

		from := &net.UDPAddr{IP: make(net.IP, 4), Port: 90}
		rand.Read(from.IP)
		ln.UDPEndpointStatement(from, predicted)
	}
	assert.Equal(t, predicted.IP, ln.Node().IP())
	assert.Equal(t, predicted.Port, ln.Node().UDP())
	assert.Equal(t, uint64(3), ln.Node().Seq())

	// Static IP overrides prediction.
	ln.SetStaticIP(staticIP)
	assert.Equal(t, staticIP, ln.Node().IP())
	assert.Equal(t, fallback.Port, ln.Node().UDP())
	assert.Equal(t, uint64(4), ln.Node().Seq())
}
