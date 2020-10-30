/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package discover

import (
	"bytes"
	"container/list"
	"context"
	"crypto/ecdsa"
	crand "crypto/rand"
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/crypto/ecc/secp256k1"
	"io"
	"net"
	"sync"
	"time"

	"github.com/Qitmeer/qitmeer/common/encode/rlp"
	"github.com/Qitmeer/qitmeer/crypto"
	"github.com/Qitmeer/qitmeer/p2p/netutil"
	"github.com/Qitmeer/qitmeer/p2p/qnode"
	"github.com/Qitmeer/qitmeer/p2p/qnr"
)

// Errors
var (
	errPacketTooSmall   = errors.New("too small")
	errBadHash          = errors.New("bad hash")
	errExpired          = errors.New("expired")
	errUnsolicitedReply = errors.New("unsolicited reply")
	errUnknownNode      = errors.New("unknown node")
	errTimeout          = errors.New("RPC timeout")
	errClockWarp        = errors.New("reply deadline too far in the future")
	errClosed           = errors.New("socket closed")
	errLowPort          = errors.New("low port")
)

const (
	respTimeout    = 500 * time.Millisecond
	expiration     = 20 * time.Second
	bondExpiration = 24 * time.Hour

	maxFindnodeFailures = 5                // nodes exceeding this limit are dropped
	ntpFailureThreshold = 32               // Continuous timeouts after which to check NTP
	ntpWarningCooldown  = 10 * time.Minute // Minimum amount of time to pass before repeating NTP warning
	driftThreshold      = 10 * time.Second // Allowed clock drift before warning user

	// Discovery packets are defined to be no larger than 1280 bytes.
	// Packets larger than this size will be cut at the end and treated
	// as invalid because their hash won't match.
	maxPacketSize = 1280
)

// RPC packet types
const (
	p_pingV4 = iota + 1 // zero is 'reserved'
	p_pongV4
	p_findnodeV4
	p_neighborsV4
	p_qnrRequestV4
	p_qnrResponseV4
)

// RPC request structures
type (
	pingV4 struct {
		senderKey *ecdsa.PublicKey // filled in by preverify

		Version    uint
		From, To   rpcEndpoint
		Expiration uint64
		// Ignore additional fields (for forward compatibility).
		Rest []rlp.RawValue `rlp:"tail"`
	}

	// pongV4 is the reply to pingV4.
	pongV4 struct {
		// This field should mirror the UDP envelope address
		// of the ping packet, which provides a way to discover the
		// the external address (after NAT).
		To rpcEndpoint

		ReplyTok   []byte // This contains the hash of the ping packet.
		Expiration uint64 // Absolute timestamp at which the packet becomes invalid.
		// Ignore additional fields (for forward compatibility).
		Rest []rlp.RawValue `rlp:"tail"`
	}

	// findnodeV4 is a query for nodes close to the given target.
	findnodeV4 struct {
		Target     encPubkey
		Expiration uint64
		// Ignore additional fields (for forward compatibility).
		Rest []rlp.RawValue `rlp:"tail"`
	}

	// neighborsV4 is the reply to findnodeV4.
	neighborsV4 struct {
		Nodes      []rpcNode
		Expiration uint64
		// Ignore additional fields (for forward compatibility).
		Rest []rlp.RawValue `rlp:"tail"`
	}

	// qnrRequestV4 queries for the remote node's record.
	qnrRequestV4 struct {
		Expiration uint64
		// Ignore additional fields (for forward compatibility).
		Rest []rlp.RawValue `rlp:"tail"`
	}

	// qnrResponseV4 is the reply to qnrRequestV4.
	qnrResponseV4 struct {
		ReplyTok []byte // Hash of the qnrRequest packet.
		Record   qnr.Record
		// Ignore additional fields (for forward compatibility).
		Rest []rlp.RawValue `rlp:"tail"`
	}

	rpcNode struct {
		IP  net.IP // len 4 for IPv4 or 16 for IPv6
		UDP uint16 // for discovery protocol
		TCP uint16 // for RLPx protocol
		ID  encPubkey
	}

	rpcEndpoint struct {
		IP  net.IP // len 4 for IPv4 or 16 for IPv6
		UDP uint16 // for discovery protocol
		TCP uint16 // for RLPx protocol
	}
)

