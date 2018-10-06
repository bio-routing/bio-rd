package packet

import (
	"bytes"
	"testing"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/stretchr/testify/assert"
)

func TestNewAreaAddressesTLV(t *testing.T) {
	tests := []struct {
		name     string
		input    []types.AreaID
		expected *AreaAddressesTLV
	}{
		{
			name:  "Without Areas",
			input: []types.AreaID{},
			expected: &AreaAddressesTLV{
				TLVType:   1,
				TLVLength: 0,
				AreaIDs:   []types.AreaID{},
			},
		},
		{
			name: "With Areas",
			input: []types.AreaID{
				{
					1, 2, 3,
				},
			},
			expected: &AreaAddressesTLV{
				TLVType:   1,
				TLVLength: 4,
				AreaIDs: []types.AreaID{
					{
						1, 2, 3,
					},
				},
			},
		},
	}

	for _, test := range tests {
		tlv := NewAreaAddressesTLV(test.input)
		assert.Equalf(t, test.expected, tlv, "Test %q", test.name)
	}
}

func TestAreaAddressesTLV(t *testing.T) {
	tlv := NewAreaAddressesTLV([]types.AreaID{})

	assert.Equal(t, uint8(1), tlv.Type())
	assert.Equal(t, uint8(0), tlv.Length())
	assert.Equal(t, AreaAddressesTLV{
		TLVType:   1,
		TLVLength: 0,
		AreaIDs:   []types.AreaID{},
	}, tlv.Value())
}

func TestAreaAddressesTLVSerialize(t *testing.T) {
	tests := []struct {
		name     string
		input    *AreaAddressesTLV
		expected []byte
	}{
		{
			name: "Full",
			input: &AreaAddressesTLV{
				TLVType:   1,
				TLVLength: 0,
				AreaIDs:   []types.AreaID{},
			},
			expected: []byte{1, 0},
		},
		{
			name: "Full",
			input: &AreaAddressesTLV{
				TLVType:   8,
				TLVLength: 4,
				AreaIDs: []types.AreaID{
					{
						1, 2, 3,
					},
				},
			},
			expected: []byte{8, 4, 3, 1, 2, 3},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		test.input.Serialize(buf)

		assert.Equalf(t, test.expected, buf.Bytes(), "Test %q", test.name)
	}
}
