package server

import (
	"testing"

	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/stretchr/testify/assert"
)

func TestP2pHelloToNeighbor(t *testing.T) {
	tests := []struct {
		name     string
		d        *dev
		h        *packet.P2PHello
		wantFail bool
		log      string
		expected *neighbor
	}{
		{
			name: "Fail. No TLVs.",
			d: &dev{
				name: "eth0",
			},
			h: &packet.P2PHello{
				CircuitType:    0x02,
				SystemID:       types.SystemID{10, 20, 30, 40, 50, 60},
				HoldingTimer:   27,
				LocalCircuitID: 123,
			},
			wantFail: true,
			log:      "Received a P2P hello PDU without P2P Adjacency TLV on eth0",
		},
		{
			name: "Fail. InterfaceAddressesesTLV missing.",
			d: &dev{
				name: "eth0",
			},
			h: &packet.P2PHello{
				CircuitType:    0x02,
				SystemID:       types.SystemID{10, 20, 30, 40, 50, 60},
				HoldingTimer:   27,
				LocalCircuitID: 123,
				TLVs: []packet.TLV{
					&packet.P2PAdjacencyStateTLV{
						TLVType:   packet.P2PAdjacencyStateTLVType,
						TLVLength: 15,
					},
				},
			},
			wantFail: true,
			log:      "Received a P2P hello PDU without IP Interface Addresses TLV on eth0",
		},
		{
			name: "OK.",
			d: &dev{
				name: "eth0",
			},
			h: &packet.P2PHello{
				CircuitType:    0x02,
				SystemID:       types.SystemID{10, 20, 30, 40, 50, 60},
				HoldingTimer:   27,
				LocalCircuitID: 123,
				TLVs: []packet.TLV{
					&packet.P2PAdjacencyStateTLV{
						TLVType:                packet.P2PAdjacencyStateTLVType,
						TLVLength:              15,
						ExtendedLocalCircuitID: 1024,
					},
					&packet.IPInterfaceAddressesTLV{
						TLVType:       packet.IPInterfaceAddressesTLVType,
						TLVLength:     4,
						IPv4Addresses: []uint32{100},
					},
				},
			},
			wantFail: false,
			expected: &neighbor{
				systemID: types.SystemID{10, 20, 30, 40, 50, 60},
				dev: &dev{
					name: "eth0",
				},
				holdingTime:            27,
				localCircuitID:         123,
				extendedLocalCircuitID: 1024,
				ipInterfaceAddresses: []uint32{
					100,
				},
			},
		},
	}

	for _, test := range tests {
		n, err := test.d.p2pHelloToNeighbor(test.h)
		if err != nil {
			if test.wantFail {
				assert.Equal(t, test.log, err.Error(), test.name)
				continue
			}

			t.Errorf("Unexpected failure for test %q: %v", test.name, err)
			continue
		}

		if test.wantFail {
			t.Errorf("Unexpected success for test %q", test.name)
			continue
		}

		assert.Equal(t, test.expected, n, test.name)
	}
}
