package packet

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProtocolsSupportedTLV(t *testing.T) {
	tlv := &ProtocolsSupportedTLV{
		TLVType:                 12,
		TLVLength:               2,
		NetworkLayerProtocolIDs: []uint8{100, 200},
	}

	assert.Equal(t, uint8(12), tlv.Type())
	assert.Equal(t, uint8(2), tlv.Length())
	assert.Equal(t, ProtocolsSupportedTLV{
		TLVType:                 12,
		TLVLength:               2,
		NetworkLayerProtocolIDs: []uint8{100, 200},
	}, tlv.Value())
}

func TestProtocolsSupportedTLVSerialize(t *testing.T) {
	tests := []struct {
		name     string
		input    *ProtocolsSupportedTLV
		expected []byte
	}{
		{
			name: "A",
			input: &ProtocolsSupportedTLV{
				TLVType:                 129,
				TLVLength:               2,
				NetworkLayerProtocolIDs: []uint8{100, 200},
			},
			expected: []byte{
				129, 2, 100, 200,
			},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		test.input.Serialize(buf)

		assert.Equalf(t, test.expected, buf.Bytes(), "Test %q", test.name)
	}
}

func TestReadProtocolsSupportedTLV(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantFail bool
		expected *ProtocolsSupportedTLV
	}{
		{
			name:     "Full",
			input:    []byte{100, 200},
			wantFail: false,
			expected: &ProtocolsSupportedTLV{
				TLVType:                 129,
				TLVLength:               2,
				NetworkLayerProtocolIDs: []uint8{100, 200},
			},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		tlv, err := readProtocolsSupportedTLV(buf, 129, uint8(len(test.input)))
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
