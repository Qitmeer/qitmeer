// Copyright (c) 2017-2018 The nox developers
package p2p

import "sync"

// TODO refactor the interface of the p2p layer
// example github.com/libp2p/go-libp2p-peer)

// Peer represents a connected p2p network remote node.
type Peer struct {
	statsMtx           sync.RWMutex
	lastBlock          uint64

}


//TODO last-block from peer
// LastBlock returns the last block of the peer.
//
// This function is safe for concurrent access.
func (p *Peer) LastBlock() uint64 {
	p.statsMtx.RLock()
	lastBlock := p.lastBlock
	p.statsMtx.RUnlock()

	return lastBlock
}

