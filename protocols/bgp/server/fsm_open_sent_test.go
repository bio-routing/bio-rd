package server

import (
	"github.com/bio-routing/bio-rd/routingtable"
	"net"
	"testing"

	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	btesting "github.com/bio-routing/bio-rd/testing"
	"github.com/stretchr/testify/assert"
)

func TestOpenMsgReceived(t *testing.T) {
	tests := []struct {
		asn      uint32
		name     string
		msg      packet.BGPOpen
		wantIdle bool
	}{
		{
			name: "valid open message (16bit ASN)",
			asn:  12345,
			msg: packet.BGPOpen{
				HoldTime:      90,
				BGPIdentifier: 1,
				Version:       4,
				ASN:           12345,
			},
		},
		{
			name: "valid open message (32bit ASN)",
			asn:  202739,
			msg: packet.BGPOpen{
				HoldTime:      90,
				BGPIdentifier: 1,
				Version:       4,
				ASN:           23456,
				OptParmLen:    1,
				OptParams: []packet.OptParam{
					{
						Type:   packet.CapabilitiesParamType,
						Length: 6,
						Value: packet.Capabilities{
							packet.Capability{
								Code:   packet.ASN4CapabilityCode,
								Length: 4,
								Value: packet.ASN4Capability{
									ASN4: 202739,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "open message does not match configured remote ASN",
			asn:  12345,
			msg: packet.BGPOpen{
				HoldTime:      90,
				BGPIdentifier: 1,
				Version:       4,
				ASN:           54321,
			},
			wantIdle: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fsm := newFSM(&peer{
				peerASN: test.asn,
			})

			conA, conB := net.Pipe()
			fsm.con = conB

			go func() {
				for {
					buf := make([]byte, 1)
					_, err := conA.Read(buf)
					if err != nil {
						return
					}
				}
			}()

			s := &openSentState{
				fsm: fsm,
			}

			state, _ := s.handleOpenMessage(&test.msg)

			if test.wantIdle {
				assert.IsType(t, &idleState{}, state, "state")
				return
			}

			assert.IsType(t, &openConfirmState{}, state, "state")
			assert.Equal(t, test.asn, s.peerASNRcvd, "asn")
		})
	}
}

func TestProcessMultiProtocolCapability(t *testing.T) {
	tests := []struct {
		name                    string
		peer                    *peer
		caps                    []packet.MultiProtocolCapability
		expectIPv4MultiProtocol bool
		expectIPv6MultiProtocol bool
	}{
		{
			name: "IPv4 only without multi protocol configuration",
			peer: &peer{
				ipv4: &peerAddressFamily{},
			},
			caps: []packet.MultiProtocolCapability{
				{
					AFI:  packet.IPv4AFI,
					SAFI: packet.UnicastSAFI,
				},
			},
		},
		{
			name: "IPv4 only with multi protocol configuration",
			peer: &peer{
				ipv4:                        &peerAddressFamily{},
				ipv4MultiProtocolAdvertised: true,
			},
			caps: []packet.MultiProtocolCapability{
				{
					AFI:  packet.IPv4AFI,
					SAFI: packet.UnicastSAFI,
				},
			},
			expectIPv4MultiProtocol: true,
		},
		{
			name: "IPv6 only",
			peer: &peer{
				ipv6: &peerAddressFamily{},
			},
			caps: []packet.MultiProtocolCapability{
				{
					AFI:  packet.IPv6AFI,
					SAFI: packet.UnicastSAFI,
				},
			},
			expectIPv6MultiProtocol: true,
		},
		{
			name: "IPv4 and IPv6, only IPv6 configured as multi protocol",
			peer: &peer{
				ipv4: &peerAddressFamily{},
				ipv6: &peerAddressFamily{},
			},
			caps: []packet.MultiProtocolCapability{
				{
					AFI:  packet.IPv6AFI,
					SAFI: packet.UnicastSAFI,
				},
				{
					AFI:  packet.IPv4AFI,
					SAFI: packet.UnicastSAFI,
				},
			},
			expectIPv6MultiProtocol: true,
		},
		{
			name: "IPv4 and IPv6 configured as multi protocol",
			peer: &peer{
				ipv4:                        &peerAddressFamily{},
				ipv6:                        &peerAddressFamily{},
				ipv4MultiProtocolAdvertised: true,
			},
			caps: []packet.MultiProtocolCapability{
				{
					AFI:  packet.IPv6AFI,
					SAFI: packet.UnicastSAFI,
				},
				{
					AFI:  packet.IPv4AFI,
					SAFI: packet.UnicastSAFI,
				},
			},
			expectIPv4MultiProtocol: true,
			expectIPv6MultiProtocol: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fsm := newFSM(test.peer)
			fsm.con = &btesting.MockConn{}

			s := &openSentState{
				fsm: fsm,
			}

			for _, cap := range test.caps {
				s.processMultiProtocolCapability(cap)
			}

			if fsm.ipv4Unicast != nil {
				assert.Equal(t, test.expectIPv4MultiProtocol, fsm.ipv4Unicast.multiProtocol)
			}

			if fsm.ipv6Unicast != nil {
				assert.Equal(t, test.expectIPv6MultiProtocol, fsm.ipv6Unicast.multiProtocol)
			}
		})
	}
}

func TestProcessAddPathCapabilityTX(t *testing.T) {
	tests := []struct {
		name     string
		peer     *peer
		caps     []packet.AddPathCapability
		expected routingtable.ClientOptions
	}{
		{
			name: "Add-Path enabled and cap received",
			peer: &peer{
				ipv4: &peerAddressFamily{
					addPathSend: routingtable.ClientOptions{MaxPaths: 3},
				},
			},
			caps: []packet.AddPathCapability{
				{
					AFI:         packet.IPv4AFI,
					SAFI:        packet.UnicastSAFI,
					SendReceive: packet.AddPathReceive,
				},
			},
			expected: routingtable.ClientOptions{MaxPaths: 3},
		},
		{
			name: "Add-Path enabled and cap not received",
			peer: &peer{
				ipv4: &peerAddressFamily{
					addPathSend: routingtable.ClientOptions{MaxPaths: 3},
				},
			},
			caps:     []packet.AddPathCapability{},
			expected: routingtable.ClientOptions{BestOnly: true},
		},
		{
			name: "Add-Path disabled and cap received",
			peer: &peer{
				ipv4: &peerAddressFamily{
					addPathSend: routingtable.ClientOptions{BestOnly: true},
				},
			},
			caps: []packet.AddPathCapability{
				{
					AFI:         packet.IPv4AFI,
					SAFI:        packet.UnicastSAFI,
					SendReceive: packet.AddPathReceive,
				},
			},
			expected: routingtable.ClientOptions{BestOnly: true},
		},
		{
			name: "Add-Path disabled and cap not received",
			peer: &peer{
				ipv4: &peerAddressFamily{
					addPathSend: routingtable.ClientOptions{BestOnly: true},
				},
			},
			caps:     []packet.AddPathCapability{},
			expected: routingtable.ClientOptions{BestOnly: true},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fsm := newFSM(test.peer)
			fsm.con = &btesting.MockConn{}

			s := &openSentState{
				fsm: fsm,
			}

			for _, cap := range test.caps {
				s.processAddPathCapability(cap)
			}

			assert.Equal(t, test.expected, fsm.ipv4Unicast.addPathTX)
		})
	}
}