// packetV4 is implemented by all v4 protocol messages.
type packetV4 interface {
	// preverify checks whether the packet is valid and should be handled at all.
	preverify(t *UDPv4, from *net.UDPAddr, fromID qnode.ID, fromKey encPubkey) error
	// handle handles the packet.
	handle(t *UDPv4, from *net.UDPAddr, fromID qnode.ID, mac []byte)
	// packet name and type for logging purposes.
	name() string
	kind() byte
}

func makeEndpoint(addr *net.UDPAddr, tcpPort uint16) rpcEndpoint {
	ip := net.IP{}
	if ip4 := addr.IP.To4(); ip4 != nil {
		ip = ip4
	} else if ip6 := addr.IP.To16(); ip6 != nil {
		ip = ip6
	}
	return rpcEndpoint{IP: ip, UDP: uint16(addr.Port), TCP: tcpPort}
}

func (t *UDPv4) nodeFromRPC(sender *net.UDPAddr, rn rpcNode) (*node, error) {
	if rn.UDP <= 1024 {
		return nil, errLowPort
	}
	if err := netutil.CheckRelayIP(sender.IP, rn.IP); err != nil {
		return nil, err
	}
	if t.netrestrict != nil && !t.netrestrict.Contains(rn.IP) {
		return nil, errors.New("not contained in netrestrict whitelist")
	}
	key, err := decodePubkey(secp256k1.S256(), rn.ID)
	if err != nil {
		return nil, err
	}
	n := wrapNode(qnode.NewV4(key, rn.IP, int(rn.TCP), int(rn.UDP)))
	err = n.ValidateComplete()
	return n, err
}

func nodeToRPC(n *node) rpcNode {
	var key ecdsa.PublicKey
	var ekey encPubkey
	if err := n.Load((*qnode.Secp256k1)(&key)); err == nil {
		ekey = encodePubkey(&key)
	}
	return rpcNode{ID: ekey, IP: n.IP(), UDP: uint16(n.UDP()), TCP: uint16(n.TCP())}
}

// UDPv4 implements the v4 wire protocol.
type UDPv4 struct {
	conn        UDPConn
	netrestrict *netutil.Netlist
	priv        *ecdsa.PrivateKey
	localNode   *qnode.LocalNode
	db          *qnode.DB
	tab         *Table
	closeOnce   sync.Once
	wg          sync.WaitGroup

	addReplyMatcher chan *replyMatcher
	gotreply        chan reply
	closeCtx        context.Context
	cancelCloseCtx  context.CancelFunc
}

// replyMatcher represents a pending reply.
//
// Some implementations of the protocol wish to send more than one
// reply packet to findnode. In general, any neighbors packet cannot
// be matched up with a specific findnode packet.
//
// Our implementation handles this by storing a callback function for
// each pending reply. Incoming packets from a node are dispatched
// to all callback functions for that node.
type replyMatcher struct {
	// these fields must match in the reply.
	from  qnode.ID
	ip    net.IP
	ptype byte

	// time when the request must complete
	deadline time.Time

	// callback is called when a matching reply arrives. If it returns matched == true, the
	// reply was acceptable. The second return value indicates whether the callback should
	// be removed from the pending reply queue. If it returns false, the reply is considered
	// incomplete and the callback will be invoked again for the next matching reply.
	callback replyMatchFunc

	// errc receives nil when the callback indicates completion or an
	// error if no further reply is received within the timeout.
	errc chan error

	// reply contains the most recent reply. This field is safe for reading after errc has
	// received a value.
	reply packetV4
}

type replyMatchFunc func(interface{}) (matched bool, requestDone bool)

// reply is a reply packet from a certain node.
type reply struct {
	from qnode.ID
	ip   net.IP
	data packetV4
	// loop indicates whether there was
	// a matching request by sending on this channel.
	matched chan<- bool
}

