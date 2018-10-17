package packet

import (
	"bytes"
	"testing"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/stretchr/testify/assert"
)

func TestReadP2PAdjacencyStateTLV(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		tlvLength uint8
		wantFail  bool
		expected  *P2PAdjacencyStateTLV
	}{
		{
			name: "Full",
			input: []byte{
				1,
				0, 0, 0, 100,
				1, 2, 3, 4, 5, 6,
				0, 0, 0, 200,
			},
			tlvLength: 15,
			wantFail:  false,
			expected: &P2PAdjacencyStateTLV{
				TLVType:                        240,
				TLVLength:                      15,
				AdjacencyState:                 1,
				ExtendedLocalCircuitID:         100,
				NeighborSystemID:               types.SystemID{1, 2, 3, 4, 5, 6},
				NeighborExtendedLocalCircuitID: 200,
			},
		},
		{
			name: "Incomplete",
			input: []byte{
				1,
				0, 0, 0, 100,
				1, 2, 3, 4, 5, 6,
				0, 0,
			},
			tlvLength: 15,
			wantFail:  true,
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		tlv, err := readP2PAdjacencyStateTLV(buf, 240, test.tlvLength)

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

		assert.Equalf(t, test.expected, tlv, "Test %q", test.name)
	}
}

func TestNewP2PAdjacencyStateTLV(t *testing.T) {
	tests := []struct {
		name                   string
		adjacencyState         uint8
		extendedLocalCircuitID uint32
		expected               *P2PAdjacencyStateTLV
	}{
		{
			name:                   "Test #1",
			adjacencyState:         2,
			extendedLocalCircuitID: 100,
			expected: &P2PAdjacencyStateTLV{
				TLVType:                240,
				TLVLength:              5,
				AdjacencyState:         2,
				ExtendedLocalCircuitID: 100,
			},
		},
	}

	for _, test := range tests {
		tlv := NewP2PAdjacencyStateTLV(test.adjacencyState, test.extendedLocalCircuitID)

		assert.Equalf(t, test.expected, tlv, "Test %q", test.name)
	}
}

func TestP2PAdjacencyStateTLV(t *testing.T) {
	tlv := &P2PAdjacencyStateTLV{
		TLVType:   240,
		TLVLength: 15,
	}

	assert.Equal(t, uint8(240), tlv.Type())
	assert.Equal(t, uint8(15), tlv.Length())
	assert.Equal(t, P2PAdjacencyStateTLV{
		TLVType:   240,
		TLVLength: 15,
	}, tlv.Value())
}

func TestP2PAdjacencyStateTLVSerialize(t *testing.T) {
	tests := []struct {
		name     string
		input    *P2PAdjacencyStateTLV
		expected []byte
	}{
		{
			name: "Test #1",
			input: &P2PAdjacencyStateTLV{
				TLVType:                        240,
				TLVLength:                      15,
				AdjacencyState:                 1,
				ExtendedLocalCircuitID:         100,
				NeighborSystemID:               types.SystemID{1, 2, 3, 4, 5, 6},
				NeighborExtendedLocalCircuitID: 200,
			},
			expected: []byte{
				240,
				15,
				1,
				0, 0, 0, 100,
				1, 2, 3, 4, 5, 6,
				0, 0, 0, 200,
			},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		test.input.Serialize(buf)

		assert.Equalf(t, test.expected, buf.Bytes(), "Test %q", test.name)
	}
}
