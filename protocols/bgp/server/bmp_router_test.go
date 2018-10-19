package server

import (
	"testing"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	bmppkt "github.com/bio-routing/bio-rd/protocols/bmp/packet"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/adjRIBIn"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	"github.com/stretchr/testify/assert"
)

func TestConfigureBySentOpen(t *testing.T) {
	tests := []struct {
		name     string
		p        *peer
		openMsg  *packet.BGPOpen
		expected *peer
	}{
		{
			name: "Test 32bit ASN",
			p:    &peer{},
			openMsg: &packet.BGPOpen{
				OptParams: []packet.OptParam{
					{
						Type: 2,
						Value: packet.Capabilities{
							{
								Code: packet.ASN4CapabilityCode,
								Value: packet.ASN4Capability{
									ASN4: 201701,
								},
							},
						},
					},
				},
			},
			expected: &peer{
				localASN: 201701,
			},
		},
		{
			name: "Test Add Path TX",
			p: &peer{
				ipv4: &peerAddressFamily{},
			},
			openMsg: &packet.BGPOpen{
				OptParams: []packet.OptParam{
					{
						Type: 2,
						Value: packet.Capabilities{
							{
								Code: packet.AddPathCapabilityCode,
								Value: packet.AddPathCapability{
									AFI:         1,
									SAFI:        1,
									SendReceive: 2,
								},
							},
						},
					},
				},
			},
			expected: &peer{
				ipv4: &peerAddressFamily{
					addPathSend: routingtable.ClientOptions{
						MaxPaths: 10,
					},
				},
			},
		},
		{
			name: "Test Add Path RX",
			p: &peer{
				ipv4: &peerAddressFamily{},
			},
			openMsg: &packet.BGPOpen{
				OptParams: []packet.OptParam{
					{
						Type: 2,
						Value: packet.Capabilities{
							{
								Code: packet.AddPathCapabilityCode,
								Value: packet.AddPathCapability{
									AFI:         1,
									SAFI:        1,
									SendReceive: 1,
								},
							},
						},
					},
				},
			},
			expected: &peer{
				ipv4: &peerAddressFamily{
					addPathReceive: true,
				},
			},
		},
		{
			name: "Test Add Path RX/TX",
			p: &peer{
				ipv4: &peerAddressFamily{},
			},
			openMsg: &packet.BGPOpen{
				OptParams: []packet.OptParam{
					{
						Type: 2,
						Value: packet.Capabilities{
							{
								Code: packet.AddPathCapabilityCode,
								Value: packet.AddPathCapability{
									AFI:         1,
									SAFI:        1,
									SendReceive: 3,
								},
							},
						},
					},
				},
			},
			expected: &peer{
				ipv4: &peerAddressFamily{
					addPathSend: routingtable.ClientOptions{
						MaxPaths: 10,
					},
					addPathReceive: true,
				},
			},
		},
	}

	for _, test := range tests {
		test.p.configureBySentOpen(test.openMsg)

		assert.Equalf(t, test.expected, test.p, "Test %q", test.name)
	}
}