func ListenV4(c UDPConn, ln *qnode.LocalNode, cfg Config) (*UDPv4, error) {
	cfg = cfg.withDefaults()
	closeCtx, cancel := context.WithCancel(context.Background())
	t := &UDPv4{
		conn:            c,
		priv:            cfg.PrivateKey,
		netrestrict:     cfg.NetRestrict,
		localNode:       ln,
		db:              ln.Database(),
		gotreply:        make(chan reply),
		addReplyMatcher: make(chan *replyMatcher),
		closeCtx:        closeCtx,
		cancelCloseCtx:  cancel,
	}

	tab, err := newTable(t, ln.Database(), cfg.Bootnodes)
	if err != nil {
		return nil, err
	}
	t.tab = tab
	go tab.loop()

	t.wg.Add(2)
	go t.loop()
	go t.readLoop(cfg.Unhandled)
	return t, nil
}

// Self returns the local node.
func (t *UDPv4) Self() *qnode.Node {
	return t.localNode.Node()
}

// Close shuts down the socket and aborts any running queries.
func (t *UDPv4) Close() {
	t.closeOnce.Do(func() {
		t.cancelCloseCtx()
		t.conn.Close()
		t.wg.Wait()
		t.tab.close()
	})
}

// Resolve searches for a specific node with the given ID and tries to get the most recent
// version of the node record for it. It returns n if the node could not be resolved.
func (t *UDPv4) Resolve(n *qnode.Node) *qnode.Node {
	// Try asking directly. This works if the node is still responding on the endpoint we have.
	if rn, err := t.RequestQNR(n); err == nil {
		return rn
	}
	// Check table for the ID, we might have a newer version there.
	if intable := t.tab.getNode(n.ID()); intable != nil && intable.Seq() > n.Seq() {
		n = intable
		if rn, err := t.RequestQNR(n); err == nil {
			return rn
		}
	}
	// Otherwise perform a network lookup.
	var key qnode.Secp256k1
	if n.Load(&key) != nil {
		return n // no secp256k1 key
	}
	result := t.LookupPubkey((*ecdsa.PublicKey)(&key))
	for _, rn := range result {
		if rn.ID() == n.ID() {
			if rn, err := t.RequestQNR(rn); err == nil {
				return rn
			}
		}
	}
	return n
}

func (t *UDPv4) ourEndpoint() rpcEndpoint {
	n := t.Self()
	a := &net.UDPAddr{IP: n.IP(), Port: n.UDP()}
	return makeEndpoint(a, uint16(n.TCP()))
}

// Ping sends a ping message to the given node.
func (t *UDPv4) Ping(n *qnode.Node) error {
	_, err := t.ping(n)
	return err
}

// ping sends a ping message to the given node and waits for a reply.
func (t *UDPv4) ping(n *qnode.Node) (seq uint64, err error) {
	rm := t.sendPing(n.ID(), &net.UDPAddr{IP: n.IP(), Port: n.UDP()}, nil)
	if err = <-rm.errc; err == nil {
		seq = seqFromTail(rm.reply.(*pongV4).Rest)
	}
	return seq, err
}

// sendPing sends a ping message to the given node and invokes the callback
// when the reply arrives.
func (t *UDPv4) sendPing(toid qnode.ID, toaddr *net.UDPAddr, callback func()) *replyMatcher {
	req := t.makePing(toaddr)
	packet, hash, err := t.encode(t.priv, req)
	if err != nil {
		errc := make(chan error, 1)
		errc <- err
		return &replyMatcher{errc: errc}
	}
	// Add a matcher for the reply to the pending reply queue. Pongs are matched if they
	// reference the ping we're about to send.
	rm := t.pending(toid, toaddr.IP, p_pongV4, func(p interface{}) (matched bool, requestDone bool) {
		matched = bytes.Equal(p.(*pongV4).ReplyTok, hash)
		if matched && callback != nil {
			callback()
		}
		return matched, matched
	})
	// Send the packet.
	t.localNode.UDPContact(toaddr)
	t.write(toaddr, toid, req.name(), packet)
	return rm
}

func (t *UDPv4) makePing(toaddr *net.UDPAddr) *pingV4 {
	seq, _ := rlp.EncodeToBytes(t.localNode.Node().Seq())
	return &pingV4{
		Version:    4,
		From:       t.ourEndpoint(),
		To:         makeEndpoint(toaddr, 0),
		Expiration: uint64(time.Now().Add(expiration).Unix()),
		Rest:       []rlp.RawValue{seq},
	}
}

