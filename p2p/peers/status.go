/*
 * Copyright (c) 2020.
 * Project:qitmeer
 * File:status.go
 * Date:7/2/20 8:14 PM
 * Author:Jin
 * Email:lochjin@gmail.com
 */

package peers

import (
	"errors"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"sync"
	"time"
)

// PeerConnectionState is the state of the connection.
type PeerConnectionState int32

const (
	// PeerDisconnected means there is no connection to the peer.
	PeerDisconnected PeerConnectionState = iota
	// PeerDisconnecting means there is an on-going attempt to disconnect from the peer.
	PeerDisconnecting
	// PeerConnected means the peer has an active connection.
	PeerConnected
	// PeerConnecting means there is an on-going attempt to connect to the peer.
	PeerConnecting
)

var (
	// ErrPeerUnknown is returned when there is an attempt to obtain data from a peer that is not known.
	ErrPeerUnknown = errors.New("peer unknown")
)

// Status is the structure holding the peer status information.
type Status struct {
	lock            sync.RWMutex
	maxBadResponses int
	status          map[peer.ID]*peerStatus
}

// peerStatus is the status of an individual peer at the protocol level.
type peerStatus struct {
	address               ma.Multiaddr
	direction             network.Direction
	peerState             PeerConnectionState
	chainStateLastUpdated time.Time
	badResponses          int
}
