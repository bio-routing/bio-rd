package packet

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeaderEncode(t *testing.T) {
	tests := []struct {
		name     string
		input    *isisHeader
		expected []byte
	}{
		{
			name: "Test #1",
			input: &isisHeader{
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
		test.input.serialize(buf)
		res := buf.Bytes()
		assert.Equalf(t, test.expected, res, "%s failed", test.name)
	}
}