// LookupPubkey finds the closest nodes to the given public key.
func (t *UDPv4) LookupPubkey(key *ecdsa.PublicKey) []*qnode.Node {
	if t.tab.len() == 0 {
		// All nodes were dropped, refresh. The very first query will hit this
		// case and run the bootstrapping logic.
		<-t.tab.refresh()
	}
	return t.newLookup(t.closeCtx, encodePubkey(key)).run()
}

// RandomNodes is an iterator yielding nodes from a random walk of the DHT.
func (t *UDPv4) RandomNodes() qnode.Iterator {
	return newLookupIterator(t.closeCtx, t.newRandomLookup)
}

// lookupRandom implements transport.
func (t *UDPv4) lookupRandom() []*qnode.Node {
	return t.newRandomLookup(t.closeCtx).run()
}

// lookupSelf implements transport.
func (t *UDPv4) lookupSelf() []*qnode.Node {
	return t.newLookup(t.closeCtx, encodePubkey(&t.priv.PublicKey)).run()
}

func (t *UDPv4) newRandomLookup(ctx context.Context) *lookup {
	var target encPubkey
	crand.Read(target[:])
	return t.newLookup(ctx, target)
}

func (t *UDPv4) newLookup(ctx context.Context, targetKey encPubkey) *lookup {
	target := qnode.ID(crypto.Keccak256Hash(targetKey[:]))
	it := newLookup(ctx, t.tab, target, func(n *node) ([]*node, error) {
		return t.findnode(n.ID(), n.addr(), targetKey)
	})
	return it
}

// findnode sends a findnode request to the given node and waits until
// the node has sent up to k neighbors.
func (t *UDPv4) findnode(toid qnode.ID, toaddr *net.UDPAddr, target encPubkey) ([]*node, error) {
	t.ensureBond(toid, toaddr)

	// Add a matcher for 'neighbours' replies to the pending reply queue. The matcher is
	// active until enough nodes have been received.
	nodes := make([]*node, 0, bucketSize)
	nreceived := 0
	rm := t.pending(toid, toaddr.IP, p_neighborsV4, func(r interface{}) (matched bool, requestDone bool) {
		reply := r.(*neighborsV4)
		for _, rn := range reply.Nodes {
			nreceived++
			n, err := t.nodeFromRPC(toaddr, rn)
			if err != nil {
				log.Trace("Invalid neighbor node received", "ip", rn.IP, "addr", toaddr, "err", err)
				continue
			}
			nodes = append(nodes, n)
		}
		return true, nreceived >= bucketSize
	})
	t.send(toaddr, toid, &findnodeV4{
		Target:     target,
		Expiration: uint64(time.Now().Add(expiration).Unix()),
	})
	return nodes, <-rm.errc
}

// RequestQNR sends qnrRequest to the given node and waits for a response.
func (t *UDPv4) RequestQNR(n *qnode.Node) (*qnode.Node, error) {
	addr := &net.UDPAddr{IP: n.IP(), Port: n.UDP()}
	t.ensureBond(n.ID(), addr)

	req := &qnrRequestV4{
		Expiration: uint64(time.Now().Add(expiration).Unix()),
	}
	packet, hash, err := t.encode(t.priv, req)
	if err != nil {
		return nil, err
	}
	// Add a matcher for the reply to the pending reply queue. Responses are matched if
	// they reference the request we're about to send.
	rm := t.pending(n.ID(), addr.IP, p_qnrResponseV4, func(r interface{}) (matched bool, requestDone bool) {
		matched = bytes.Equal(r.(*qnrResponseV4).ReplyTok, hash)
		return matched, matched
	})
	// Send the packet and wait for the reply.
	t.write(addr, n.ID(), req.name(), packet)
	if err := <-rm.errc; err != nil {
		return nil, err
	}
	// Verify the response record.
	respN, err := qnode.New(qnode.ValidSchemes, &rm.reply.(*qnrResponseV4).Record)
	if err != nil {
		return nil, err
	}
	if respN.ID() != n.ID() {
		return nil, fmt.Errorf("invalid ID in response record")
	}
	if respN.Seq() < n.Seq() {
		return n, nil // response record is older
	}
	if err := netutil.CheckRelayIP(addr.IP, respN.IP()); err != nil {
		return nil, fmt.Errorf("invalid IP in response record: %v", err)
	}
	return respN, nil
}

