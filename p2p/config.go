package p2p

import (
	"github.com/Qitmeer/qitmeer/core/protocol"
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
	MetaDataDir          string
	ReadWritePermissions os.FileMode
	AllowListCIDR        string
	DenyListCIDR         []string
	TCPPort              uint
	UDPPort              uint
	// EnableNoise enables the beacon node to use NOISE instead of SECIO when performing a handshake with another peer.
	EnableNoise   bool
	RelayNodeAddr string
	LocalIP       string
	HostAddress   string
	HostDNS       string
	PrivateKey    string
	Encoding      string
	// ProtocolVersion specifies the maximum protocol version to use and
	// advertise.
	ProtocolVersion uint32
	Services        protocol.ServiceFlag
	UserAgent       string
	// DisableRelayTx specifies if the remote peer should be informed to
	// not send inv messages for transactions.
	DisableRelayTx bool
}
