// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2016-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package peer

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/roughtime"
	"github.com/Qitmeer/qitmeer/core/message"
	"github.com/Qitmeer/qitmeer/log"
	"io"
	"net"
	"sync/atomic"
	"time"
)

// inHandler handles all incoming messages for the peer.  It must be run as a
// goroutine.
func (p *Peer) inHandler() {
	// Peers must complete the initial version negotiation within a shorter
	// timeframe than a general idle timeout.  The timer is then reset below
	// to idleTimeout for all future messages.
	idleTimer := time.AfterFunc(idleTimeout, func() {
		log.Warn(fmt.Sprintf("Peer no answer for %s -- disconnecting", idleTimeout), "peer", p)
		p.Disconnect()
	})

out:
	for atomic.LoadInt32(&p.disconnect) == 0 {
		// Read a message and stop the idle timer as soon as the read
		// is done.  The timer is reset below for the next iteration if
		// needed.
		rmsg, _, err := p.readMessage()
		idleTimer.Stop()
		if err != nil {
			// Only log the error and send reject message if the
			// local peer is not forcibly disconnecting and the
			// remote peer has not disconnected.
			if p.shouldHandleReadError(err) {
				errMsg := fmt.Sprintf("Can't read message from %s: %v", p.addr, err)
				log.Error(errMsg)

				// Push a reject message for the malformed message and wait for
				// the message to be sent before disconnecting.
				//
				// NOTE: Ideally this would include the command in the header if
				// at least that much of the message was valid, but that is not
				// currently exposed by wire, so just used malformed for the
				// command.
				p.PushRejectMsg("malformed", message.RejectMalformed, errMsg, nil,
					true)
			}
			break out
		}
		atomic.StoreInt64(&p.lastRecv, roughtime.Now().Unix())
		p.stallControl <- stallControlMsg{sccReceiveMessage, rmsg}

		// Handle each supported message type.
		p.stallControl <- stallControlMsg{sccHandlerStart, rmsg}
		// log.Trace("peer inHandler", "rmsg", rmsg, "buff",buf)
		switch msg := rmsg.(type) {
		case *message.MsgVersion:
			// Limit to one version message per peer.
			p.PushRejectMsg(msg.Command(), message.RejectDuplicate,
				"duplicate version message", nil, true)
			break out

		case *message.MsgVerAck:
			// No read lock is necessary because verAckReceived is not written
			// to in any other goroutine.
			if p.verAckReceived {
				log.Info(fmt.Sprintf("Already received 'verack' from peer %v -- "+
					"disconnecting", p))
				break out
			}
			p.flagsMtx.Lock()
			p.verAckReceived = true
			p.flagsMtx.Unlock()
			if p.cfg.Listeners.OnVerAck != nil {
				p.cfg.Listeners.OnVerAck(p, msg)
			}

		case *message.MsgTx:
			if p.cfg.Listeners.OnTx != nil {
				p.cfg.Listeners.OnTx(p, msg)
			}

		case *message.MsgGetBlocks:
			if p.cfg.Listeners.OnGetBlocks != nil {
				p.cfg.Listeners.OnGetBlocks(p, msg)
			}

		case *message.MsgGetData:
			if p.cfg.Listeners.OnGetData != nil {
				p.cfg.Listeners.OnGetData(p, msg)
			}

		case *message.MsgGetMiningState:
			if p.cfg.Listeners.OnGetMiningState != nil {
				p.cfg.Listeners.OnGetMiningState(p, msg)
			}

		case *message.MsgMiningState:
			if p.cfg.Listeners.OnMiningState != nil {
				p.cfg.Listeners.OnMiningState(p, msg)
			}

		case *message.MsgNotFound:
			if p.cfg.Listeners.OnNotFound != nil {
				p.cfg.Listeners.OnNotFound(p, msg)
			}

		case *message.MsgGraphState:
			if p.cfg.Listeners.OnGraphState != nil {
				p.cfg.Listeners.OnGraphState(p, msg)
			}

		case *message.MsgGetHeaders:
			if p.cfg.Listeners.OnGetHeaders != nil {
				p.cfg.Listeners.OnGetHeaders(p, msg)
			}

		case *message.MsgMemPool:
			if p.cfg.Listeners.OnMemPool != nil {
				p.cfg.Listeners.OnMemPool(p, msg)
			}

		case *message.MsgSyncResult:
			if p.cfg.Listeners.OnSyncResult != nil {
				p.cfg.Listeners.OnSyncResult(p, msg)
			}

		case *message.MsgSyncDAG:
			if p.cfg.Listeners.OnSyncDAG != nil {
				p.cfg.Listeners.OnSyncDAG(p, msg)
			}

		case *message.MsgSyncPoint:
			if p.cfg.Listeners.OnSyncPoint != nil {
				p.cfg.Listeners.OnSyncPoint(p, msg)
			}

		case *message.MsgFeeFilter:
			if p.cfg.Listeners.OnFeeFilter != nil {
				p.cfg.Listeners.OnFeeFilter(p, msg)
			}
		case *message.MsgReject:
			if p.cfg.Listeners.OnReject != nil {
				p.cfg.Listeners.OnReject(p, msg)
			}
		default:
			log.Debug("Received unhandled message", "command", rmsg.Command(), "peer", p)
		}
		p.stallControl <- stallControlMsg{sccHandlerDone, rmsg}

		// A message was received so reset the idle timer.
		idleTimer.Reset(idleTimeout)
	}

	// Ensure the idle timer is stopped to avoid leaking the resource.
	idleTimer.Stop()

	// Ensure connection is closed.
	p.Disconnect()

	close(p.inQuit)
	log.Trace("Peer input handler done", "peer", p.addr)
}

// shouldHandleReadError returns whether or not the passed error, which is
// expected to have come from reading from the remote peer in the inHandler,
// should be logged and responded to with a reject message.
func (p *Peer) shouldHandleReadError(err error) bool {
	// No logging or reject message when the peer is being forcibly
	// disconnected.
	if atomic.LoadInt32(&p.disconnect) != 0 {
		return false
	}

	// No logging or reject message when the remote peer has been
	// disconnected.
	if err == io.EOF {
		return false
	}
	if opErr, ok := err.(*net.OpError); ok && !opErr.Temporary() {
		return false
	}

	return true
}
