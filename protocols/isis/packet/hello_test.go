package packet

import (
	"bytes"
	"testing"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/stretchr/testify/assert"
)

func TestP2PHelloSerialize(t *testing.T) {
	tests := []struct {
		name     string
		input    *P2PHello
		expected []byte
	}{
		{
			name: "Full",
			input: &P2PHello{
				CircuitType:    2,
				SystemID:       types.SystemID{1, 2, 3, 4, 5, 6},
				HoldingTimer:   27,
				PDULength:      19,
				LocalCircuitID: 1,
				TLVs: []TLV{
					&AreaAddressesTLV{
						TLVType:   1,
						TLVLength: 4,
						AreaIDs: []types.AreaID{
							{
								1, 2, 3,
							},
						},
					},
				},
			},
			expected: []byte{
				2,                // Circuit Type
				1, 2, 3, 4, 5, 6, // SystemID
				0, 27, // Holding Timer
				0, 19, // PDU Length
				1,       // Local Circuits ID
				1,       // Area addresses TLV
				4,       // TLV Length
				3,       // Area length
				1, 2, 3, // Area
			},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		test.input.Serialize(buf)

		assert.Equalf(t, test.expected, buf.Bytes(), "Test %q", test.name)
	}
}

func TestDecodeP2PHello(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantFail bool
		expected *P2PHello
	}{
		{
			name: "Full",
			input: []byte{
				2,                // Circuit Type
				1, 2, 3, 4, 5, 6, // SystemID
				0, 27, // Holding Timer
				0, 19, // PDU Length
				1,       // Local Circuits ID
				1,       // Area addresses TLV
				4,       // TLV Length
				3,       // Area length
				1, 2, 3, // Area
			},
			wantFail: false,
			expected: &P2PHello{
				CircuitType:    2,
				SystemID:       types.SystemID{1, 2, 3, 4, 5, 6},
				HoldingTimer:   27,
				PDULength:      19,
				LocalCircuitID: 1,
				TLVs: []TLV{
					&AreaAddressesTLV{
						TLVType:   1,
						TLVLength: 4,
						AreaIDs: []types.AreaID{
							{
								1, 2, 3,
							},
						},
					},
				},
			},
		},
		{
			name: "Incomplete hello",
			input: []byte{
				2,                // Circuit Type
				1, 2, 3, 4, 5, 6, // SystemID
				0, 27, // Holding Timer
				0, 19, // PDU Length
			},
			wantFail: true,
		},
		{
			name: "Incomplete TLV",
			input: []byte{
				2,                // Circuit Type
				1, 2, 3, 4, 5, 6, // SystemID
				0, 27, // Holding Timer
				0, 19, // PDU Length
				1, // Local Circuits ID
				1, // Area addresses TLV
				4, // TLV Length
			},
			wantFail: true,
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		pkt, err := DecodeP2PHello(buf)
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

		assert.Equalf(t, test.expected, pkt, "Test %q", test.name)
	}
}

func TestDecodeISISHello(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantFail bool
		expected *L2Hello
	}{
		{
			name: "No TLVs",
			input: []byte{
				2,
				1, 2, 3, 4, 5, 6,
				0, 200,
				0, 18,
				150,
				0,
				1, 1, 1, 2, 2, 2,
			},
			expected: &L2Hello{
				CircuitType:  2,
				SystemID:     [6]byte{1, 2, 3, 4, 5, 6},
				HoldingTimer: 200,
				PDULength:    18,
				Priority:     150,
				DesignatedIS: [6]byte{1, 1, 1, 2, 2, 2},
				TLVs:         []TLV{},
			},
		},
		{
			name: "Hello with IS Neighbor TLV",
			input: []byte{
				2,
				1, 2, 3, 4, 5, 6,
				0, 200,
				0, 26,
				150,
				0,
				1, 1, 1, 2, 2, 2,
				6,
				6,
				2, 2, 2, 3, 3, 3,
			},
			expected: &L2Hello{
				CircuitType:  2,
				SystemID:     [6]byte{1, 2, 3, 4, 5, 6},
				HoldingTimer: 200,
				PDULength:    26,
				Priority:     150,
				DesignatedIS: [6]byte{1, 1, 1, 2, 2, 2},
				TLVs: []TLV{
					&ISNeighborsTLV{
						TLVType:      6,
						TLVLength:    6,
						NeighborSNPA: [6]byte{2, 2, 2, 3, 3, 3},
					},
				},
			},
		},
		{
			name: "Full Hello (JunOS)",
			input: []byte{
				2,                // CircuitType
				1, 2, 3, 4, 5, 6, // SystemID
				0, 200, // Holding Timer
				0, 41, // PDU Length
				150,              // Prio
				0,                // Reserved
				1, 1, 1, 2, 2, 2, // Designated IS
				6,                // Type = ISNeighborsTLV
				6,                // Length
				2, 2, 2, 3, 3, 3, // Neighbor
				129,
				2,
				0xcc, 0x8e,
				132,
				4,
				10, 0, 0, 0,
				1,
				7,
				6, 49, 10, 0, 0, 20, 30,
			},
			expected: &L2Hello{
				CircuitType:  2,
				SystemID:     [6]byte{1, 2, 3, 4, 5, 6},
				HoldingTimer: 200,
				PDULength:    41,
				Priority:     150,
				DesignatedIS: [6]byte{1, 1, 1, 2, 2, 2},
				TLVs: []TLV{
					&ISNeighborsTLV{
						TLVType:      6,
						TLVLength:    6,
						NeighborSNPA: [6]byte{2, 2, 2, 3, 3, 3},
					},
					&ProtocolsSupportedTLV{
						TLVType:                 129,
						TLVLength:               2,
						NetworkLayerProtocolIDs: []byte{0xcc, 0x8e},
					},
					&IPInterfaceAddressTLV{
						TLVType:     132,
						TLVLength:   4,
						IPv4Address: 167772160,
					},
					&AreaAddressesTLV{
						TLVType:   1,
						TLVLength: 7,
						AreaIDs: []types.AreaID{
							{
								49, 10, 0, 0, 20, 30,
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		buffer := bytes.NewBuffer(test.input)
		pdu, err := DecodeL2Hello(buffer)

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

		assert.Equalf(t, test.expected, pdu, "Test: %q", test.name)
	}
}
