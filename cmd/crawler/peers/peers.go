package peers

import (
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"sync"
)

type Peer struct {
	Id        peer.ID
	Conn      network.Conn
	Connected bool
}

type Peers struct {
	list  map[peer.ID]*Peer
	mutex sync.RWMutex
}

func NewPeers() *Peers {
	return &Peers{list: make(map[peer.ID]*Peer)}
}

func (p *Peers) Add(id peer.ID, conn network.Conn) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if _, ok := p.list[id]; ok {
		return
	}
	p.list[id] = &Peer{id, conn, false}
}

func (p *Peers) Remove(id peer.ID) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	delete(p.list, id)
}

func (p *Peers) UpdateConnected(id peer.ID) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if _, ok := p.list[id]; ok {
		p.list[id].Connected = true
	}
}

func (p *Peers) Count() int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return len(p.list)
}

func (p *Peers) All() []*Peer {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	list := []*Peer{}
	for _, p := range p.list {
		list = append(list, p)
	}
	return list
}

func (p *Peers) UnConnected() []*Peer {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	list := []*Peer{}
	for _, p := range p.list {
		if !p.Connected {
			list = append(list, p)
		}
	}
	return list
}