package server

import (
	"sync"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
)

type PeerKey struct {
	vrf        *vrf.VRF
	neighborIP *bnet.IP
}

func (pk PeerKey) VRF() *vrf.VRF {
	return pk.vrf
}

func (pk PeerKey) Addr() *bnet.IP {
	return pk.neighborIP
}

type peerManager struct {
	peers   map[PeerKey]*peer
	peersMu sync.RWMutex
}

func newPeerManager() *peerManager {
	return &peerManager{
		peers: make(map[PeerKey]*peer),
	}
}

func (m *peerManager) add(p *peer) {
	m.peersMu.Lock()
	defer m.peersMu.Unlock()

	m.peers[p.peerKey()] = p
}

func (m *peerManager) remove(pk PeerKey) {
	m.peersMu.Lock()
	defer m.peersMu.Unlock()

	delete(m.peers, pk)
}

func (m *peerManager) get(vrf *vrf.VRF, neighborIP *bnet.IP) *peer {
	m.peersMu.RLock()
	defer m.peersMu.RUnlock()

	p := m.peers[PeerKey{
		vrf:        vrf,
		neighborIP: neighborIP.Dedup(),
	}]
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
