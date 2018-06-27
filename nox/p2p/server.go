package p2p


// Use start to begin accepting connections from peers.
// peer server handling communications to and from nox peers.
type PeerServer struct{
    // address manager caching the peers
	addrManager          *AddrManager
	// conn manager handles network connections.
	connManager          *ConnManager
}

type AddrManager struct {}
type ConnManager struct {}
