package server

import (
	"time"

	"github.com/bio-routing/bio-rd/protocols/bgp/metrics"
)

const (
	statusDown        = 0
	statusIdle        = 1
	statusConnect     = 2
	statusActive      = 3
	statusOpenSent    = 4
	statusOpenConfirm = 5
	statusEstablished = 6
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
		if len(peer.fsms) == 0 {
			continue
		}

		m := b.metricsForPeer(peer)
		peers = append(peers, m)
	}

	return peers
}

func (b *metricsService) metricsForPeer(peer *peer) *metrics.BGPPeerMetrics {
	m := &metrics.BGPPeerMetrics{
		ASN:             peer.peerASN,
		LocalASN:        peer.localASN,
		IP:              peer.addr,
		AddressFamilies: make([]*metrics.BGPAddressFamilyMetrics, 0),
	}

	fsm := peer.fsms[0]
	m.Status = b.statusFromFSM(fsm)
	m.Up = m.Status == statusEstablished

	if m.Up {
		m.Since = time.Since(fsm.establishedDate)
	}

	m.UpdatesReceived = fsm.counters.updatesReceived
	m.UpdatesSent = fsm.counters.updatesSent

	if peer.ipv4 != nil {
		m.AddressFamilies = append(m.AddressFamilies, b.metricsForFamily(fsm.ipv4Unicast))
	}

	if peer.ipv6 != nil {
		m.AddressFamilies = append(m.AddressFamilies, b.metricsForFamily(fsm.ipv6Unicast))
	}

	return m
}

func (b *metricsService) metricsForFamily(family *fsmAddressFamily) *metrics.BGPAddressFamilyMetrics {
	return &metrics.BGPAddressFamilyMetrics{
		AFI:            family.afi,
		SAFI:           family.safi,
		RoutesReceived: uint64(family.adjRIBIn.RouteCount()),
		RoutesSent:     uint64(family.adjRIBOut.RouteCount()),
	}
}

func (b *metricsService) statusFromFSM(fsm *FSM) uint8 {
	switch fsm.state.(type) {
	case *idleState:
		return statusIdle
	case *connectState:
		return statusConnect
	case *activeState:
		return statusActive
	case *openSentState:
		return statusOpenSent
	case *openConfirmState:
		return statusOpenConfirm
	case *establishedState:
		return statusEstablished
	}

	return statusDown
}
