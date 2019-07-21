// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2016-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package peer

import (
	"github.com/HalalChain/qitmeer-lib/core/message"
)

// MessageListeners defines callback function pointers to invoke with message
// listeners for a peer. Any listener which is not set to a concrete callback
// during peer initialization is ignored. Execution of multiple message
// listeners occurs serially, so one callback blocks the execution of the next.
//
// NOTE: Unless otherwise documented, these listeners must NOT directly call any
// blocking calls (such as WaitForShutdown) on the peer instance since the input
// handler goroutine blocks until the callback has completed.  Doing so will
// result in a deadlock.
type MessageListeners struct {
	// OnGetAddr is invoked when a peer receives a getaddr wire message.
	OnGetAddr func(p *Peer, msg *message.MsgGetAddr)

	// OnAddr is invoked when a peer receives an addr wire message.
	OnAddr func(p *Peer, msg *message.MsgAddr)

	// OnPing is invoked when a peer receives a ping wire message.
	OnPing func(p *Peer, msg *message.MsgPing)

	// OnPong is invoked when a peer receives a pong wire message.
	OnPong func(p *Peer, msg *message.MsgPong)

	// OnVersion is invoked when a peer receives a version wire message.
	// The caller may return a reject message in which case the message will
	// be sent to the peer and the peer will be disconnected.
	OnVersion func(p *Peer, msg *message.MsgVersion) *message.MsgReject

	// OnVerAck is invoked when a peer receives a verack wire message.
	OnVerAck func(p *Peer, msg *message.MsgVerAck)

	// OnReject is invoked when a peer receives a reject wire message.
	OnReject func(p *Peer, msg *message.MsgReject)

	// OnRead is invoked when a peer receives a wire message.  It consists
	// of the number of bytes read, the message, and whether or not an error
	// in the read occurred.  Typically, callers will opt to use the
	// callbacks for the specific message types, however this can be useful
	// for circumstances such as keeping track of server-wide byte counts or
	// working with custom message types for which the peer does not
	// directly provide a callback.
	OnRead func(p *Peer, bytesRead int, msg message.Message, err error)

	// OnWrite is invoked when we write a wire message to a peer.  It
	// consists of the number of bytes written, the message, and whether or
	// not an error in the write occurred.  This can be useful for
	// circumstances such as keeping track of server-wide byte counts.
	OnWrite func(p *Peer, bytesWritten int, msg message.Message, err error)

	// OnTx is invoked when a peer receives a tx wire message.
	OnTx func(p *Peer, msg *message.MsgTx)

	// OnGetBlocks is invoked when a peer receives a getblocks wire message.
	OnGetBlocks func(p *Peer, msg *message.MsgGetBlocks)

	// OnBlock is invoked when a peer receives a block wire message.
	OnBlock func(p *Peer, msg *message.MsgBlock, buf []byte)

	// OnInv is invoked when a peer receives an inv message.
	OnInv func(p *Peer, msg *message.MsgInv)

	// OnGetData is invoked when a peer receives a getdata wire message.
	OnGetData func(p *Peer, msg *message.MsgGetData)

	// OnNotFound is invoked when a peer receives a notfound message.
	OnNotFound func(p *Peer, msg *message.MsgNotFound)

	// OnGetMiningState is invoked when a peer receives a getminings wire
	// message.
	OnGetMiningState func(p *Peer, msg *message.MsgGetMiningState)

	// OnMiningState is invoked when a peer receives a miningstate wire
	// message.
	OnMiningState func(p *Peer, msg *message.MsgMiningState)

	/*
	// OnSendHeaders is invoked when a peer receives a sendheaders message.
	OnSendHeaders func(p *Peer, msg *message.MsgSendHeaders)

	// OnMemPool is invoked when a peer receives a mempool wire message.
	OnMemPool func(p *Peer, msg *message.MsgMemPool)

	// OnCFilter is invoked when a peer receives a cfilter wire message.
	OnCFilter func(p *Peer, msg *message.MsgCFilter)

	// OnCFHeaders is invoked when a peer receives a cfheaders wire
	// message.
	OnCFHeaders func(p *Peer, msg *message.MsgCFHeaders)

	// OnCFTypes is invoked when a peer receives a cftypes wire message.
	OnCFTypes func(p *Peer, msg *message.MsgCFTypes)

	// OnHeaders is invoked when a peer receives a headers wire message.
	OnHeaders func(p *Peer, msg *message.MsgHeaders)


	// OnGetHeaders is invoked when a peer receives a getheaders wire
	// message.
	OnGetHeaders func(p *Peer, msg *message.MsgGetHeaders)

	// OnGetCFilter is invoked when a peer receives a getcfilter wire
	// message.
	OnGetCFilter func(p *Peer, msg *message.MsgGetCFilter)

	// OnGetCFHeaders is invoked when a peer receives a getcfheaders
	// wire message.
	OnGetCFHeaders func(p *Peer, msg *message.MsgGetCFHeaders)

	// OnGetCFTypes is invoked when a peer receives a getcftypes wire
	// message.
	OnGetCFTypes func(p *Peer, msg *message.MsgGetCFTypes)

	// OnFeeFilter is invoked when a peer receives a feefilter wire message.
	OnFeeFilter func(p *Peer, msg *message.MsgFeeFilter)
	*/
}

