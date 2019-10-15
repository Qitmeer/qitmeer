package config

import (
	"github.com/Qitmeer/qitmeer/core/types"
	"net"
	"time"
)

type Config struct {
	HomeDir              string        `short:"A" long:"appdata" description:"Path to application home directory"`
	ShowVersion          bool          `short:"V" long:"version" description:"Display version information and exit"`
	ConfigFile           string        `short:"C" long:"configfile" description:"Path to configuration file"`
	DataDir              string        `short:"b" long:"datadir" description:"Directory to store data"`
	LogDir               string        `long:"logdir" description:"Directory to log output."`
	NoFileLogging        bool          `long:"nofilelogging" description:"Disable file logging."`
	Listeners            []string      `long:"listen" description:"Add an interface/port to listen for connections (default all interfaces port: 8130, testnet: 18130)"`
	RPCListeners         []string      `long:"rpclisten" description:"Add an interface/port to listen for RPC connections (default port: 8131 , testnet: 18131)"`
	MaxPeers             int           `long:"maxpeers" description:"Max number of inbound and outbound peers"`
	DisableListen        bool          `long:"nolisten" description:"Disable listening for incoming connections"`
	RPCUser              string        `short:"u" long:"rpcuser" description:"Username for RPC connections"`
	RPCPass              string        `short:"P" long:"rpcpass" default-mask:"-" description:"Password for RPC connections"`
	RPCCert              string        `long:"rpccert" description:"File containing the certificate file"`
	RPCKey               string        `long:"rpckey" description:"File containing the certificate key"`
	RPCMaxClients        int           `long:"rpcmaxclients" description:"Max number of RPC clients for standard connections"`
	DisableRPC           bool          `long:"norpc" description:"Disable built-in RPC server -- NOTE: The RPC server is disabled by default if no rpcuser/rpcpass or rpclimituser/rpclimitpass is specified"`
	DisableTLS           bool          `long:"notls" description:"Disable TLS for the RPC server -- NOTE: This is only allowed if the RPC server is bound to localhost"`
	Modules              []string      `long:"modules" description:"Modules is a list of API modules(See GetNodeInfo) to expose via the HTTP RPC interface. If the module list is empty, all RPC API endpoints designated public will be exposed."`
	DisableDNSSeed       bool          `long:"nodnsseed" description:"Disable DNS seeding for peers"`
	DisableCheckpoints   bool          `long:"nocheckpoints" description:"Disable built-in checkpoints.  Don't do this unless you know what you're doing."`
	TxIndex              bool          `long:"txindex" description:"Maintain a full hash-based transaction index which makes all transactions available via the getrawtransaction RPC"`
	DropTxIndex          bool          `long:"droptxindex" description:"Deletes the hash-based transaction index from the database on start up and then exits."`
	AddrIndex            bool          `long:"addrindex" description:"Maintain a full address-based transaction index which makes the getrawtransactions RPC available"`
	DropAddrIndex        bool          `long:"dropaddrindex" description:"Deletes the address-based transaction index from the database on start up and then exits."`
	LightNode            bool          `long:"light" description:"start as a qitmeer light node"`
	SigCacheMaxSize      uint          `long:"sigcachemaxsize" description:"The maximum number of entries in the signature verification cache"`
	DumpBlockchain       string        `long:"dumpblockchain" description:"Write blockchain as a flat file of blocks for use with addblock, to the specified filename"`
	TestNet              bool          `long:"testnet" description:"Use the test network"`
	MixNet               bool          `long:"mixnet" description:"Use the test mix pow network"`
	PrivNet              bool          `long:"privnet" description:"Use the private network"`
	DbType               string        `long:"dbtype" description:"Database backend to use for the Block Chain"`
	Profile              string        `long:"profile" description:"Enable HTTP profiling on given [addr:]port -- NOTE port must be between 1024 and 65536"`
	DebugLevel           string        `short:"d" long:"debuglevel" description:"Logging level {trace, debug, info, warn, error, critical} "`
	DebugPrintOrigins    bool          `long:"printorigin" description:"Print log debug location (file:line) "`
	// MemPool Config
	NoRelayPriority      bool          `long:"norelaypriority" description:"Do not require free or low-fee transactions to have high priority for relaying"`
	FreeTxRelayLimit     float64       `long:"limitfreerelay" description:"Limit relay of transactions with no transaction fee to the given amount in thousands of bytes per minute"`
	AcceptNonStd         bool          `long:"acceptnonstd" description:"Accept and relay non-standard transactions to the network regardless of the default settings for the active network."`
	MaxOrphanTxs         int           `long:"maxorphantx" description:"Max number of orphan transactions to keep in memory"`
	MinTxFee             int64         `long:"mintxfee" description:"The minimum transaction fee in AtomMEER/kB."`
	// Miner
	Generate             bool          `long:"generate" description:"Generate (mine) coins using the CPU"`
	MiningAddrs          []string      `long:"miningaddr" description:"Add the specified payment address to the list of addresses to use for generated blocks -- At least one address is required if the generate option is set"`
	MiningTimeOffset     int           `long:"miningtimeoffset" description:"Offset the mining timestamp of a block by this many seconds (positive values are in the past)"`
	BlockMinSize         uint32        `long:"blockminsize" description:"Mininum block size in bytes to be used when creating a block"`
	BlockMaxSize         uint32        `long:"blockmaxsize" description:"Maximum block size in bytes to be used when creating a block"`
	BlockPrioritySize    uint32        `long:"blockprioritysize" description:"Size in bytes for high-priority/low-fee transactions when creating a block"`
	miningAddrs          []types.Address
	//WebSocket support
	RPCMaxWebsockets     int           `long:"rpcmaxwebsockets" description:"Max number of RPC websocket connections"`
	//P2P
	BlocksOnly           bool          `long:"blocksonly" description:"Do not accept transactions from remote peers."`
	MiningStateSync      bool          `long:"miningstatesync" description:"Synchronizing the mining state with other nodes"`
	AddPeers             []string      `short:"a" long:"addpeer" description:"Add a peer to connect with at startup"`
	ConnectPeers         []string      `long:"connect" description:"Connect only to the specified peers at startup"`
	ExternalIPs          []string      `long:"externalip" description:"list of local addresses we claim to listen on to peers"`
	Upnp                 bool          `long:"upnp" description:"Use UPnP to map our listening port outside of NAT"`
	Whitelists           []string      `long:"whitelist" description:"Add an IP network or IP that will not be banned. (eg. 192.168.1.0/24 or ::1)"`
	whitelists           []*net.IPNet
	//P2P - server ban
	DisableBanning       bool          `long:"nobanning" description:"Disable banning of misbehaving peers"`
	BanDuration          time.Duration `long:"banduration" description:"How long to ban misbehaving peers.  Valid time units are {s, m, h}.  Minimum 1 second"`
	BanThreshold         uint32        `long:"banthreshold" description:"Maximum allowed ban score before disconnecting and banning misbehaving peers."`
	GetAddrPercent       int           `short:"T" long:"getaddrpercent" description:"It is the percentage of total addresses known that we will share with a call to AddressCache."`

	DAGType              string        `short:"G" long:"dagtype" description:"DAG type {phantom,conflux,spectre} "`
	Cleanup              bool          `short:"L" long:"cleanup" description:"Cleanup the block database "`
	BuildLedger          bool          `long:"buildledger" description:"Generate the genesis ledger for the next qitmeer version."`
}

func (c *Config) GetMinningAddrs() []types.Address {
	return c.miningAddrs
}

func (c *Config) SetMiningAddrs(addr types.Address) {
	c.miningAddrs = append(c.miningAddrs,addr)
}
func (c *Config) GetWhitelists() []*net.IPNet {
	return c.whitelists
}

func (c *Config) AddToWhitelists(ip *net.IPNet) {
	c.whitelists = append(c.whitelists,ip)
}

