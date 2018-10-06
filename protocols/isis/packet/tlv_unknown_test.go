package packet

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadUnknownTLV(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		tlvType   uint8
		tlvLength uint8
		wantFail  bool
		expected  *UnknownTLV
	}{
		{
			name:      "Full",
			input:     []byte{1, 1, 1},
			tlvType:   100,
			tlvLength: 3,
			wantFail:  false,
			expected: &UnknownTLV{
				TLVType:   100,
				TLVLength: 3,
				TLVValue:  []byte{1, 1, 1},
			},
		},
		{
			name:      "Incomplete",
			input:     []byte{1, 1},
			tlvType:   100,
			tlvLength: 3,
			wantFail:  true,
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		tlv, err := readUnknownTLV(buf, test.tlvType, test.tlvLength)

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

func TestUnknownTLVSerialize(t *testing.T) {
	tests := []struct {
		name     string
		input    *UnknownTLV
		expected []byte
	}{
		{
			name: "Full",
			input: &UnknownTLV{
				TLVType:   100,
				TLVLength: 3,
				TLVValue:  []byte{1, 2, 3},
			},
			expected: []byte{
				100, 3, 1, 2, 3,
			},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		test.input.Serialize(buf)
		assert.Equalf(t, test.expected, buf.Bytes(), "Test %q", test.name)
	}
}

func TestUnknownTLV(t *testing.T) {
	tlv := &UnknownTLV{
		TLVType:   100,
		TLVLength: 1,
		TLVValue:  []byte{1},
	}

	assert.Equal(t, uint8(100), tlv.Type())
	assert.Equal(t, uint8(1), tlv.Length())
	assert.Equal(t, &UnknownTLV{
		TLVType:   100,
		TLVLength: 1,
		TLVValue:  []byte{1},
	}, tlv.Value())
}