func TestProcessPeerUpNotification(t *testing.T) {
	tests := []struct {
		name     string
		router   *router
		pkt      *bmppkt.PeerUpNotification
		wantFail bool
		expected *router
	}{
		{
			name: "Invalid sent open message",
			router: &router{
				neighbors: make(map[[16]byte]*neighbor),
			},
			pkt: &bmppkt.PeerUpNotification{
				PerPeerHeader: &bmppkt.PerPeerHeader{
					PeerType:          0,
					PeerFlags:         0,
					PeerDistinguisher: 0,
					PeerAddress:       [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 0, 255, 1},
					PeerAS:            51324,
					PeerBGPID:         100,
				},
				LocalAddress:    [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 0, 255, 0},
				LocalPort:       179,
				RemotePort:      34542,
				SentOpenMsg:     []byte{},
				ReceivedOpenMsg: []byte{},
				Information:     []byte{},
			},
			wantFail: true,
		},
		{
			name: "Invalid received open message",
			router: &router{
				neighbors: make(map[[16]byte]*neighbor),
			},
			pkt: &bmppkt.PeerUpNotification{
				PerPeerHeader: &bmppkt.PerPeerHeader{
					PeerType:          0,
					PeerFlags:         0,
					PeerDistinguisher: 0,
					PeerAddress:       [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 0, 255, 1},
					PeerAS:            51324,
					PeerBGPID:         100,
				},
				LocalAddress: [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 0, 255, 0},
				LocalPort:    179,
				RemotePort:   34542,
				SentOpenMsg: []byte{
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					0, 29,
					1,
					4,
					100, 200,
					20, 0,
					10, 20, 30, 40,
					0,
				},
				ReceivedOpenMsg: []byte{},
				Information:     []byte{},
			},
			wantFail: true,
		},
		{
			name: "Regular BGP by RFC4271",
			router: &router{
				rib4:      locRIB.New(),
				rib6:      locRIB.New(),
				neighbors: make(map[[16]byte]*neighbor),
			},
			pkt: &bmppkt.PeerUpNotification{
				PerPeerHeader: &bmppkt.PerPeerHeader{
					PeerType:          0,
					PeerFlags:         0,
					PeerDistinguisher: 0,
					PeerAddress:       [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 0, 255, 1},
					PeerAS:            100,
					PeerBGPID:         100,
				},
				LocalAddress: [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 0, 255, 0},
				LocalPort:    179,
				RemotePort:   34542,
				SentOpenMsg: []byte{
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					0, 29,
					1,
					4,
					0, 200,
					20, 0,
					10, 20, 30, 40,
					0,
				},
				ReceivedOpenMsg: []byte{
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					0, 29,
					1,
					4,
					0, 100,
					20, 0,
					10, 20, 30, 50,
					0,
				},
				Information: []byte{},
			},
			wantFail: false,
			expected: &router{
				rib4: locRIB.New(),
				rib6: locRIB.New(),
				neighbors: map[[16]byte]*neighbor{
					[16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 0, 255, 1}: {
						localAS:  200,
						peerAS:   100,
						address:  [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 0, 255, 1},
						routerID: 169090610,
						opt: &packet.DecodeOptions{
							AddPath:     false,
							Use32BitASN: false,
						},
						fsm: &FSM{
							isBMP:      true,
							neighborID: 169090610,
							state:      &establishedState{},
							peer: &peer{
								addr:      bnet.IPv4FromOctets(10, 0, 255, 1),
								localAddr: bnet.IPv4FromOctets(10, 0, 255, 0),
								peerASN:   100,
								localASN:  200,
								ipv4:      &peerAddressFamily{},
								ipv6:      &peerAddressFamily{},
							},
							ipv4Unicast: &fsmAddressFamily{
								afi:          1,
								safi:         1,
								adjRIBIn:     adjRIBIn.New(filter.NewAcceptAllFilter(), &routingtable.ContributingASNs{}, 0, 0, false),
								importFilter: filter.NewAcceptAllFilter(),
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		err := test.router.processPeerUpNotification(test.pkt)
		if err != nil {
			if test.wantFail {
				continue
			}

			t.Errorf("Unexpected failure for test %q: %v", test.name, err)
			continue
		}

		if test.wantFail {
			t.Errorf("Unexpected success for test %q", test.name)
			continue
		}

		test.expected.neighbors[test.pkt.PerPeerHeader.PeerAddress].fsm.state = &establishedState{fsm: test.expected.neighbors[test.pkt.PerPeerHeader.PeerAddress].fsm}

		if test.expected.neighbors[test.pkt.PerPeerHeader.PeerAddress].fsm.ipv4Unicast != nil {
			test.expected.neighbors[test.pkt.PerPeerHeader.PeerAddress].fsm.ipv4Unicast.rib = test.router.rib4
			test.expected.neighbors[test.pkt.PerPeerHeader.PeerAddress].fsm.ipv4Unicast.fsm = test.expected.neighbors[test.pkt.PerPeerHeader.PeerAddress].fsm
			test.expected.neighbors[test.pkt.PerPeerHeader.PeerAddress].fsm.ipv4Unicast.adjRIBIn.Register(test.router.rib4)
		}

		if test.expected.neighbors[test.pkt.PerPeerHeader.PeerAddress].fsm.ipv6Unicast != nil {
			test.expected.neighbors[test.pkt.PerPeerHeader.PeerAddress].fsm.ipv6Unicast.rib = test.router.rib6
			test.expected.neighbors[test.pkt.PerPeerHeader.PeerAddress].fsm.ipv6Unicast.fsm = test.expected.neighbors[test.pkt.PerPeerHeader.PeerAddress].fsm
			test.expected.neighbors[test.pkt.PerPeerHeader.PeerAddress].fsm.ipv6Unicast.adjRIBIn.Register(test.router.rib6)
		}

		assert.Equalf(t, test.expected, test.router, "Test %q", test.name)
	}

}
