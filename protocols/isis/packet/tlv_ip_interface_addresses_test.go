package packet

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewIPInterfaceAddressTLV(t *testing.T) {
	tests := []struct {
		name     string
		addrs    []uint32
		expected *IPInterfaceAddressesTLV
	}{
		{
			name:  "Test #1",
			addrs: []uint32{100},
			expected: &IPInterfaceAddressesTLV{
				TLVType:       132,
				TLVLength:     4,
				IPv4Addresses: []uint32{100},
			},
		},
	}

	for _, test := range tests {
		tlv := NewIPInterfaceAddressesTLV(test.addrs)
		assert.Equalf(t, test.expected, tlv, "Test %q", test.name)
	}
}

func TestReadIPInterfaceAddressTLV(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		tlvLength uint8
		wantFail  bool
		expected  *IPInterfaceAddressesTLV
	}{
		{
			name: "Full",
			input: []byte{
				0, 0, 0, 100,
			},
			tlvLength: 4,
			expected: &IPInterfaceAddressesTLV{
				TLVType:       132,
				TLVLength:     4,
				IPv4Addresses: []uint32{100},
			},
		},
		{
			name: "Incomplete",
			input: []byte{
				0, 0, 0,
			},
			tlvLength: 4,
			wantFail:  true,
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		tlv, err := readIPInterfaceAddressesTLV(buf, 132, test.tlvLength)

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
