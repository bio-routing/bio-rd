package server

import (
	"net"
	"testing"

	"github.com/bio-routing/bio-rd/routingtable"

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
					AFI:  packet.AFIIPv4,
					SAFI: packet.SAFIUnicast,
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
					AFI:  packet.AFIIPv4,
					SAFI: packet.SAFIUnicast,
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
					AFI:  packet.AFIIPv6,
					SAFI: packet.SAFIUnicast,
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
					AFI:  packet.AFIIPv6,
					SAFI: packet.SAFIUnicast,
				},
				{
					AFI:  packet.AFIIPv4,
					SAFI: packet.SAFIUnicast,
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
					AFI:  packet.AFIIPv6,
					SAFI: packet.SAFIUnicast,
				},
				{
					AFI:  packet.AFIIPv4,
					SAFI: packet.SAFIUnicast,
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
					packet.AddPathCapabilityTuple{
						AFI:         packet.AFIIPv4,
						SAFI:        packet.SAFIUnicast,
						SendReceive: packet.AddPathReceive,
					},
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
					packet.AddPathCapabilityTuple{
						AFI:         packet.AFIIPv4,
						SAFI:        packet.SAFIUnicast,
						SendReceive: packet.AddPathReceive,
					},
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

func TestProcessPeerRoleCapability(t *testing.T) {
	tests := []struct {
		name               string
		msg                packet.BGPOpen
		peerRoleEnabled    bool
		peerRoleStrictMode bool
		peerRoleLocalRole  uint8
		wantIdle           bool
		errmsg             string
	}{
		// mode: off
		{
			name:            "peer role mode off, no peer role received",
			peerRoleEnabled: false,
			msg: packet.BGPOpen{
				HoldTime:      90,
				BGPIdentifier: 1,
				Version:       4,
				ASN:           23456,
				OptParmLen:    0,
				OptParams:     []packet.OptParam{},
			},
			wantIdle: false,
		},
		{
			name:            "peer role mode off, peer role received",
			peerRoleEnabled: false,
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
								Code:   packet.PeerRoleCapabilityCode,
								Length: 4,
								Value: packet.PeerRoleCapability{
									PeerRole: packet.PeerRoleRolePeer,
								},
							},
						},
					},
				},
			},
			wantIdle: false,
		},

		// mode allow
		{
			name:               "peer role mode allow, no peer role received",
			peerRoleEnabled:    true,
			peerRoleStrictMode: false,
			msg: packet.BGPOpen{
				HoldTime:      90,
				BGPIdentifier: 1,
				Version:       4,
				ASN:           23456,
				OptParmLen:    0,
				OptParams:     []packet.OptParam{},
			},
			wantIdle: false,
		},
		{
			name:               "peer role mode allow, peer role received",
			peerRoleEnabled:    true,
			peerRoleStrictMode: false,
			peerRoleLocalRole:  packet.PeerRoleRolePeer,
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
								Code:   packet.PeerRoleCapabilityCode,
								Length: 4,
								Value: packet.PeerRoleCapability{
									PeerRole: packet.PeerRoleRolePeer,
								},
							},
						},
					},
				},
			},
			wantIdle: false,
		},

		// mode strict
		{
			name:               "peer role mode strict, peer role received",
			peerRoleEnabled:    true,
			peerRoleStrictMode: true,
			peerRoleLocalRole:  packet.PeerRoleRoleProvider,
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
								Code:   packet.PeerRoleCapabilityCode,
								Length: 4,
								Value: packet.PeerRoleCapability{
									PeerRole: packet.PeerRoleRoleCustomer,
								},
							},
						},
					},
				},
			},
			wantIdle: false,
		},
		{
			name:               "peer role mode strict, no peer role received",
			peerRoleEnabled:    true,
			peerRoleStrictMode: true,
			msg: packet.BGPOpen{
				HoldTime:      90,
				BGPIdentifier: 1,
				Version:       4,
				ASN:           23456,
				OptParmLen:    0,
				OptParams:     []packet.OptParam{},
			},
			wantIdle: true,
			errmsg:   "role misatch error: Strict mode configured but peer didn't advertise a BGP role",
		},

		// Multiple peer roles
		{
			name:               "peer role mode strict, multiple identical peer role received",
			peerRoleEnabled:    true,
			peerRoleStrictMode: true,
			peerRoleLocalRole:  packet.PeerRoleRoleCustomer,
			errmsg:             "role misatch error: Multiple different BGP roles received from peer",
			msg: packet.BGPOpen{
				HoldTime:      90,
				BGPIdentifier: 1,
				Version:       4,
				ASN:           23456,
				OptParmLen:    1,
				OptParams: []packet.OptParam{
					{
						Type:   packet.CapabilitiesParamType,
						Length: 10,
						Value: packet.Capabilities{
							packet.Capability{
								Code:   packet.PeerRoleCapabilityCode,
								Length: 4,
								Value: packet.PeerRoleCapability{
									PeerRole: packet.PeerRoleRoleProvider,
								},
							},
							packet.Capability{
								Code:   packet.PeerRoleCapabilityCode,
								Length: 4,
								Value: packet.PeerRoleCapability{
									PeerRole: packet.PeerRoleRoleProvider,
								},
							},
						},
					},
				},
			},
			wantIdle: false,
		},
		{
			name:               "peer role mode strict, multiple different peer role received",
			peerRoleEnabled:    true,
			peerRoleStrictMode: true,
			peerRoleLocalRole:  packet.PeerRoleRolePeer,
			errmsg:             "role misatch error: Multiple different BGP roles received from peer",
			msg: packet.BGPOpen{
				HoldTime:      90,
				BGPIdentifier: 1,
				Version:       4,
				ASN:           23456,
				OptParmLen:    1,
				OptParams: []packet.OptParam{
					{
						Type:   packet.CapabilitiesParamType,
						Length: 10,
						Value: packet.Capabilities{
							packet.Capability{
								Code:   packet.PeerRoleCapabilityCode,
								Length: 4,
								Value: packet.PeerRoleCapability{
									PeerRole: packet.PeerRoleRolePeer,
								},
							},
							packet.Capability{
								Code:   packet.PeerRoleCapabilityCode,
								Length: 4,
								Value: packet.PeerRoleCapability{
									PeerRole: packet.PeerRoleRoleRSClient,
								},
							},
						},
					},
				},
			},
			wantIdle: true,
		},

		// (Remaining) Role pair tests for allowed pairs
		// Peer <-> Peer, Provider <-> Customer, and Customer <-> Provider implicitly tested above
		{
			name:               "valid peer role pair RS <-> RS-Client",
			peerRoleEnabled:    true,
			peerRoleStrictMode: true,
			peerRoleLocalRole:  packet.PeerRoleRoleRS,
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
								Code:   packet.PeerRoleCapabilityCode,
								Length: 4,
								Value: packet.PeerRoleCapability{
									PeerRole: packet.PeerRoleRoleRSClient,
								},
							},
						},
					},
				},
			},
			wantIdle: false,
		},
		{
			name:               "valid peer role pair RS-Client <-> RS",
			peerRoleEnabled:    true,
			peerRoleStrictMode: true,
			peerRoleLocalRole:  packet.PeerRoleRoleRSClient,
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
								Code:   packet.PeerRoleCapabilityCode,
								Length: 4,
								Value: packet.PeerRoleCapability{
									PeerRole: packet.PeerRoleRoleRS,
								},
							},
						},
					},
				},
			},
			wantIdle: false,
		},

		// (Remaining) Role pair tests for disallowed pairs
		{
			name:               "invalid peer role pair Provider <-> Provider",
			peerRoleEnabled:    true,
			peerRoleStrictMode: true,
			peerRoleLocalRole:  packet.PeerRoleRoleProvider,
			errmsg:             "role misatch error: Local role Provider incompatible to remote role Provider",
			msg: packet.BGPOpen{
				HoldTime:      90,
				BGPIdentifier: 1,
				Version:       4,
				ASN:           23456,
				OptParmLen:    1,
				OptParams: []packet.OptParam{
					{
						Type:   packet.CapabilitiesParamType,
						Length: 10,
						Value: packet.Capabilities{
							packet.Capability{
								Code:   packet.PeerRoleCapabilityCode,
								Length: 4,
								Value: packet.PeerRoleCapability{
									PeerRole: packet.PeerRoleRoleProvider,
								},
							},
						},
					},
				},
			},
			wantIdle: true,
		},
		{
			name:               "invalid peer role pair Provider <-> RS",
			peerRoleEnabled:    true,
			peerRoleStrictMode: true,
			peerRoleLocalRole:  packet.PeerRoleRoleProvider,
			errmsg:             "role misatch error: Local role Provider incompatible to remote role RS",
			msg: packet.BGPOpen{
				HoldTime:      90,
				BGPIdentifier: 1,
				Version:       4,
				ASN:           23456,
				OptParmLen:    1,
				OptParams: []packet.OptParam{
					{
						Type:   packet.CapabilitiesParamType,
						Length: 10,
						Value: packet.Capabilities{
							packet.Capability{
								Code:   packet.PeerRoleCapabilityCode,
								Length: 4,
								Value: packet.PeerRoleCapability{
									PeerRole: packet.PeerRoleRoleRS,
								},
							},
						},
					},
				},
			},
			wantIdle: true,
		},
		{
			name:               "invalid peer role pair Provider <-> RS-Client",
			peerRoleEnabled:    true,
			peerRoleStrictMode: true,
			peerRoleLocalRole:  packet.PeerRoleRoleProvider,
			errmsg:             "role misatch error: Local role Provider incompatible to remote role RS-Client",
			msg: packet.BGPOpen{
				HoldTime:      90,
				BGPIdentifier: 1,
				Version:       4,
				ASN:           23456,
				OptParmLen:    1,
				OptParams: []packet.OptParam{
					{
						Type:   packet.CapabilitiesParamType,
						Length: 10,
						Value: packet.Capabilities{
							packet.Capability{
								Code:   packet.PeerRoleCapabilityCode,
								Length: 4,
								Value: packet.PeerRoleCapability{
									PeerRole: packet.PeerRoleRoleRSClient,
								},
							},
						},
					},
				},
			},
			wantIdle: true,
		},
		{
			name:               "invalid peer role pair Provider <-> Peer",
			peerRoleEnabled:    true,
			peerRoleStrictMode: true,
			peerRoleLocalRole:  packet.PeerRoleRoleProvider,
			errmsg:             "role misatch error: Local role Provider incompatible to remote role Peer",
			msg: packet.BGPOpen{
				HoldTime:      90,
				BGPIdentifier: 1,
				Version:       4,
				ASN:           23456,
				OptParmLen:    1,
				OptParams: []packet.OptParam{
					{
						Type:   packet.CapabilitiesParamType,
						Length: 10,
						Value: packet.Capabilities{
							packet.Capability{
								Code:   packet.PeerRoleCapabilityCode,
								Length: 4,
								Value: packet.PeerRoleCapability{
									PeerRole: packet.PeerRoleRolePeer,
								},
							},
						},
					},
				},
			},
			wantIdle: true,
		},

		// Invalid peer role pairs RS <-> *
		{
			name:               "invalid peer role pair RS <-> Provider",
			peerRoleEnabled:    true,
			peerRoleStrictMode: true,
			peerRoleLocalRole:  packet.PeerRoleRoleRS,
			errmsg:             "role misatch error: Local role RS incompatible to remote role Provider",
			msg: packet.BGPOpen{
				HoldTime:      90,
				BGPIdentifier: 1,
				Version:       4,
				ASN:           23456,
				OptParmLen:    1,
				OptParams: []packet.OptParam{
					{
						Type:   packet.CapabilitiesParamType,
						Length: 10,
						Value: packet.Capabilities{
							packet.Capability{
								Code:   packet.PeerRoleCapabilityCode,
								Length: 4,
								Value: packet.PeerRoleCapability{
									PeerRole: packet.PeerRoleRoleProvider,
								},
							},
						},
					},
				},
			},
			wantIdle: true,
		},
		{
			name:               "invalid peer role pair RS <-> RS",
			peerRoleEnabled:    true,
			peerRoleStrictMode: true,
			peerRoleLocalRole:  packet.PeerRoleRoleRS,
			errmsg:             "role misatch error: Local role RS incompatible to remote role RS",
			msg: packet.BGPOpen{
				HoldTime:      90,
				BGPIdentifier: 1,
				Version:       4,
				ASN:           23456,
				OptParmLen:    1,
				OptParams: []packet.OptParam{
					{
						Type:   packet.CapabilitiesParamType,
						Length: 10,
						Value: packet.Capabilities{
							packet.Capability{
								Code:   packet.PeerRoleCapabilityCode,
								Length: 4,
								Value: packet.PeerRoleCapability{
									PeerRole: packet.PeerRoleRoleRS,
								},
							},
						},
					},
				},
			},
			wantIdle: true,
		},
		{
			name:               "invalid peer role pair RS <-> Customer",
			peerRoleEnabled:    true,
			peerRoleStrictMode: true,
			peerRoleLocalRole:  packet.PeerRoleRoleRS,
			errmsg:             "role misatch error: Local role RS incompatible to remote role Customer",
			msg: packet.BGPOpen{
				HoldTime:      90,
				BGPIdentifier: 1,
				Version:       4,
				ASN:           23456,
				OptParmLen:    1,
				OptParams: []packet.OptParam{
					{
						Type:   packet.CapabilitiesParamType,
						Length: 10,
						Value: packet.Capabilities{
							packet.Capability{
								Code:   packet.PeerRoleCapabilityCode,
								Length: 4,
								Value: packet.PeerRoleCapability{
									PeerRole: packet.PeerRoleRoleCustomer,
								},
							},
						},
					},
				},
			},
			wantIdle: true,
		},
		{
			name:               "invalid peer role pair RS <-> Peer",
			peerRoleEnabled:    true,
			peerRoleStrictMode: true,
			peerRoleLocalRole:  packet.PeerRoleRoleRS,
			errmsg:             "role misatch error: Local role RS incompatible to remote role Peer",
			msg: packet.BGPOpen{
				HoldTime:      90,
				BGPIdentifier: 1,
				Version:       4,
				ASN:           23456,
				OptParmLen:    1,
				OptParams: []packet.OptParam{
					{
						Type:   packet.CapabilitiesParamType,
						Length: 10,
						Value: packet.Capabilities{
							packet.Capability{
								Code:   packet.PeerRoleCapabilityCode,
								Length: 4,
								Value: packet.PeerRoleCapability{
									PeerRole: packet.PeerRoleRolePeer,
								},
							},
						},
					},
				},
			},
			wantIdle: true,
		},

		// Invalid role pairs for RS-Client <-> *
		{
			name:               "invalid peer role pair RS-Client <-> Provider",
			peerRoleEnabled:    true,
			peerRoleStrictMode: true,
			peerRoleLocalRole:  packet.PeerRoleRoleRSClient,
			errmsg:             "role misatch error: Local role RS-Client incompatible to remote role Provider",
			msg: packet.BGPOpen{
				HoldTime:      90,
				BGPIdentifier: 1,
				Version:       4,
				ASN:           23456,
				OptParmLen:    1,
				OptParams: []packet.OptParam{
					{
						Type:   packet.CapabilitiesParamType,
						Length: 10,
						Value: packet.Capabilities{
							packet.Capability{
								Code:   packet.PeerRoleCapabilityCode,
								Length: 4,
								Value: packet.PeerRoleCapability{
									PeerRole: packet.PeerRoleRoleProvider,
								},
							},
						},
					},
				},
			},
			wantIdle: true,
		},
		{
			name:               "invalid peer role pair RS-Client <-> RS-Client",
			peerRoleEnabled:    true,
			peerRoleStrictMode: true,
			peerRoleLocalRole:  packet.PeerRoleRoleRSClient,
			errmsg:             "role misatch error: Local role RS-Client incompatible to remote role RS-Client",
			msg: packet.BGPOpen{
				HoldTime:      90,
				BGPIdentifier: 1,
				Version:       4,
				ASN:           23456,
				OptParmLen:    1,
				OptParams: []packet.OptParam{
					{
						Type:   packet.CapabilitiesParamType,
						Length: 10,
						Value: packet.Capabilities{
							packet.Capability{
								Code:   packet.PeerRoleCapabilityCode,
								Length: 4,
								Value: packet.PeerRoleCapability{
									PeerRole: packet.PeerRoleRoleRSClient,
								},
							},
						},
					},
				},
			},
			wantIdle: true,
		},
		{
			name:               "invalid peer role pair RS-Clinet <-> Customer",
			peerRoleEnabled:    true,
			peerRoleStrictMode: true,
			peerRoleLocalRole:  packet.PeerRoleRoleRSClient,
			errmsg:             "role misatch error: Local role RS-Client incompatible to remote role Customer",
			msg: packet.BGPOpen{
				HoldTime:      90,
				BGPIdentifier: 1,
				Version:       4,
				ASN:           23456,
				OptParmLen:    1,
				OptParams: []packet.OptParam{
					{
						Type:   packet.CapabilitiesParamType,
						Length: 10,
						Value: packet.Capabilities{
							packet.Capability{
								Code:   packet.PeerRoleCapabilityCode,
								Length: 4,
								Value: packet.PeerRoleCapability{
									PeerRole: packet.PeerRoleRoleCustomer,
								},
							},
						},
					},
				},
			},
			wantIdle: true,
		},
		{
			name:               "invalid peer role pair RS-Client <-> Peer",
			peerRoleEnabled:    true,
			peerRoleStrictMode: true,
			peerRoleLocalRole:  packet.PeerRoleRoleRSClient,
			errmsg:             "role misatch error: Local role RS-Client incompatible to remote role Peer",
			msg: packet.BGPOpen{
				HoldTime:      90,
				BGPIdentifier: 1,
				Version:       4,
				ASN:           23456,
				OptParmLen:    1,
				OptParams: []packet.OptParam{
					{
						Type:   packet.CapabilitiesParamType,
						Length: 10,
						Value: packet.Capabilities{
							packet.Capability{
								Code:   packet.PeerRoleCapabilityCode,
								Length: 4,
								Value: packet.PeerRoleCapability{
									PeerRole: packet.PeerRoleRolePeer,
								},
							},
						},
					},
				},
			},
			wantIdle: true,
		},

		// Invalid role pairs for Peer <-> *
		{
			name:               "invalid peer role pair Peer <-> Provider",
			peerRoleEnabled:    true,
			peerRoleStrictMode: true,
			peerRoleLocalRole:  packet.PeerRoleRolePeer,
			errmsg:             "role misatch error: Local role Peer incompatible to remote role Provider",
			msg: packet.BGPOpen{
				HoldTime:      90,
				BGPIdentifier: 1,
				Version:       4,
				ASN:           23456,
				OptParmLen:    1,
				OptParams: []packet.OptParam{
					{
						Type:   packet.CapabilitiesParamType,
						Length: 10,
						Value: packet.Capabilities{
							packet.Capability{
								Code:   packet.PeerRoleCapabilityCode,
								Length: 4,
								Value: packet.PeerRoleCapability{
									PeerRole: packet.PeerRoleRoleProvider,
								},
							},
						},
					},
				},
			},
			wantIdle: true,
		},
		{
			name:               "invalid peer role pair Peer <-> RS",
			peerRoleEnabled:    true,
			peerRoleStrictMode: true,
			peerRoleLocalRole:  packet.PeerRoleRolePeer,
			errmsg:             "role misatch error: Local role Peer incompatible to remote role RS",
			msg: packet.BGPOpen{
				HoldTime:      90,
				BGPIdentifier: 1,
				Version:       4,
				ASN:           23456,
				OptParmLen:    1,
				OptParams: []packet.OptParam{
					{
						Type:   packet.CapabilitiesParamType,
						Length: 10,
						Value: packet.Capabilities{
							packet.Capability{
								Code:   packet.PeerRoleCapabilityCode,
								Length: 4,
								Value: packet.PeerRoleCapability{
									PeerRole: packet.PeerRoleRoleRS,
								},
							},
						},
					},
				},
			},
			wantIdle: true,
		},
		{
			name:               "invalid peer role pair Peer <-> RS-Client",
			peerRoleEnabled:    true,
			peerRoleStrictMode: true,
			peerRoleLocalRole:  packet.PeerRoleRolePeer,
			errmsg:             "role misatch error: Local role Peer incompatible to remote role RS-Client",
			msg: packet.BGPOpen{
				HoldTime:      90,
				BGPIdentifier: 1,
				Version:       4,
				ASN:           23456,
				OptParmLen:    1,
				OptParams: []packet.OptParam{
					{
						Type:   packet.CapabilitiesParamType,
						Length: 10,
						Value: packet.Capabilities{
							packet.Capability{
								Code:   packet.PeerRoleCapabilityCode,
								Length: 4,
								Value: packet.PeerRoleCapability{
									PeerRole: packet.PeerRoleRoleRSClient,
								},
							},
						},
					},
				},
			},
			wantIdle: true,
		},
		{
			name:               "invalid peer role pair Peer <-> Customer",
			peerRoleEnabled:    true,
			peerRoleStrictMode: true,
			peerRoleLocalRole:  packet.PeerRoleRolePeer,
			errmsg:             "role misatch error: Local role Peer incompatible to remote role Customer",
			msg: packet.BGPOpen{
				HoldTime:      90,
				BGPIdentifier: 1,
				Version:       4,
				ASN:           23456,
				OptParmLen:    1,
				OptParams: []packet.OptParam{
					{
						Type:   packet.CapabilitiesParamType,
						Length: 10,
						Value: packet.Capabilities{
							packet.Capability{
								Code:   packet.PeerRoleCapabilityCode,
								Length: 4,
								Value: packet.PeerRoleCapability{
									PeerRole: packet.PeerRoleRoleCustomer,
								},
							},
						},
					},
				},
			},
			wantIdle: true,
		},

		// Invalid role pairs for * <-> unassigned
		{
			name:               "invalid peer role pair Provider <-> unassigned",
			peerRoleEnabled:    true,
			peerRoleStrictMode: true,
			peerRoleLocalRole:  packet.PeerRoleRoleProvider,
			errmsg:             "role misatch error: Local role Provider incompatible to remote role Unknown",
			msg: packet.BGPOpen{
				HoldTime:      90,
				BGPIdentifier: 1,
				Version:       4,
				ASN:           23456,
				OptParmLen:    1,
				OptParams: []packet.OptParam{
					{
						Type:   packet.CapabilitiesParamType,
						Length: 10,
						Value: packet.Capabilities{
							packet.Capability{
								Code:   packet.PeerRoleCapabilityCode,
								Length: 4,
								Value: packet.PeerRoleCapability{
									PeerRole: 23,
								},
							},
						},
					},
				},
			},
			wantIdle: true,
		},
		{
			name:               "invalid peer role pair RS <-> unassigned",
			peerRoleEnabled:    true,
			peerRoleStrictMode: true,
			peerRoleLocalRole:  packet.PeerRoleRoleRS,
			errmsg:             "role misatch error: Local role RS incompatible to remote role Unknown",
			msg: packet.BGPOpen{
				HoldTime:      90,
				BGPIdentifier: 1,
				Version:       4,
				ASN:           23456,
				OptParmLen:    1,
				OptParams: []packet.OptParam{
					{
						Type:   packet.CapabilitiesParamType,
						Length: 10,
						Value: packet.Capabilities{
							packet.Capability{
								Code:   packet.PeerRoleCapabilityCode,
								Length: 4,
								Value: packet.PeerRoleCapability{
									PeerRole: 23,
								},
							},
						},
					},
				},
			},
			wantIdle: true,
		},
		{
			name:               "invalid peer role pair RS-Clinet <-> unassigned",
			peerRoleEnabled:    true,
			peerRoleStrictMode: true,
			peerRoleLocalRole:  packet.PeerRoleRoleRSClient,
			errmsg:             "role misatch error: Local role RS-Client incompatible to remote role Unknown",
			msg: packet.BGPOpen{
				HoldTime:      90,
				BGPIdentifier: 1,
				Version:       4,
				ASN:           23456,
				OptParmLen:    1,
				OptParams: []packet.OptParam{
					{
						Type:   packet.CapabilitiesParamType,
						Length: 10,
						Value: packet.Capabilities{
							packet.Capability{
								Code:   packet.PeerRoleCapabilityCode,
								Length: 4,
								Value: packet.PeerRoleCapability{
									PeerRole: 23,
								},
							},
						},
					},
				},
			},
			wantIdle: true,
		},
		{
			name:               "invalid peer role pair Customer <-> unassigned",
			peerRoleEnabled:    true,
			peerRoleStrictMode: true,
			peerRoleLocalRole:  packet.PeerRoleRoleCustomer,
			errmsg:             "role misatch error: Local role Customer incompatible to remote role Unknown",
			msg: packet.BGPOpen{
				HoldTime:      90,
				BGPIdentifier: 1,
				Version:       4,
				ASN:           23456,
				OptParmLen:    1,
				OptParams: []packet.OptParam{
					{
						Type:   packet.CapabilitiesParamType,
						Length: 10,
						Value: packet.Capabilities{
							packet.Capability{
								Code:   packet.PeerRoleCapabilityCode,
								Length: 4,
								Value: packet.PeerRoleCapability{
									PeerRole: 23,
								},
							},
						},
					},
				},
			},
			wantIdle: true,
		},
		{
			name:               "invalid peer role pair Peer <-> unassigned",
			peerRoleEnabled:    true,
			peerRoleStrictMode: true,
			peerRoleLocalRole:  packet.PeerRoleRolePeer,
			errmsg:             "role misatch error: Local role Peer incompatible to remote role Unknown",
			msg: packet.BGPOpen{
				HoldTime:      90,
				BGPIdentifier: 1,
				Version:       4,
				ASN:           23456,
				OptParmLen:    1,
				OptParams: []packet.OptParam{
					{
						Type:   packet.CapabilitiesParamType,
						Length: 10,
						Value: packet.Capabilities{
							packet.Capability{
								Code:   packet.PeerRoleCapabilityCode,
								Length: 4,
								Value: packet.PeerRoleCapability{
									PeerRole: 23,
								},
							},
						},
					},
				},
			},
			wantIdle: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fsm := newFSM(&peer{
				peerASN: 23456,
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

			s.fsm.peer.peerRoleEnabled = test.peerRoleEnabled
			s.fsm.peer.peerRoleStrictMode = test.peerRoleStrictMode
			s.fsm.peer.peerRoleLocal = test.peerRoleLocalRole

			state, errmsg := s.handleOpenMessage(&test.msg)

			if test.wantIdle {
				assert.IsType(t, &idleState{}, state, "state")
				assert.Equal(t, test.errmsg, errmsg, "errmsg")
				return
			}

			assert.IsType(t, &openConfirmState{}, state, errmsg)
		})
	}
}
