package packet

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPaddingTLV(t *testing.T) {
	tests := []struct {
		name     string
		input    uint8
		expected *PaddingTLV
	}{
		{
			name:  "A",
			input: 2,
			expected: &PaddingTLV{
				TLVType:     8,
				TLVLength:   2,
				PaddingData: []byte{0, 0},
			},
		},
		{
			name:  "B",
			input: 4,
			expected: &PaddingTLV{
				TLVType:     8,
				TLVLength:   4,
				PaddingData: []byte{0, 0, 0, 0},
			},
		},
	}

	for _, test := range tests {
		tlv := NewPaddingTLV(test.input)
		assert.Equalf(t, test.expected, tlv, "Test %q", test.name)
	}
}

func TestPaddingTLV(t *testing.T) {
	tlv := NewPaddingTLV(2)

	assert.Equal(t, uint8(8), tlv.Type())
	assert.Equal(t, uint8(2), tlv.Length())
	assert.Equal(t, &PaddingTLV{
		TLVType:     8,
		TLVLength:   2,
		PaddingData: []byte{0, 0},
	}, tlv.Value())
}

func TestPaddingTLVSerialize(t *testing.T) {
	tests := []struct {
		name     string
		input    *PaddingTLV
		expected []byte
	}{
		{
			name: "Full",
			input: &PaddingTLV{
				TLVType:     8,
				TLVLength:   2,
				PaddingData: []byte{0, 0},
			},
			expected: []byte{8, 2, 0, 0},
		},
		{
			name: "Full",
			input: &PaddingTLV{
				TLVType:     8,
				TLVLength:   3,
				PaddingData: []byte{0, 0, 0},
			},
			expected: []byte{8, 3, 0, 0, 0},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		test.input.Serialize(buf)

		assert.Equalf(t, test.expected, buf.Bytes(), "Test %q", test.name)
	}
}
