// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package peerserver

import (
	"github.com/Qitmeer/qitmeer-lib/common/hash"
	"github.com/Qitmeer/qitmeer-lib/core/message"
	"github.com/Qitmeer/qitmeer-lib/core/types"
	"github.com/Qitmeer/qitmeer-lib/log"
	"github.com/Qitmeer/qitmeer/p2p/addmgr"
	"github.com/Qitmeer/qitmeer/p2p/connmgr"
	"github.com/Qitmeer/qitmeer/p2p/peer"
	"sync"
)

// serverPeer extends the peer to maintain state shared by the p2p server and
// the blockmanager.
type serverPeer struct {
	*peer.Peer

	connReq         *connmgr.ConnReq
	server          *PeerServer
	persistent      bool
	relayMtx        sync.Mutex
	disableRelayTx  bool
	isWhitelisted   bool
	requestQueue    []*message.InvVect
	requestedTxns   map[hash.Hash]struct{}
	knownAddresses  map[string]struct{}
	banScore        connmgr.DynamicBanScore
	quit            chan struct{}

	// addrsSent tracks whether or not the peer has responded to a getaddr
	// request.  It is used to prevent more than one response per connection.
	addrsSent bool

	// The following chans are used to sync blockmanager and server.
	syncPeer        *peer.ServerPeer
}



// newServerPeer returns a new serverPeer instance. The peer needs to be set by
// the caller.
func newServerPeer(s *PeerServer, isPersistent bool) *serverPeer {
	return &serverPeer{
		server:          s,
		persistent:      isPersistent,
		knownAddresses:  make(map[string]struct{}),
		quit:            make(chan struct{}),
		syncPeer : &peer.ServerPeer{
			TxProcessed:     make(chan struct{}, 1),
			BlockProcessed:  make(chan struct{}, 1),
			RequestedBlocks: make(map[hash.Hash]struct{}),
			RequestedTxns:   make(map[hash.Hash]struct{}),
	}}
}

// pushAddrMsg sends an addr message to the connected peer using the provided
// addresses.
func (sp *serverPeer) pushAddrMsg(addresses []*types.NetAddress) {
	// Filter addresses already known to the peer.
	addrs := make([]*types.NetAddress, 0, len(addresses))
	for _, addr := range addresses {
		if !sp.addressKnown(addr) {
			addrs = append(addrs, addr)
		}
	}
	known, err := sp.PushAddrMsg(addrs)
	if err != nil {
		log.Error("Can't push address message to %s: %v", sp.Peer, err)
		sp.Disconnect()
		return
	}
	sp.addKnownAddresses(known)
}

// addressKnown true if the given address is already known to the peer.
func (sp *serverPeer) addressKnown(na *types.NetAddress) bool {
	_, exists := sp.knownAddresses[addmgr.NetAddressKey(na)]
	return exists
}

// addKnownAddresses adds the given addresses to the set of known addreses to
// the peer to prevent sending duplicate addresses.
func (sp *serverPeer) addKnownAddresses(addresses []*types.NetAddress) {
	for _, na := range addresses {
		sp.knownAddresses[addmgr.NetAddressKey(na)] = struct{}{}
	}
}

// setDisableRelayTx toggles relaying of transactions for the given peer.
// It is safe for concurrent access.
func (sp *serverPeer) setDisableRelayTx(disable bool) {
	sp.relayMtx.Lock()
	sp.disableRelayTx = disable
	sp.relayMtx.Unlock()
}

// relayTxDisabled returns whether or not relaying of transactions for the given
// peer is disabled.
// It is safe for concurrent access.
func (sp *serverPeer) relayTxDisabled() bool {
	sp.relayMtx.Lock()
	isDisabled := sp.disableRelayTx
	sp.relayMtx.Unlock()

	return isDisabled
}

// IsTxRelayDisabled returns whether or not the peer has disabled transaction
// relay.
func (sp *serverPeer) IsTxRelayDisabled() bool {
	return sp.disableRelayTx
}

// BanScore returns the current integer value that represents how close the peer
// is to being banned.
func (sp *serverPeer) BanScore() uint32 {
	return sp.banScore.Int()
}