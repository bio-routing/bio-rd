package server

import (
	"sync"

	bnet "github.com/bio-routing/bio-rd/net"
)

type peerManager struct {
	peers   map[bnet.IP]*peer
	peersMu sync.RWMutex
}

func newPeerManager() *peerManager {
	return &peerManager{
		peers: make(map[bnet.IP]*peer),
	}
}

func (m *peerManager) add(p *peer) {
	m.peersMu.Lock()
	defer m.peersMu.Unlock()

	m.peers[*p.GetAddr()] = p
}

func (m *peerManager) remove(neighborIP *bnet.IP) {
	m.peersMu.Lock()
	defer m.peersMu.Unlock()

	delete(m.peers, *neighborIP)
}

func (m *peerManager) get(neighborIP *bnet.IP) *peer {
	m.peersMu.RLock()
	defer m.peersMu.RUnlock()

	p, _ := m.peers[*neighborIP]
	return p
}

func (m *peerManager) list() []*peer {
	m.peersMu.RLock()
	defer m.peersMu.RUnlock()

	res := make([]*peer, len(m.peers))
	i := 0
	for _, p := range m.peers {
		res[i] = p
		i++
	}

	return res
}
