package packet

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelloSerialize(t *testing.T) {
	tests := []struct {
		name     string
		hello    *L2Hello
		expected []byte
	}{
		{
			name: "Test #1",
			hello: &L2Hello{
				CircuitType:  2,
				SystemID:     [6]byte{1, 2, 3, 4, 5, 6},
				HoldingTimer: 27,
				PDULength:    100,
				Priority:     128,
				DesignatedIS: [6]byte{1, 2, 3, 4, 5, 6},
				TLVs:         []TLV{},
			},
		},
	}

	for _, test := range tests {
		res := test.hello.serialize()
		assert.Equalf(t, test.expected, res, "Test: %q", test.name)
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
			name: "Header only",
			input: []byte{
				2,
				1, 2, 3, 4, 5, 6,
				0, 200,
				0, 18,
				150,
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
			name: "Unknown TLVs",
			input: []byte{
				2,
				1, 2, 3, 4, 5, 6,
				0, 200,
				0, 22,
				150,
				1, 1, 1, 2, 2, 2,
				0, 2, 10, 10,
			},
			expected: &L2Hello{
				CircuitType:  2,
				SystemID:     [6]byte{1, 2, 3, 4, 5, 6},
				HoldingTimer: 200,
				PDULength:    22,
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
				1, 1, 1, 2, 2, 2,
				6,
				6,
				2, 2, 2, 3, 3, 3, 3,
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
				2,
				1, 2, 3, 4, 5, 6,
				0, 200,
				0, 41,
				150,
				1, 1, 1, 2, 2, 2,
				6,
				6,
				2, 2, 2, 3, 3, 3,
				129,
				2,
				0xcc, 0x8e,
				132,
				4,
				10, 0, 0, 0,
				1,
				3,
				2,
				10, 20,
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
						TLVType:                129,
						TLVLength:              2,
						NerworkLayerProtocolID: []byte{0xcc, 0x8e},
					},
					&IPInterfaceAddressTLV{
						TLVType:     132,
						TLVLength:   4,
						IPv4Address: 167772160,
					},
					&AreaAddressTLV{
						TLVType:    1,
						TLVLength:  3,
						AreaLength: 2,
						AreaID:     []byte{10, 20},
					},
				},
			},
		},
	}

	for _, test := range tests {
		buffer := bytes.NewBuffer(test.input)
		pdu, err := decodeISISHello(buffer)

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