// pending adds a reply matcher to the pending reply queue.
// see the documentation of type replyMatcher for a detailed explanation.
func (t *UDPv4) pending(id qnode.ID, ip net.IP, ptype byte, callback replyMatchFunc) *replyMatcher {
	ch := make(chan error, 1)
	p := &replyMatcher{from: id, ip: ip, ptype: ptype, callback: callback, errc: ch}
	select {
	case t.addReplyMatcher <- p:
		// loop will handle it
	case <-t.closeCtx.Done():
		ch <- errClosed
	}
	return p
}

// handleReply dispatches a reply packet, invoking reply matchers. It returns
// whether any matcher considered the packet acceptable.
func (t *UDPv4) handleReply(from qnode.ID, fromIP net.IP, req packetV4) bool {
	matched := make(chan bool, 1)
	select {
	case t.gotreply <- reply{from, fromIP, req, matched}:
		// loop will handle it
		return <-matched
	case <-t.closeCtx.Done():
		return false
	}
}

// loop runs in its own goroutine. it keeps track of
// the refresh timer and the pending reply queue.
func (t *UDPv4) loop() {
	defer t.wg.Done()

	var (
		plist        = list.New()
		timeout      = time.NewTimer(0)
		nextTimeout  *replyMatcher // head of plist when timeout was last reset
		contTimeouts = 0           // number of continuous timeouts to do NTP checks
		ntpWarnTime  = time.Unix(0, 0)
	)
	<-timeout.C // ignore first timeout
	defer timeout.Stop()

	resetTimeout := func() {
		if plist.Front() == nil || nextTimeout == plist.Front().Value {
			return
		}
		// Start the timer so it fires when the next pending reply has expired.
		now := time.Now()
		for el := plist.Front(); el != nil; el = el.Next() {
			nextTimeout = el.Value.(*replyMatcher)
			if dist := nextTimeout.deadline.Sub(now); dist < 2*respTimeout {
				timeout.Reset(dist)
				return
			}
			// Remove pending replies whose deadline is too far in the
			// future. These can occur if the system clock jumped
			// backwards after the deadline was assigned.
			nextTimeout.errc <- errClockWarp
			plist.Remove(el)
		}
		nextTimeout = nil
		timeout.Stop()
	}

	for {
		resetTimeout()

		select {
		case <-t.closeCtx.Done():
			for el := plist.Front(); el != nil; el = el.Next() {
				el.Value.(*replyMatcher).errc <- errClosed
			}
			return

		case p := <-t.addReplyMatcher:
			p.deadline = time.Now().Add(respTimeout)
			plist.PushBack(p)

		case r := <-t.gotreply:
			var matched bool // whether any replyMatcher considered the reply acceptable.
			for el := plist.Front(); el != nil; el = el.Next() {
				p := el.Value.(*replyMatcher)
				if p.from == r.from && p.ptype == r.data.kind() && p.ip.Equal(r.ip) {
					ok, requestDone := p.callback(r.data)
					matched = matched || ok
					// Remove the matcher if callback indicates that all replies have been received.
					if requestDone {
						p.reply = r.data
						p.errc <- nil
						plist.Remove(el)
					}
					// Reset the continuous timeout counter (time drift detection)
					contTimeouts = 0
				}
			}
			r.matched <- matched

		case now := <-timeout.C:
			nextTimeout = nil

			// Notify and remove callbacks whose deadline is in the past.
			for el := plist.Front(); el != nil; el = el.Next() {
				p := el.Value.(*replyMatcher)
				if now.After(p.deadline) || now.Equal(p.deadline) {
					p.errc <- errTimeout
					plist.Remove(el)
					contTimeouts++
				}
			}
			// If we've accumulated too many timeouts, do an NTP time sync check
			if contTimeouts > ntpFailureThreshold {
				if time.Since(ntpWarnTime) >= ntpWarningCooldown {
					ntpWarnTime = time.Now()
					go checkClockDrift()
				}
				contTimeouts = 0
			}
		}
	}
}

const (
	macSize  = 256 / 8
	sigSize  = 520 / 8
	headSize = macSize + sigSize // space of packet frame data
)

var (
	headSpace = make([]byte, headSize)

	// Neighbors replies are sent across multiple packets to
	// stay below the packet size limit. We compute the maximum number
	// of entries by stuffing a packet until it grows too large.
	maxNeighbors int
)

