package server

import (
	"github.com/bio-routing/bio-rd/protocols/bgp/metrics"
)

type metricsService struct {
	server *bgpServer
}

func (b *metricsService) metrics() *metrics.BGPMetrics {
	return &metrics.BGPMetrics{
		Peers: b.peerMetrics(),
	}
}

func (b *metricsService) peerMetrics() []*metrics.BGPPeerMetrics {
	peers := make([]*metrics.BGPPeerMetrics, 0)

	for _, peer := range b.server.peers.list() {
		m := metricsForPeer(peer)
		peers = append(peers, m)
	}

	return peers
}

func metricsForPeer(peer *peer) *metrics.BGPPeerMetrics {
	m := &metrics.BGPPeerMetrics{
		ASN:             peer.peerASN,
		LocalASN:        peer.localASN,
		IP:              peer.addr,
		AddressFamilies: make([]*metrics.BGPAddressFamilyMetrics, 0),
		VRF:             peer.vrf.Name(),
	}

	var fsms = peer.fsms
	if len(fsms) == 0 {
		return m
	}

	fsm := fsms[0]
	m.State = statusFromFSM(fsm)

	if m.State == metrics.StateEstablished {
		m.Since = fsm.establishedTime
	}

	m.UpdatesReceived = fsm.counters.updatesReceived
	m.UpdatesSent = fsm.counters.updatesSent

	fsm.stateMu.RLock()
	defer fsm.stateMu.RUnlock()

	if fsm.ribsInitialized {
		if peer.ipv4 != nil {
			m.AddressFamilies = append(m.AddressFamilies, metricsForFamily(fsm.ipv4Unicast))
		}

		if peer.ipv6 != nil {
			m.AddressFamilies = append(m.AddressFamilies, metricsForFamily(fsm.ipv6Unicast))
		}
	}

	return m
}

func metricsForFamily(family *fsmAddressFamily) *metrics.BGPAddressFamilyMetrics {
	m := &metrics.BGPAddressFamilyMetrics{
		AFI:                    family.afi,
		SAFI:                   family.safi,
		RoutesReceived:         uint64(family.adjRIBIn.RouteCount()),
		EndOfRIBMarkerReceived: family.endOfRIBMarkerReceived.Load(),
	}

	if family.adjRIBOut != nil {
		m.RoutesSent = uint64(family.adjRIBOut.RouteCount())
	}

	return m
}

func statusFromFSM(fsm *FSM) uint8 {
	switch fsm.state.(type) {
	case *idleState:
		return metrics.StateIdle
	case *connectState:
		return metrics.StateConnect
	case *activeState:
		return metrics.StateActive
	case *openSentState:
		return metrics.StateOpenSent
	case *openConfirmState:
		return metrics.StateOpenConfirm
	case *establishedState:
		return metrics.StateEstablished
	}

	return metrics.StateDown
}
