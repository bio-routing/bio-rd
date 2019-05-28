package server

import (
	"testing"
	"time"

	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
	"github.com/stretchr/testify/assert"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/metrics"
)

func TestMetrics(t *testing.T) {
	vrf, _ := vrf.New("inet.0", 0)
	establishedTime := time.Now()

	tests := []struct {
		name               string
		peer               *peer
		withoutFSM         bool
		state              state
		updatesReceived    uint64
		updatesSent        uint64
		ipv4RoutesReceived int64
		ipv4RoutesSent     int64
		ipv6RoutesReceived int64
		ipv6RoutesSent     int64
		expected           *metrics.BGPMetrics
	}{
		{
			name: "Established",
			peer: &peer{
				peerASN:  202739,
				localASN: 201701,
				addr:     bnet.IPv4(100),
				ipv4:     &peerAddressFamily{},
				ipv6:     &peerAddressFamily{},
				vrf:      vrf,
			},
			state:              &establishedState{},
			updatesReceived:    3,
			updatesSent:        4,
			ipv4RoutesReceived: 5,
			ipv4RoutesSent:     6,
			ipv6RoutesReceived: 7,
			ipv6RoutesSent:     8,
			expected: &metrics.BGPMetrics{
				Peers: []*metrics.BGPPeerMetrics{
					{
						IP:              bnet.IPv4(100),
						ASN:             202739,
						LocalASN:        201701,
						UpdatesReceived: 3,
						UpdatesSent:     4,
						VRF:             "inet.0",
						Up:              true,
						State:           metrics.StateEstablished,
						Since:           establishedTime,
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
			},
		},
		{
			name: "Idle",
			peer: &peer{
				peerASN:  202739,
				localASN: 201701,
				addr:     bnet.IPv4(100),
				ipv4:     &peerAddressFamily{},
				ipv6:     &peerAddressFamily{},
				vrf:      vrf,
			},
			state: &idleState{},
			expected: &metrics.BGPMetrics{
				Peers: []*metrics.BGPPeerMetrics{
					{
						IP:       bnet.IPv4(100),
						ASN:      202739,
						LocalASN: 201701,
						VRF:      "inet.0",
						State:    metrics.StateIdle,
						AddressFamilies: []*metrics.BGPAddressFamilyMetrics{
							{
								AFI:  packet.IPv4AFI,
								SAFI: packet.UnicastSAFI,
							},
							{
								AFI:  packet.IPv6AFI,
								SAFI: packet.UnicastSAFI,
							},
						},
					},
				},
			},
		},
		{
			name: "Active",
			peer: &peer{
				peerASN:  202739,
				localASN: 201701,
				addr:     bnet.IPv4(100),
				ipv4:     &peerAddressFamily{},
				ipv6:     &peerAddressFamily{},
				vrf:      vrf,
			},
			state: &activeState{},
			expected: &metrics.BGPMetrics{
				Peers: []*metrics.BGPPeerMetrics{
					{
						IP:       bnet.IPv4(100),
						ASN:      202739,
						LocalASN: 201701,
						VRF:      "inet.0",
						State:    metrics.StateActive,
						AddressFamilies: []*metrics.BGPAddressFamilyMetrics{
							{
								AFI:  packet.IPv4AFI,
								SAFI: packet.UnicastSAFI,
							},
							{
								AFI:  packet.IPv6AFI,
								SAFI: packet.UnicastSAFI,
							},
						},
					},
				},
			},
		},
		{
			name: "OpenSent",
			peer: &peer{
				peerASN:  202739,
				localASN: 201701,
				addr:     bnet.IPv4(100),
				ipv4:     &peerAddressFamily{},
				ipv6:     &peerAddressFamily{},
				vrf:      vrf,
			},
			state: &openSentState{},
			expected: &metrics.BGPMetrics{
				Peers: []*metrics.BGPPeerMetrics{
					{
						IP:       bnet.IPv4(100),
						ASN:      202739,
						LocalASN: 201701,
						VRF:      "inet.0",
						State:    metrics.StateOpenSent,
						AddressFamilies: []*metrics.BGPAddressFamilyMetrics{
							{
								AFI:  packet.IPv4AFI,
								SAFI: packet.UnicastSAFI,
							},
							{
								AFI:  packet.IPv6AFI,
								SAFI: packet.UnicastSAFI,
							},
						},
					},
				},
			},
		},
		{
			name: "OpenConfirm",
			peer: &peer{
				peerASN:  202739,
				localASN: 201701,
				addr:     bnet.IPv4(100),
				ipv4:     &peerAddressFamily{},
				ipv6:     &peerAddressFamily{},
				vrf:      vrf,
			},
			state: &openConfirmState{},
			expected: &metrics.BGPMetrics{
				Peers: []*metrics.BGPPeerMetrics{
					{
						IP:       bnet.IPv4(100),
						ASN:      202739,
						LocalASN: 201701,
						VRF:      "inet.0",
						State:    metrics.StateOpenConfirm,
						AddressFamilies: []*metrics.BGPAddressFamilyMetrics{
							{
								AFI:  packet.IPv4AFI,
								SAFI: packet.UnicastSAFI,
							},
							{
								AFI:  packet.IPv6AFI,
								SAFI: packet.UnicastSAFI,
							},
						},
					},
				},
			},
		},
		{
			name: "Connect",
			peer: &peer{
				peerASN:  202739,
				localASN: 201701,
				addr:     bnet.IPv4(100),
				ipv4:     &peerAddressFamily{},
				ipv6:     &peerAddressFamily{},
				vrf:      vrf,
			},
			state: &connectState{},
			expected: &metrics.BGPMetrics{
				Peers: []*metrics.BGPPeerMetrics{
					{
						IP:       bnet.IPv4(100),
						ASN:      202739,
						LocalASN: 201701,
						VRF:      "inet.0",
						State:    metrics.StateConnect,
						AddressFamilies: []*metrics.BGPAddressFamilyMetrics{
							{
								AFI:  packet.IPv4AFI,
								SAFI: packet.UnicastSAFI,
							},
							{
								AFI:  packet.IPv6AFI,
								SAFI: packet.UnicastSAFI,
							},
						},
					},
				},
			},
		},
		{
			name:       "without fsm",
			withoutFSM: true,
			peer: &peer{
				peerASN:  202739,
				localASN: 201701,
				addr:     bnet.IPv4(100),
				ipv4:     &peerAddressFamily{},
				ipv6:     &peerAddressFamily{},
				vrf:      vrf,
			},
			expected: &metrics.BGPMetrics{
				Peers: []*metrics.BGPPeerMetrics{
					{
						IP:              bnet.IPv4(100),
						ASN:             202739,
						LocalASN:        201701,
						VRF:             "inet.0",
						AddressFamilies: []*metrics.BGPAddressFamilyMetrics{},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !test.withoutFSM {
				fsm := newFSM(test.peer)
				test.peer.fsms = append(test.peer.fsms, fsm)

				fsm.state = test.state
				fsm.counters.updatesReceived = test.updatesReceived
				fsm.counters.updatesSent = test.updatesSent

				fsm.ipv4Unicast.adjRIBIn = &routingtable.RTMockClient{FakeRouteCount: test.ipv4RoutesReceived}
				fsm.ipv4Unicast.adjRIBOut = &routingtable.RTMockClient{FakeRouteCount: test.ipv4RoutesSent}
				fsm.ipv6Unicast.adjRIBIn = &routingtable.RTMockClient{FakeRouteCount: test.ipv6RoutesReceived}
				fsm.ipv6Unicast.adjRIBOut = &routingtable.RTMockClient{FakeRouteCount: test.ipv6RoutesSent}

				fsm.establishedTime = establishedTime
			}

			s := newBgpServer()
			s.peers.add(test.peer)

			actual, err := s.Metrics()
			if err != nil {
				t.Fatalf("unecpected error: %v", err)
			}

			assert.Equal(t, test.expected, actual)
		})
	}
}
