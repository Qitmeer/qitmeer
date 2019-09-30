// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package message

import (
	"bytes"
	"fmt"
	"github.com/Qitmeer/qitmeer/core/dag"
	"github.com/satori/go.uuid"
	"io"
	"net"
	"github.com/Qitmeer/qitmeer/common/hash"
	"strings"
	"time"
	s "github.com/Qitmeer/qitmeer/core/serialization"
	"github.com/Qitmeer/qitmeer/core/protocol"
	"github.com/Qitmeer/qitmeer/core/types"
)

// MaxUserAgentLen is the maximum allowed length for the user agent field in a
// version message (MsgVersion).
const MaxUserAgentLen = 256

// UUID for peer
var UUID = uuid.NewV4()

// MsgVersion implements the Message interface and represents a version message.
// It is used for a peer to advertise itself as soon as an outbound connection
// is made.  The remote peer then uses this information along with its own to
// negotiate.  The remote peer must then respond with a version message of its
// own containing the negotiated values followed by a verack message (MsgVerAck).
// This exchange must take place before any further communication is allowed
// to proceed.
type MsgVersion struct {
	// Version of the protocol the node is using.
	ProtocolVersion int32

	// Bitfield which identifies the enabled services.
	Services protocol.ServiceFlag

	// Time the message was generated.  This is encoded as an int64 on the wire.
	Timestamp time.Time

	// Address of the remote peer.
	AddrYou types.NetAddress

	// Address of the local peer.
	AddrMe types.NetAddress

	// Unique value associated with message that is used to detect self
	// connections.
	Nonce uint64

	// The user agent that generated messsage.  This is a encoded as a varString
	// on the wire.  This has a max length of MaxUserAgentLen.
	UserAgent string

	// Last DAG graph state seen by the generator of the version message.
	LastGS *dag.GraphState

	// Don't announce transactions to peer.
	DisableRelayTx bool
}

// HasService returns whether the specified service is supported by the peer
// that generated the message.
func (msg *MsgVersion) HasService(service protocol.ServiceFlag) bool {
	return msg.Services&service == service
}

// AddService adds service as a supported service by the peer generating the
// message.
func (msg *MsgVersion) AddService(service protocol.ServiceFlag) {
	msg.Services |= service
}

// Decode decodes r encoding into the receiver.
// The version message is special in that the protocol version hasn't been
// negotiated yet.  As a result, the pver field is ignored and any fields which
// are added in new versions are optional.  This also mean that r must be a
// *bytes.Buffer so the number of remaining bytes can be ascertained.
//
// This is part of the Message interface implementation.
func (msg *MsgVersion) Decode(r io.Reader, pver uint32) error {
	buf, ok := r.(*bytes.Buffer)
	if !ok {
		return fmt.Errorf("in method MsgVersion.Decode reader is not a " +
			"*bytes.Buffer")
	}

	err := s.ReadElements(buf, &msg.ProtocolVersion, &msg.Services,
		(*s.Int64Time)(&msg.Timestamp))
	if err != nil {
		return err
	}

	err = types.ReadNetAddress(buf, pver, &msg.AddrYou, false)
	if err != nil {
		return err
	}

	// Protocol versions >= 106 added a from address, nonce, and user agent
	// field and they are only considered present if there are bytes
	// remaining in the message.
	if buf.Len() > 0 {
		err = types.ReadNetAddress(buf, pver, &msg.AddrMe, false)
		if err != nil {
			return err
		}
	}
	if buf.Len() > 0 {
		err = s.ReadElements(buf, &msg.Nonce)
		if err != nil {
			return err
		}
	}
	if buf.Len() > 0 {
		userAgent, err := s.ReadVarString(buf, pver)
		if err != nil {
			return err
		}
		err = validateUserAgent(userAgent)
		if err != nil {
			return err
		}
		msg.UserAgent = userAgent
	}

	// Protocol versions >= 209 added a last known block field.  It is only
	// considered present if there are bytes remaining in the message.
	if buf.Len() > 0 {
		msg.LastGS=dag.NewGraphState()
		err=msg.LastGS.Decode(buf,pver)
		if err != nil {
			return err
		}
	}

	// There was no relay transactions field before BIP0037Version, but
	// the default behavior prior to the addition of the field was to always
	// relay transactions.
	if buf.Len() > 0 {
		// It's safe to ignore the error here since the buffer has at
		// least one byte and that byte will result in a boolean value
		// regardless of its value.  Also, the wire encoding for the
		// field is true when transactions should be relayed, so reverse
		// it for the DisableRelayTx field.
		var relayTx bool
		s.ReadElements(r, &relayTx)
		msg.DisableRelayTx = !relayTx
	}

	return nil
}

