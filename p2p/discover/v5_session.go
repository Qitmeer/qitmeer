/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package discover

import (
	crand "crypto/rand"

	"github.com/Qitmeer/qitmeer/common/mclock"
	"github.com/Qitmeer/qitmeer/p2p/qnode"
	"github.com/hashicorp/golang-lru/simplelru"
)

// The sessionCache keeps negotiated encryption keys and
// state for in-progress handshakes in the Discovery v5 wire protocol.
type sessionCache struct {
	sessions   *simplelru.LRU
	handshakes map[sessionID]*whoareyouV5
	clock      mclock.Clock
}

// sessionID identifies a session or handshake.
type sessionID struct {
	id   qnode.ID
	addr string
}

// session contains session information
type session struct {
	writeKey     []byte
	readKey      []byte
	nonceCounter uint32
}

func newSessionCache(maxItems int, clock mclock.Clock) *sessionCache {
	cache, err := simplelru.NewLRU(maxItems, nil)
	if err != nil {
		panic("can't create session cache")
	}
	return &sessionCache{
		sessions:   cache,
		handshakes: make(map[sessionID]*whoareyouV5),
		clock:      clock,
	}
}

// nextNonce creates a nonce for encrypting a message to the given session.
func (sc *sessionCache) nextNonce(id qnode.ID, addr string) []byte {
	n := make([]byte, gcmNonceSize)
	crand.Read(n)
	return n
}

// session returns the current session for the given node, if any.
func (sc *sessionCache) session(id qnode.ID, addr string) *session {
	item, ok := sc.sessions.Get(sessionID{id, addr})
	if !ok {
		return nil
	}
	return item.(*session)
}

// readKey returns the current read key for the given node.
func (sc *sessionCache) readKey(id qnode.ID, addr string) []byte {
	if s := sc.session(id, addr); s != nil {
		return s.readKey
	}
	return nil
}

// writeKey returns the current read key for the given node.
func (sc *sessionCache) writeKey(id qnode.ID, addr string) []byte {
	if s := sc.session(id, addr); s != nil {
		return s.writeKey
	}
	return nil
}

// storeNewSession stores new encryption keys in the cache.
func (sc *sessionCache) storeNewSession(id qnode.ID, addr string, r, w []byte) {
	sc.sessions.Add(sessionID{id, addr}, &session{
		readKey: r, writeKey: w,
	})
}

// getHandshake gets the handshake challenge we previously sent to the given remote node.
func (sc *sessionCache) getHandshake(id qnode.ID, addr string) *whoareyouV5 {
	return sc.handshakes[sessionID{id, addr}]
}

// storeSentHandshake stores the handshake challenge sent to the given remote node.
func (sc *sessionCache) storeSentHandshake(id qnode.ID, addr string, challenge *whoareyouV5) {
	challenge.sent = sc.clock.Now()
	sc.handshakes[sessionID{id, addr}] = challenge
}

// deleteHandshake deletes handshake data for the given node.
func (sc *sessionCache) deleteHandshake(id qnode.ID, addr string) {
	delete(sc.handshakes, sessionID{id, addr})
}

// handshakeGC deletes timed-out handshakes.
func (sc *sessionCache) handshakeGC() {
	deadline := sc.clock.Now().Add(-handshakeTimeout)
	for key, challenge := range sc.handshakes {
		if challenge.sent < deadline {
			delete(sc.handshakes, key)
		}
	}
}
