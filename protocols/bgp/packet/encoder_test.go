package packet

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/taktv6/tflow2/convert"
)

func TestSerializeKeepaliveMsg(t *testing.T) {
	expected := []byte{
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0x00, 0x13, 0x04,
	}
	res := SerializeKeepaliveMsg()

	assert.Equal(t, expected, res)
}

func TestSerializeNotificationMsg(t *testing.T) {
	tests := []struct {
		name     string
		input    *BGPNotification
		expected []byte
	}{
		{
			name: "Valid #1",
			input: &BGPNotification{
				ErrorCode:    10,
				ErrorSubcode: 5,
			},
			expected: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0x00, 0x15, // Length
				0x03, // Type
				0x0a, // Error Code
				0x05, // Error Subcode
			},
		},
		{
			name: "Valid #2",
			input: &BGPNotification{
				ErrorCode:    11,
				ErrorSubcode: 6,
			},
			expected: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0x00, 0x15, // Length
				0x03, // Type
				0x0b, // Error Code
				0x06, // Error Subcode
			},
		},
	}

	for _, test := range tests {
		res := SerializeNotificationMsg(test.input)
		assert.Equal(t, test.expected, res)
	}
}

func TestSerializeOpenMsg(t *testing.T) {
	tests := []struct {
		name     string
		input    *BGPOpen
		expected []byte
	}{
		{
			name: "Valid #1",
			input: &BGPOpen{
				Version:       4,
				AS:            15169,
				HoldTime:      120,
				BGPIdentifier: convert.Uint32([]byte{100, 111, 120, 130}),
				OptParmLen:    0,
			},
			expected: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0x00, 0x1d, // Length
				0x01,       // Type
				0x04,       // Version
				0x3b, 0x41, // ASN
				0x00, 0x78, // Holdtime
				130, 120, 111, 100, // BGP Identifier
				0x00, // Opt. Param Length
			},
		},
	}

	for _, test := range tests {
		res := SerializeOpenMsg(test.input)
		assert.Equal(t, test.expected, res)
	}
}

/*func TestSerializeOptParams(t *testing.T) {
	tests := []struct {
		name      string
		optParams []OptParam
		expected  []byte
	}{
		{
			name:      "empty",
			optParams: []OptParam{},
			expected:  []byte{},
		},
		{
			name: "1 Option",
			optParams: []OptParam{
				{
					Type:   100,
					Length: 2,
					Value:  []byte{1, 2},
				},
			},
			expected: []byte{100, 2, 1, 2},
		},
		{
			name: "2 Options",
			optParams: []OptParam{
				{
					Type:   100,
					Length: 2,
					Value:  []byte{1, 2},
				},
				{
					Type:   200,
					Length: 2,
					Value:  []byte{3, 4},
				},
			},
			expected: []byte{100, 2, 1, 2, 200, 2, 3, 4},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(make([]byte, 0))
		serializeOptParams(buf, test.optParams)
		assert.Equal(t, test.expected, buf.Bytes())
	}
}*/

func TestSerializeHeader(t *testing.T) {
	tests := []struct {
		name     string
		length   uint16
		typ      uint8
		expected []byte
	}{
		{
			name:   "Valid #1",
			length: 10,
			typ:    5,
			expected: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0x00, 0x0a, 0x05,
			},
		},
		{
			name:   "Valid #12",
			length: 256,
			typ:    255,
			expected: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0x01, 0x00, 0xff,
			},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer([]byte{})
		serializeHeader(buf, test.length, test.typ)

		assert.Equal(t, test.expected, buf.Bytes())
	}
}