// Encode encodes the receiver to w
// This is part of the Message interface implementation.
func (msg *MsgVersion) Encode(w io.Writer, pver uint32) error {
	err := validateUserAgent(msg.UserAgent)
	if err != nil {
		return err
	}

	err = s.WriteElements(w, msg.ProtocolVersion,msg.Services,
		msg.Timestamp.Unix())
	if err != nil {
		return err
	}

	err = types.WriteNetAddress(w, pver, &msg.AddrYou, false)
	if err != nil {
		return err
	}

	err = types.WriteNetAddress(w, pver, &msg.AddrMe, false)
	if err != nil {
		return err
	}

	err = s.WriteElements(w, msg.Nonce)
	if err != nil {
		return err
	}

	err = s.WriteVarString(w, pver, msg.UserAgent)
	if err != nil {
		return err
	}
	err = msg.LastGS.Encode(w,pver)
	if err != nil {
		return err
	}

	return s.WriteElements(w, !msg.DisableRelayTx)
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgVersion) Command() string {
	return CmdVersion
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgVersion) MaxPayloadLength(pver uint32) uint32 {
	// XXX: <= 106 different

	// Protocol version 4 bytes + services 8 bytes + timestamp 8 bytes +
	// remote and local net addresses + nonce 8 bytes + length of user
	// agent (varInt) + max allowed useragent length + last block 4 bytes +
	// relay transactions flag 1 byte.
	return 29 + (types.MaxNetAddressPayload(pver) * 2) + s.MaxVarIntPayload +
		MaxUserAgentLen+8 + 4 + (dag.MaxTips * hash.HashSize)
}

// NewMsgVersion returns a new Version message that conforms to the Message
// interface using the passed parameters and defaults for the remaining
// fields.
func NewMsgVersion(me *types.NetAddress, you *types.NetAddress, nonce uint64,
	lastGS *dag.GraphState) *MsgVersion {

	// Limit the timestamp to one second precision since the protocol
	// doesn't support better.
	return &MsgVersion{
		ProtocolVersion: int32(protocol.ProtocolVersion),
		Timestamp:       time.Unix(time.Now().Unix(), 0),
		AddrYou:         *you,
		AddrMe:          *me,
		Nonce:           nonce,
		UserAgent:       UUID.String(),
		LastGS:          lastGS,
		DisableRelayTx:  false,
	}
}

// NewMsgVersionFromConn is a convenience function that extracts the remote
// and local address from conn and returns a new version message that
// conforms to the Message interface.  See NewMsgVersion.
func NewMsgVersionFromConn(conn net.Conn, nonce uint64,
	lastGS *dag.GraphState) (*MsgVersion, error) {

	// TODO, should define unknown flag instead of using hard-coding 0
	// Don't assume any services until we know otherwise.
	lna, err := types.NewNetAddress(conn.LocalAddr(),0)
	if err != nil {
		return nil, err
	}

	// Don't assume any services until we know otherwise.
	rna, err := types.NewNetAddress(conn.RemoteAddr(),0)
	if err != nil {
		return nil, err
	}

	return NewMsgVersion(lna, rna, nonce, lastGS), nil
}

// validateUserAgent checks userAgent length against MaxUserAgentLen
func validateUserAgent(userAgent string) error {
	if len(userAgent) > MaxUserAgentLen {
		str := fmt.Sprintf("user agent too long [len %v, max %v]",
			len(userAgent), MaxUserAgentLen)
		return messageError("MsgVersion", str)
	}
	return nil
}

// AddUserAgent adds a user agent to the user agent string for the version
// message.  The version string is not defined to any strict format, although
// it is recommended to use the form "major.minor.revision" e.g. "2.6.41".
func (msg *MsgVersion) AddUserAgent(name string, version string,
	comments ...string) error {

	newUserAgent := fmt.Sprintf("%s:%s", name, version)
	if len(comments) != 0 {
		newUserAgent = fmt.Sprintf("%s(%s)", newUserAgent,
			strings.Join(comments, "; "))
	}
	newUserAgent = fmt.Sprintf("%s@%s", msg.UserAgent, newUserAgent)
	err := validateUserAgent(newUserAgent)
	if err != nil {
		return err
	}
	msg.UserAgent = newUserAgent
	return nil
}
