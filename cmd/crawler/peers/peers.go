package peers

import (
	"github.com/Qitmeer/qitmeer/cmd/crawler/config"
	"github.com/Qitmeer/qitmeer/cmd/crawler/db"
	"github.com/Qitmeer/qitmeer/cmd/crawler/rpc"
	"github.com/Qitmeer/qitmeer/core/json"
	"sync"
	"time"
)

const (
	find_peer_interval = 60 * 60 * 1
)

type Peers struct {
	list  map[string]*db.Peer
	mutex sync.RWMutex
	db    *db.PeerDB
}

func NewPeers() (*Peers, error) {
	storage, err := db.OpenPeerDB(config.DefaultDB)
	if err != nil {
		return nil, err
	}
	return &Peers{list: make(map[string]*db.Peer), db: storage}, nil
}

func (p *Peers) Start() error {
	p.LoadPeers()
	return nil
}

func (p *Peers) Stop() error {
	return p.db.Close()
}

func (p *Peers) APIs() []rpc.API {
	return []rpc.API{
		{
			NameSpace: "crawler",
			Service:   NewPeersApi(p),
			Public:    true,
		},
	}
}

func (p *Peers) LoadPeers() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	peerList := p.db.PeerList()
	for _, peerInfo := range peerList {
		p.list[peerInfo.Id] = peerInfo
	}
}

func (p *Peers) Add(id string, Addr string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if _, ok := p.list[id]; ok {
		return
	}
	peer := &db.Peer{Id: id, Addr: Addr, ConnectTime: 0, Connected: true}
	p.list[id] = peer
	p.db.UpdatePeer(peer)
}

func (p *Peers) UpdateUnConnected(id string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if peer, ok := p.list[id]; !ok {
		return
	} else {
		peer.Connected = false
		p.db.UpdatePeer(peer)
	}
}

func (p *Peers) UpdateConnectTime(id string, timestamp int64) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if peer, ok := p.list[id]; !ok {
		return
	} else {
		peer.ConnectTime = timestamp
		p.db.UpdatePeer(peer)
	}
}

func (p *Peers) Count() int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return len(p.list)
}

func (p *Peers) All() []*db.Peer {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	list := []*db.Peer{}
	for _, p := range p.list {
		list = append(list, p)
	}
	return list
}

func (p *Peers) FindPeerList() []*db.Peer {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	list := []*db.Peer{}
	now := time.Now().Unix()
	for _, p := range p.list {
		if now-p.ConnectTime >= find_peer_interval {
			list = append(list, p)
		}
	}
	return list
}

type PeersApi struct {
	peers *Peers
}

func NewPeersApi(peers *Peers) *PeersApi {
	return &PeersApi{peers: peers}
}

func (api *PeersApi) GetNodeList() (interface{}, error) {
	rs := []json.OrderedResult{}

	list := api.peers.All()
	for _, node := range list {
		rs = append(rs, json.OrderedResult{
			{Key: "id", Val: node.Id},
			{Key: "ip", Val: node.Addr},
			{Key: "connecttime", Val: time.Unix(node.ConnectTime, 0).String()},
			{Key: "connected", Val: node.Connected},
		})
	}
	return rs, nil
}
