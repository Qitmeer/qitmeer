package p2p

import "os"

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
	MetaDataDir          string
	ReadWritePermissions os.FileMode
	AllowListCIDR        string
	DenyListCIDR         []string
	TCPPort              uint
	UDPPort              uint
	// EnableNoise enables the beacon node to use NOISE instead of SECIO when performing a handshake with another peer.
	EnableNoise   bool
	RelayNodeAddr string

	//TODO

	LocalIP     string
	HostAddress string
	HostDNS     string
	PrivateKey  string
	Encoding    string
}
