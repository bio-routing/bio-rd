package packet

import (
	"bytes"
	"testing"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/stretchr/testify/assert"
)

func TestIsNeighborsTLV(t *testing.T) {
	tlv := &ISNeighborsTLV{
		TLVType:      6,
		TLVLength:    6,
		NeighborSNPA: types.SystemID{1, 2, 3, 4, 5, 6},
	}

	assert.Equal(t, uint8(6), tlv.Type())
	assert.Equal(t, uint8(6), tlv.Length())
	assert.Equal(t, ISNeighborsTLV{
		TLVType:      6,
		TLVLength:    6,
		NeighborSNPA: types.SystemID{1, 2, 3, 4, 5, 6},
	}, tlv.Value())
}

func TestReadISNeighborsTLV(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		tlvLength uint8
		wantFail  bool
		expected  *ISNeighborsTLV
	}{
		{
			name:      "Full",
			input:     []byte{1, 2, 3, 4, 5, 6},
			tlvLength: 6,
			wantFail:  false,
			expected: &ISNeighborsTLV{
				TLVType:      6,
				TLVLength:    6,
				NeighborSNPA: types.SystemID{1, 2, 3, 4, 5, 6},
			},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		tlv, err := readISNeighborsTLV(buf, 6, test.tlvLength)

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

func TestISNeighborsTLVSerialize(t *testing.T) {
	tests := []struct {
		name     string
		input    *ISNeighborsTLV
		expected []byte
	}{
		{
			name: "Test #1",
			input: &ISNeighborsTLV{
				TLVType:      6,
				TLVLength:    6,
				NeighborSNPA: types.SystemID{1, 2, 3, 4, 5, 6},
			},
			expected: []byte{6, 6, 1, 2, 3, 4, 5, 6},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		test.input.Serialize(buf)

		assert.Equalf(t, test.expected, buf.Bytes(), "Test %q", test.name)
	}
}
