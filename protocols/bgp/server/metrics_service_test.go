package server

import (
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/stretchr/testify/assert"
	"testing"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/metrics"
)

func TestMetrics(t *testing.T) {
	expected := &metrics.BGPMetrics{
		Peers: []*metrics.BGPPeerMetrics{
			{
				IP:              bnet.IPv4(100),
				ASN:             202739,
				LocalASN:        201701,
				UpdatesReceived: 3,
				UpdatesSent:     4,
				AddressFamilies: []*metrics.BGPAddressFamilyMetrics{
					{
						AFI:            packet.IPv4AFI,
						SAFI:           packet.UnicastSAFI,
						RoutesReceived: 5,
						RoutesSent:     6,
					},
					{
						AFI:            packet.IPv6AFI,
						SAFI:           packet.UnicastSAFI,
						RoutesReceived: 7,
						RoutesSent:     8,
					},
				},
			},
		},
	}

	p := &peer{
		peerASN:  202739,
		localASN: 201701,
		addr:     bnet.IPv4(100),
		ipv4:     &peerAddressFamily{},
		ipv6:     &peerAddressFamily{},
	}
	fsm := newFSM(p)
	p.fsms = append(p.fsms, fsm)
	fsm.counters.updatesReceived = 3
	fsm.counters.updatesSent = 4

	fsm.ipv4Unicast.adjRIBIn = &routingtable.RTMockClient{FakeRouteCount: 5}
	fsm.ipv4Unicast.adjRIBOut = &routingtable.RTMockClient{FakeRouteCount: 6}
	fsm.ipv6Unicast.adjRIBIn = &routingtable.RTMockClient{FakeRouteCount: 7}
	fsm.ipv6Unicast.adjRIBOut = &routingtable.RTMockClient{FakeRouteCount: 8}

	s := newBgpServer()
	s.peers.add(p)

	actual, err := s.Metrics()
	if err != nil {
		t.Fatalf("unecpected error: %v", err)
	}

	assert.Equal(t, expected, actual)
}
