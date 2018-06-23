package server

import (
	"testing"
	"time"

	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/stretchr/testify/assert"
)

func TestOpenMessage(t *testing.T) {
	tests := []struct {
		name     string
		localASN uint32
		holdTime time.Duration
		routerID uint32
		expected packet.BGPOpen
	}{
		{
			name:     "16bit ASN",
			localASN: 12345,
			holdTime: time.Duration(30 * time.Second),
			routerID: 1,
			expected: packet.BGPOpen{
				ASN:           12345,
				BGPIdentifier: 1,
				HoldTime:      30,
				OptParams: []packet.OptParam{
					packet.OptParam{
						Type: packet.CapabilitiesParamType,
						Value: packet.Capabilities{
							packet.Capability{
								Code: 65,
								Value: packet.ASN4Capability{
									ASN4: 12345,
								},
							},
						},
					},
				},
				Version: 4,
			},
		},
		{
			name:     "32bit ASN",
			localASN: 202739,
			holdTime: time.Duration(30 * time.Second),
			routerID: 1,
			expected: packet.BGPOpen{
				ASN:           23456,
				BGPIdentifier: 1,
				HoldTime:      30,
				OptParams: []packet.OptParam{
					packet.OptParam{
						Type: packet.CapabilitiesParamType,
						Value: packet.Capabilities{
							packet.Capability{
								Code: 65,
								Value: packet.ASN4Capability{
									ASN4: 202739,
								},
							},
						},
					},
				},
				Version: 4,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := Peer{
				localASN: test.localASN,
				holdTime: test.holdTime,
				routerID: test.routerID,
				optOpenParams: []packet.OptParam{
					packet.OptParam{
						Type: packet.CapabilitiesParamType,
						Value: packet.Capabilities{
							packet.Capability{
								Code: 65,
								Value: packet.ASN4Capability{
									ASN4: test.localASN,
								},
							},
						},
					},
				},
			}

			fsm := newFSM2(&p)
			msg := fsm.openMessage()

			assert.Equal(t, &test.expected, msg)
		})
	}
}
