// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2015 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package types

import (
	"encoding/binary"
	"errors"
	"github.com/Qitmeer/qitmeer/common/network"
	"github.com/Qitmeer/qitmeer/common/roughtime"
	"github.com/Qitmeer/qitmeer/core/protocol"
	s "github.com/Qitmeer/qitmeer/core/serialization"
	"io"
	"net"
	"strconv"
	"time"
)

// ErrInvalidNetAddr describes an error that indicates the caller didn't specify
// a TCP address as required.
var ErrInvalidNetAddr = errors.New("provided net.Addr is not a net.TCPAddr")

// MaxNetAddressPayload returns the max payload size for the NetAddress
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
	// not present in the version message (MsgVersion) nor was it
	// added until protocol version >= NetAddressTimeVersion.
	Timestamp time.Time

	// Bitfield which identifies the services supported by the address.
	Services protocol.ServiceFlag

	// IP address of the peer.
	IP net.IP

	// Port the peer is using.  This is encoded in big endian on the wire
	// which differs from most everything else.
	Port uint16
}

// HasService returns whether the specified service is supported by the address.
func (na *NetAddress) HasService(service protocol.ServiceFlag) bool {
	return na.Services&service == service
}

// AddService adds service as a supported service by the peer generating the
// message.
func (na *NetAddress) AddService(service protocol.ServiceFlag) {
	na.Services |= service
}

// NewNetAddressIPPort returns a new NetAddress using the provided IP, port, and
// supported services with defaults for the remaining fields.
func NewNetAddressIPPort(ip net.IP, port uint16, services protocol.ServiceFlag) *NetAddress {
	return NewNetAddressTimestamp(roughtime.Now(), services, ip, port)
}

// NewNetAddressTimestamp returns a new NetAddress using the provided
// timestamp, IP, port, and supported services. The timestamp is rounded to
// single second precision.
func NewNetAddressTimestamp(
	timestamp time.Time, services protocol.ServiceFlag, ip net.IP, port uint16) *NetAddress {
	// Limit the timestamp to one second precision since the protocol
	// doesn't support better.
	na := NetAddress{
		Timestamp: time.Unix(timestamp.Unix(), 0),
		Services:  services,
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
func NewNetAddress(addr net.Addr, services protocol.ServiceFlag) (*NetAddress, error) {
	tcpAddr, ok := addr.(*net.TCPAddr)
	if !ok {
		return nil, ErrInvalidNetAddr
	}

	na := NewNetAddressIPPort(tcpAddr.IP, uint16(tcpAddr.Port), services)
	return na, nil
}

// NewNetAddressFailBack add more attempts to extract the IP address and port from the passed
// net.Addr interface and create a NetAddress structure using that information.
func NewNetAddressFailBack(addr net.Addr, services protocol.ServiceFlag) (*NetAddress, error) {
	// addr will be a net.TCPAddr when not using a proxy.
	if tcpAddr, ok := addr.(*net.TCPAddr); ok {
		ip := tcpAddr.IP
		port := uint16(tcpAddr.Port)
		na := NewNetAddressIPPort(ip, port, services)
		return na, nil
	}

	// addr will be a socks.ProxiedAddr when using a proxy.
	if proxiedAddr, ok := addr.(*network.ProxiedAddr); ok {
		ip := net.ParseIP(proxiedAddr.Host)
		if ip == nil {
			ip = net.ParseIP("0.0.0.0")
		}
		port := uint16(proxiedAddr.Port)
		na := NewNetAddressIPPort(ip, port, services)
		return na, nil
	}

	// For the most part, addr should be one of the two above cases, but
	// to be safe, fall back to trying to parse the information from the
	// address string as a last resort.
	host, portStr, err := net.SplitHostPort(addr.String())
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(host)
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return nil, err
	}
	na := NewNetAddressIPPort(ip, uint16(port), services)
	return na, nil
}

// readNetAddress reads an encoded NetAddress from r depending on the protocol
// version and whether or not the timestamp is included per ts.  Some messages
// like version do not include the timestamp.
func ReadNetAddress(r io.Reader, pver uint32, na *NetAddress, ts bool) error {
	var ip [16]byte

	// TODO fix time ambiguous
	// NOTE: The protocol uses a uint32 for the timestamp so it will
	// stop working somewhere around 2106.  Also timestamp wasn't added until
	// protocol version >= NetAddressTimeVersion
	if ts {
		err := s.ReadElements(r, (*s.Uint32Time)(&na.Timestamp))
		if err != nil {
			return err
		}
	}

	err := s.ReadElements(r, &na.Services, &ip)
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
		Services:  na.Services,
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
	err := s.WriteElements(w, na.Services, ip)
	if err != nil {
		return err
	}

	// TODO unify endian
	// Sigh.  protocol mixes little and big endian.
	return binary.Write(w, binary.BigEndian, na.Port)
}