func init() {
	p := neighborsV4{Expiration: ^uint64(0)}
	maxSizeNode := rpcNode{IP: make(net.IP, 16), UDP: ^uint16(0), TCP: ^uint16(0)}
	for n := 0; ; n++ {
		p.Nodes = append(p.Nodes, maxSizeNode)
		size, _, err := rlp.EncodeToReader(p)
		if err != nil {
			// If this ever happens, it will be caught by the unit tests.
			panic("cannot encode: " + err.Error())
		}
		if headSize+size+1 >= maxPacketSize {
			maxNeighbors = n
			break
		}
	}
}

func (t *UDPv4) send(toaddr *net.UDPAddr, toid qnode.ID, req packetV4) ([]byte, error) {
	packet, hash, err := t.encode(t.priv, req)
	if err != nil {
		return hash, err
	}
	return hash, t.write(toaddr, toid, req.name(), packet)
}

func (t *UDPv4) write(toaddr *net.UDPAddr, toid qnode.ID, what string, packet []byte) error {
	_, err := t.conn.WriteToUDP(packet, toaddr)
	log.Trace(">> "+what, "id", toid, "addr", toaddr, "err", err)
	return err
}

func (t *UDPv4) encode(priv *ecdsa.PrivateKey, req packetV4) (packet, hash []byte, err error) {
	name := req.name()
	b := new(bytes.Buffer)
	b.Write(headSpace)
	b.WriteByte(req.kind())
	if err := rlp.Encode(b, req); err != nil {
		log.Error(fmt.Sprintf("Can't encode %s packet", name), "err", err)
		return nil, nil, err
	}
	packet = b.Bytes()
	sig, err := crypto.Sign(crypto.Keccak256(packet[headSize:]), priv)
	if err != nil {
		log.Error(fmt.Sprintf("Can't sign %s packet", name), "err", err)
		return nil, nil, err
	}
	copy(packet[macSize:], sig)
	// add the hash to the front. Note: this doesn't protect the
	// packet in any way. Our public key will be part of this hash in
	// The future.
	hash = crypto.Keccak256(packet[macSize:])
	copy(packet, hash)
	return packet, hash, nil
}

// readLoop runs in its own goroutine. it handles incoming UDP packets.
func (t *UDPv4) readLoop(unhandled chan<- ReadPacket) {
	defer t.wg.Done()
	if unhandled != nil {
		defer close(unhandled)
	}

	buf := make([]byte, maxPacketSize)
	for {
		nbytes, from, err := t.conn.ReadFromUDP(buf)
		if netutil.IsTemporaryError(err) {
			// Ignore temporary read errors.
			log.Debug("Temporary UDP read error", "err", err)
			continue
		} else if err != nil {
			// Shut down the loop for permament errors.
			if err != io.EOF {
				log.Debug("UDP read error", "err", err)
			}
			return
		}
		if t.handlePacket(from, buf[:nbytes]) != nil && unhandled != nil {
			select {
			case unhandled <- ReadPacket{buf[:nbytes], from}:
			default:
			}
		}
	}
}

func (t *UDPv4) handlePacket(from *net.UDPAddr, buf []byte) error {
	packet, fromKey, hash, err := decodeV4(buf)
	if err != nil {
		log.Debug("Bad discv4 packet", "addr", from, "err", err)
		return err
	}
	fromID := fromKey.id()
	if err == nil {
		err = packet.preverify(t, from, fromID, fromKey)
	}
	log.Trace("<< "+packet.name(), "id", fromID, "addr", from, "err", err)
	if err == nil {
		packet.handle(t, from, fromID, hash)
	}
	return err
}

