package server

import (
	"testing"

	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/stretchr/testify/assert"
)

type mockNeighborManager struct {
	wantFailGetNeighbor    bool
	callCounterSetNeighbor int
}

func (mnm *mockNeighborManager) setNeighbor(n *neighbor) {
	mnm.callCounterSetNeighbor++
}

func (mnm *mockNeighborManager) getNeighbor(addr types.MACAddress) *neighbor {
	if mnm.wantFailGetNeighbor {
		return nil
	}

	return &neighbor{}
}

func TestProcessP2PHello(t *testing.T) {
	tests := []struct {
		name                    string
		d                       *dev
		h                       *packet.P2PHello
		wantFail                bool
		wantErr                 string
		expectedNeighborManager *mockNeighborManager
	}{
		{
			name: "Invalid Circuit Type",
			d: &dev{
				level2: &level{
					neighborManager: &mockNeighborManager{},
				},
			},
			h: &packet.P2PHello{
				CircuitType:    0x01,
				SystemID:       types.SystemID{10, 20, 30, 40, 50, 60},
				HoldingTimer:   27,
				LocalCircuitID: 123,
			},
			wantFail: true,
			wantErr:  "Unsupported P2P Hello: Level 1",
		},
		{
			name: "p2pHelloToNeighbor fail",
			d: &dev{
				name: "eth0",
			},
			h: &packet.P2PHello{ // No TLVs
				CircuitType:    0x02,
				SystemID:       types.SystemID{10, 20, 30, 40, 50, 60},
				HoldingTimer:   27,
				LocalCircuitID: 123,
			},
			wantFail: true,
			wantErr:  "Unable to create neighbor object from hello: Received a P2P hello PDU without P2P Adjacency TLV on eth0",
		},
		{
			name: "Check neighbor manager called",
			d: &dev{
				level2: &level{
					neighborManager: &mockNeighborManager{},
				},
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
			expectedNeighborManager: &mockNeighborManager{
				callCounterSetNeighbor: 1,
			},
		},
	}

	for _, test := range tests {
		err := test.d.processP2PHello(test.h, types.MACAddress{10, 20, 30, 40, 50, 60})
		if err != nil && test.wantFail {
			assert.Equal(t, test.wantErr, err.Error(), test.name)
			continue
		}

		if err != nil && !test.wantFail {
			t.Errorf("Unexpected failure for test %q: %v", test.name, err)
			continue
		}

		if err == nil && test.wantFail {
			t.Errorf("Unexpected success for test %q", test.name)
			continue
		}

		assert.Equal(t, test.expectedNeighborManager, test.d.level2.neighborManager, test.name)
	}
}

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
		n, err := test.d.p2pHelloToNeighbor(test.h, types.MACAddress{})
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
