package packet

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeaderDecode(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantFail bool
		expected *ISISHeader
	}{
		{
			name: "Full",
			input: []byte{
				0, 0, 0, // SNAP
				0x83,
				27,
				0,
				0,
				16,
				1,
				0,
				0,
			},
			wantFail: false,
			expected: &ISISHeader{
				ProtoDiscriminator:  0x83,
				LengthIndicator:     27,
				ProtocolIDExtension: 0,
				IDLength:            0,
				PDUType:             16,
				Version:             1,
				MaxAreaAddresses:    0,
			},
		},
		{
			name: "Partial",
			input: []byte{
				0, 0, 0, // SNAP
				0x83,
				27,
				0,
				0,
				16,
				1,
				0,
			},
			wantFail: true,
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		hdr, err := DecodeHeader(buf)
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

		assert.Equalf(t, test.expected, hdr, "Test %q", test.name)
	}
}

func TestHeaderSerialize(t *testing.T) {
	tests := []struct {
		name     string
		input    *ISISHeader
		expected []byte
	}{
		{
			name: "Test #1",
			input: &ISISHeader{
				ProtoDiscriminator:  0x83,
				LengthIndicator:     27,
				ProtocolIDExtension: 0,
				IDLength:            0,
				PDUType:             16,
				Version:             1,
				MaxAreaAddresses:    0,
			},
			expected: []byte{
				0x83,
				27,
				0,
				0,
				16,
				1,
				0,
				0,
			},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		test.input.Serialize(buf)
		res := buf.Bytes()
		assert.Equalf(t, test.expected, res, "%s failed", test.name)
	}
}