func decodeV4(buf []byte) (packetV4, encPubkey, []byte, error) {
	if len(buf) < headSize+1 {
		return nil, encPubkey{}, nil, errPacketTooSmall
	}
	hash, sig, sigdata := buf[:macSize], buf[macSize:headSize], buf[headSize:]
	shouldhash := crypto.Keccak256(buf[macSize:])
	if !bytes.Equal(hash, shouldhash) {
		return nil, encPubkey{}, nil, errBadHash
	}
	fromKey, err := recoverNodeKey(crypto.Keccak256(buf[headSize:]), sig)
	if err != nil {
		return nil, fromKey, hash, err
	}

	var req packetV4
	switch ptype := sigdata[0]; ptype {
	case p_pingV4:
		req = new(pingV4)
	case p_pongV4:
		req = new(pongV4)
	case p_findnodeV4:
		req = new(findnodeV4)
	case p_neighborsV4:
		req = new(neighborsV4)
	case p_qnrRequestV4:
		req = new(qnrRequestV4)
	case p_qnrResponseV4:
		req = new(qnrResponseV4)
	default:
		return nil, fromKey, hash, fmt.Errorf("unknown type: %d", ptype)
	}
	s := rlp.NewStream(bytes.NewReader(sigdata[1:]), 0)
	err = s.Decode(req)
	return req, fromKey, hash, err
}

// checkBond checks if the given node has a recent enough endpoint proof.
func (t *UDPv4) checkBond(id qnode.ID, ip net.IP) bool {
	return time.Since(t.db.LastPongReceived(id, ip)) < bondExpiration
}

// ensureBond solicits a ping from a node if we haven't seen a ping from it for a while.
// This ensures there is a valid endpoint proof on the remote end.
func (t *UDPv4) ensureBond(toid qnode.ID, toaddr *net.UDPAddr) {
	tooOld := time.Since(t.db.LastPingReceived(toid, toaddr.IP)) > bondExpiration
	if tooOld || t.db.FindFails(toid, toaddr.IP) > maxFindnodeFailures {
		rm := t.sendPing(toid, toaddr, nil)
		<-rm.errc
		// Wait for them to ping back and process our pong.
		time.Sleep(respTimeout)
	}
}

// expired checks whether the given UNIX time stamp is in the past.
func expired(ts uint64) bool {
	return time.Unix(int64(ts), 0).Before(time.Now())
}

func seqFromTail(tail []rlp.RawValue) uint64 {
	if len(tail) == 0 {
		return 0
	}
	var seq uint64
	rlp.DecodeBytes(tail[0], &seq)
	return seq
}

// PING/v4

func (req *pingV4) name() string { return "PING/v4" }
func (req *pingV4) kind() byte   { return p_pingV4 }

func (req *pingV4) preverify(t *UDPv4, from *net.UDPAddr, fromID qnode.ID, fromKey encPubkey) error {
	if expired(req.Expiration) {
		return errExpired
	}
	key, err := decodePubkey(secp256k1.S256(), fromKey)
	if err != nil {
		return errors.New("invalid public key")
	}
	req.senderKey = key
	return nil
}

func (req *pingV4) handle(t *UDPv4, from *net.UDPAddr, fromID qnode.ID, mac []byte) {
	// Reply.
	seq, _ := rlp.EncodeToBytes(t.localNode.Node().Seq())
	t.send(from, fromID, &pongV4{
		To:         makeEndpoint(from, req.From.TCP),
		ReplyTok:   mac,
		Expiration: uint64(time.Now().Add(expiration).Unix()),
		Rest:       []rlp.RawValue{seq},
	})

	// Ping back if our last pong on file is too far in the past.
	n := wrapNode(qnode.NewV4(req.senderKey, from.IP, int(req.From.TCP), from.Port))
	if time.Since(t.db.LastPongReceived(n.ID(), from.IP)) > bondExpiration {
		t.sendPing(fromID, from, func() {
			t.tab.addVerifiedNode(n)
		})
	} else {
		t.tab.addVerifiedNode(n)
	}

	// Update node database and endpoint predictor.
	t.db.UpdateLastPingReceived(n.ID(), from.IP, time.Now())
	t.localNode.UDPEndpointStatement(from, &net.UDPAddr{IP: req.To.IP, Port: int(req.To.UDP)})
}

// PONG/v4

func (req *pongV4) name() string { return "PONG/v4" }
func (req *pongV4) kind() byte   { return p_pongV4 }

func (req *pongV4) preverify(t *UDPv4, from *net.UDPAddr, fromID qnode.ID, fromKey encPubkey) error {
	if expired(req.Expiration) {
		return errExpired
	}
	if !t.handleReply(fromID, from.IP, req) {
		return errUnsolicitedReply
	}
	return nil
}

