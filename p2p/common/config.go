/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package common

import (
	"github.com/Qitmeer/qng-core/core/protocol"
	"github.com/Qitmeer/qng-core/params"
	"os"
)

// Config for the p2p service.
// to initialize the p2p service.
type Config struct {
	NoDiscovery          bool
	EnableUPnP           bool
	StaticPeers          []string
	BootstrapNodeAddr    []string
	Discv5BootStrapAddr  []string
	DataDir              string
	MaxPeers             uint
	MaxInbound           int
	MetaDataDir          string
	ReadWritePermissions os.FileMode
	AllowListCIDR        string
	DenyListCIDR         []string
	TCPPort              uint
	UDPPort              uint
	EnableNoise          bool
	RelayNodeAddr        string
	LocalIP              string
	HostAddress          string
	HostDNS              string
	PrivateKey           string
	Encoding             string
	// ProtocolVersion specifies the maximum protocol version to use and
	// advertise.
	ProtocolVersion uint32
	Services        protocol.ServiceFlag
	UserAgent       string
	// DisableRelayTx specifies if the remote peer should be informed to
	// not send inv messages for transactions.
	DisableRelayTx bool
	MaxOrphanTxs   int
	Params         *params.Params
	Banning        bool // Open or not ban module
	DisableListen  bool
	LANPeers       []string
	IsCircuit      bool
}
