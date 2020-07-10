package p2p

import "os"

// Config for the p2p service.
// to initialize the p2p service.
type Config struct {
	NoDiscovery         bool
	EnableUPnP          bool
	DisableDiscv5       bool
	StaticPeers         []string
	BootstrapNodeAddr   []string
	Discv5BootStrapAddr []string
	RelayNodeAddr       string
	LocalIP             string
	HostAddress         string
	HostDNS             string
	PrivateKey          string
	DataDir             string
	MetaDataDir         string
	TCPPort             uint
	UDPPort             uint
	MaxPeers            uint
	AllowListCIDR       string
	DenyListCIDR        []string
	Encoding            string

	ReadWritePermissions os.FileMode
}