func (req *pongV4) handle(t *UDPv4, from *net.UDPAddr, fromID qnode.ID, mac []byte) {
	t.localNode.UDPEndpointStatement(from, &net.UDPAddr{IP: req.To.IP, Port: int(req.To.UDP)})
	t.db.UpdateLastPongReceived(fromID, from.IP, time.Now())
}

// FINDNODE/v4

func (req *findnodeV4) name() string { return "FINDNODE/v4" }
func (req *findnodeV4) kind() byte   { return p_findnodeV4 }

func (req *findnodeV4) preverify(t *UDPv4, from *net.UDPAddr, fromID qnode.ID, fromKey encPubkey) error {
	if expired(req.Expiration) {
		return errExpired
	}
	if !t.checkBond(fromID, from.IP) {
		// No endpoint proof pong exists, we don't process the packet. This prevents an
		// attack vector where the discovery protocol could be used to amplify traffic in a
		// DDOS attack. A malicious actor would send a findnode request with the IP address
		// and UDP port of the target as the source address. The recipient of the findnode
		// packet would then send a neighbors packet (which is a much bigger packet than
		// findnode) to the victim.
		return errUnknownNode
	}
	return nil
}

func (req *findnodeV4) handle(t *UDPv4, from *net.UDPAddr, fromID qnode.ID, mac []byte) {
	// Determine closest nodes.
	target := qnode.ID(crypto.Keccak256Hash(req.Target[:]))
	t.tab.mutex.Lock()
	closest := t.tab.closest(target, bucketSize, true).entries
	t.tab.mutex.Unlock()

	// Send neighbors in chunks with at most maxNeighbors per packet
	// to stay below the packet size limit.
	p := neighborsV4{Expiration: uint64(time.Now().Add(expiration).Unix())}
	var sent bool
	for _, n := range closest {
		if netutil.CheckRelayIP(from.IP, n.IP()) == nil {
			p.Nodes = append(p.Nodes, nodeToRPC(n))
		}
		if len(p.Nodes) == maxNeighbors {
			t.send(from, fromID, &p)
			p.Nodes = p.Nodes[:0]
			sent = true
		}
	}
	if len(p.Nodes) > 0 || !sent {
		t.send(from, fromID, &p)
	}
}

// NEIGHBORS/v4

func (req *neighborsV4) name() string { return "NEIGHBORS/v4" }
func (req *neighborsV4) kind() byte   { return p_neighborsV4 }

func (req *neighborsV4) preverify(t *UDPv4, from *net.UDPAddr, fromID qnode.ID, fromKey encPubkey) error {
	if expired(req.Expiration) {
		return errExpired
	}
	if !t.handleReply(fromID, from.IP, req) {
		return errUnsolicitedReply
	}
	return nil
}

func (req *neighborsV4) handle(t *UDPv4, from *net.UDPAddr, fromID qnode.ID, mac []byte) {
}

// QNRREQUEST/v4

func (req *qnrRequestV4) name() string { return "QNRREQUEST/v4" }
func (req *qnrRequestV4) kind() byte   { return p_qnrRequestV4 }

func (req *qnrRequestV4) preverify(t *UDPv4, from *net.UDPAddr, fromID qnode.ID, fromKey encPubkey) error {
	if expired(req.Expiration) {
		return errExpired
	}
	if !t.checkBond(fromID, from.IP) {
		return errUnknownNode
	}
	return nil
}

func (req *qnrRequestV4) handle(t *UDPv4, from *net.UDPAddr, fromID qnode.ID, mac []byte) {
	t.send(from, fromID, &qnrResponseV4{
		ReplyTok: mac,
		Record:   *t.localNode.Node().Record(),
	})
}

// QNRRESPONSE/v4

func (req *qnrResponseV4) name() string { return "QNRRESPONSE/v4" }
func (req *qnrResponseV4) kind() byte   { return p_qnrResponseV4 }

func (req *qnrResponseV4) preverify(t *UDPv4, from *net.UDPAddr, fromID qnode.ID, fromKey encPubkey) error {
	if !t.handleReply(fromID, from.IP, req) {
		return errUnsolicitedReply
	}
	return nil
}

func (req *qnrResponseV4) handle(t *UDPv4, from *net.UDPAddr, fromID qnode.ID, mac []byte) {
}
