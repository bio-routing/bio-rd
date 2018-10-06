package packet

import (
	"bytes"
	"testing"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantFail bool
		expected *ISISPacket
	}{
		{
			name: "P2P Hello",
			input: []byte{
				// LLC
				0xfe, // DSAP
				0xfe, // SSAP
				0x03, // Control Fields

				// Header
				0x83,
				20,
				1,
				0,
				17, // PDU Type P2P Hello
				1,
				0,
				0,

				// P2P Hello
				02,
				0, 0, 0, 0, 0, 2,
				0, 27,
				0, 50,
				1,

				//TLVs
				240, 5, 0x02, 0x00, 0x00, 0x01, 0x4b,
				129, 2, 0xcc, 0x8e,
				132, 4, 192, 168, 1, 0,
				1, 6, 0x05, 0x49, 0x00, 0x01, 0x00, 0x10,
				211, 3, 0, 0, 0,
			},
			wantFail: false,
			expected: &ISISPacket{
				Header: &ISISHeader{
					ProtoDiscriminator:  0x83,
					LengthIndicator:     20,
					ProtocolIDExtension: 1,
					IDLength:            0,
					PDUType:             17,
					Version:             1,
					MaxAreaAddresses:    0,
				},
				Body: &P2PHello{
					CircuitType:    2,
					SystemID:       types.SystemID{0, 0, 0, 0, 0, 2},
					HoldingTimer:   27,
					PDULength:      50,
					LocalCircuitID: 1,
					TLVs: []TLV{
						&P2PAdjacencyStateTLV{
							TLVType:                240,
							TLVLength:              5,
							AdjacencyState:         2,
							ExtendedLocalCircuitID: 0x0000014b,
						},
						&ProtocolsSupportedTLV{
							TLVType:                 129,
							TLVLength:               2,
							NerworkLayerProtocolIDs: []uint8{0xcc, 0x8e},
						},
						&IPInterfaceAddressTLV{
							TLVType:     132,
							TLVLength:   4,
							IPv4Address: 3232235776,
						},
						&AreaAddressesTLV{
							TLVType:   1,
							TLVLength: 6,
							AreaIDs: []types.AreaID{
								{
									0x49, 0x00, 0x01, 0x00, 0x10,
								},
							},
						},
						&UnknownTLV{
							TLVType:   211,
							TLVLength: 3,
							TLVValue:  []byte{0, 0, 0},
						},
					},
				},
			},
		},
		{
			name: "Incomplete header",
			input: []byte{
				// LLC
				0xfe, // DSAP
				0xfe, // SSAP
				0x03, // Control Fields

				// Header
				0x83,
				20,
				1,
				0,
			},
			wantFail: true,
		},
		{
			name: "Incomplete P2P Hello",
			input: []byte{
				// LLC
				0xfe, // DSAP
				0xfe, // SSAP
				0x03, // Control Fields

				// Header
				0x83,
				20,
				1,
				0,
				17, // PDU Type P2P Hello
				1,
				0,
				0,

				// P2P Hello
				02,
				0, 0, 0, 0, 0, 2,
				0, 27,
			},
			wantFail: true,
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		pkt, err := Decode(buf)

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
