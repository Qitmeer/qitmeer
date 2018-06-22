// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2013-2015 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package types

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"time"
	s "github.com/noxproject/nox/core/serialization"
)

// ErrInvalidNetAddr describes an error that indicates the caller didn't specify
// a TCP address as required.
var ErrInvalidNetAddr = errors.New("provided net.Addr is not a net.TCPAddr")

// MaxNetAddressPayload returns the max payload size for a Decred NetAddress
// based on the protocol version.
func MaxNetAddressPayload(pver uint32) uint32 {
	// Services 8 bytes + ip 16 bytes + port 2 bytes.
	plen := uint32(26)

	// Timestamp 4 bytes.
	plen += 4

	return plen
}

// NetAddress defines information about a peer on the network including the time
// it was last seen, the services it supports, its IP address, and port.
type NetAddress struct {
	// Last time the address was seen.  This is, unfortunately, encoded as a
	// uint32 on the wire and therefore is limited to 2106.  This field is
	// not present in the Decred version message (MsgVersion) nor was it
	// added until protocol version >= NetAddressTimeVersion.
	Timestamp time.Time

	// IP address of the peer.
	IP net.IP

	// Port the peer is using.  This is encoded in big endian on the wire
	// which differs from most everything else.
	Port uint16
}


// NewNetAddressIPPort returns a new NetAddress using the provided IP, port, and
// supported services with defaults for the remaining fields.
func NewNetAddressIPPort(ip net.IP, port uint16 ) *NetAddress {
	return NewNetAddressTimestamp(time.Now(),  ip, port)
}

// NewNetAddressTimestamp returns a new NetAddress using the provided
// timestamp, IP, port, and supported services. The timestamp is rounded to
// single second precision.
func NewNetAddressTimestamp(
	timestamp time.Time, ip net.IP, port uint16) *NetAddress {
	// Limit the timestamp to one second precision since the protocol
	// doesn't support better.
	na := NetAddress{
		Timestamp: time.Unix(timestamp.Unix(), 0),
		IP:        ip,
		Port:      port,
	}
	return &na
}

// NewNetAddress returns a new NetAddress using the provided TCP address and
// supported services with defaults for the remaining fields.
//
// Note that addr must be a net.TCPAddr.  An ErrInvalidNetAddr is returned
// if it is not.
func NewNetAddress(addr net.Addr) (*NetAddress, error) {
	tcpAddr, ok := addr.(*net.TCPAddr)
	if !ok {
		return nil, ErrInvalidNetAddr
	}

	na := NewNetAddressIPPort(tcpAddr.IP, uint16(tcpAddr.Port))
	return na, nil
}

// readNetAddress reads an encoded NetAddress from r depending on the protocol
// version and whether or not the timestamp is included per ts.  Some messages
// like version do not include the timestamp.
func ReadNetAddress(r io.Reader, pver uint32, na *NetAddress, ts bool) error {
	var ip [16]byte

	// NOTE: The Decred protocol uses a uint32 for the timestamp so it will
	// stop working somewhere around 2106.  Also timestamp wasn't added until
	// protocol version >= NetAddressTimeVersion
	if ts {
		err := s.ReadElements(r, (*s.Uint32Time)(&na.Timestamp))
		if err != nil {
			return err
		}
	}

	err := s.ReadElements(r, &ip)
	if err != nil {
		return err
	}

	// TODO unify endian
	// Sigh. protocol mixes little and big endian.
	port, err := s.BinarySerializer.Uint16(r, binary.BigEndian)
	if err != nil {
		return err
	}

	*na = NetAddress{
		Timestamp: na.Timestamp,
		IP:        net.IP(ip[:]),
		Port:      port,
	}
	return nil
}

// writeNetAddress serializes a NetAddress to w depending on the protocol
// version and whether or not the timestamp is included per ts.  Some messages
// like version do not include the timestamp.
func WriteNetAddress(w io.Writer, pver uint32, na *NetAddress, ts bool) error {
	// TODO fix time ambiguous
	// NOTE: The protocol uses a uint32 for the timestamp so it will
	// stop working somewhere around 2106.  Also timestamp wasn't added until
	// until protocol version >= NetAddressTimeVersion.
	if ts {
		err := s.WriteElements(w, uint32(na.Timestamp.Unix()))
		if err != nil {
			return err
		}
	}

	// Ensure to always write 16 bytes even if the ip is nil.
	var ip [16]byte
	if na.IP != nil {
		copy(ip[:], na.IP.To16())
	}
	err := s.WriteElements(w, ip)
	if err != nil {
		return err
	}

	// TODO unify endian
	// Sigh.  protocol mixes little and big endian.
	return binary.Write(w, binary.BigEndian, na.Port)
}
