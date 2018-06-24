package server

import (
	"testing"

	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/stretchr/testify/assert"
)

func TestOpenMsgReceived(t *testing.T) {
	tests := []struct {
		asn        uint32
		name       string
		msg        packet.BGPOpen
		wantsCease bool
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
			wantsCease: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fsm := newFSM2(&peer{
				peerASN: test.asn,
			})

			s := &openSentState{
				fsm: fsm,
			}

			state, _ := s.handleOpenMessage(&test.msg)

			if test.wantsCease {
				assert.IsType(t, &ceaseState{}, state, "state")
				return
			}

			assert.IsType(t, &openConfirmState{}, state, "state")
			assert.Equal(t, test.asn, s.peerASNRcvd, "asn")
		})
	}
}
