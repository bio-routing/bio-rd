package server

import (
	"github.com/bio-routing/bio-rd/protocols/bgp/metrics"
)

type metricsService struct {
	server *bgpServer
}

func (b *metricsService) metrics() *metrics.BGPMetrics {
	return &metrics.BGPMetrics{
		Neighbors: b.peerMetrics(),
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
		UpdatesReceived: 0,
		UpdatesSent:     0,
	}

	fsm := peer.fsms[0]

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
		RoutesReceived: family.adjRIBIn.RouteCount(),
		RoutesSent:     family.adjRIBOut.RouteCount(),
	}
}
