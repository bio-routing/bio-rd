package packet

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTrafficEngineeringRouterIDTLV(t *testing.T) {
	tests := []struct {
		name     string
		addr     [4]byte
		expected *TrafficEngineeringRouterIDTLV
	}{
		{
			name: "Test #1",
			addr: [4]byte{10, 20, 30, 40},
			expected: &TrafficEngineeringRouterIDTLV{
				TLVType:   134,
				TLVLength: 4,
				Address:   [4]byte{10, 20, 30, 40},
			},
		},
	}

	for _, test := range tests {
		tlv := NewTrafficEngineeringRouterIDTLV(test.addr)
		assert.Equal(t, test.expected, tlv, test.name)
	}
}

func TestTrafficEngineeringRouterIDTLVSerialize(t *testing.T) {
	tests := []struct {
		name     string
		tlv      *TrafficEngineeringRouterIDTLV
		expected []byte
	}{
		{
			name: "Test #1",
			tlv: &TrafficEngineeringRouterIDTLV{
				TLVType:   134,
				TLVLength: 4,
				Address:   [4]byte{10, 0, 0, 123},
			},
			expected: []byte{134, 4, 10, 0, 0, 123},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		test.tlv.Serialize(buf)
		assert.Equal(t, test.expected, buf.Bytes(), test.name)
	}
}

func TestReadTrafficEngineeringRouterIDTLV(t *testing.T) {
	tests := []struct {
		name      string
		tlvType   uint8
		tlvLength uint8
		pkt       []byte
		expected  *TrafficEngineeringRouterIDTLV
		wantFail  bool
	}{
		{
			name:      "Normal packet",
			tlvType:   134,
			tlvLength: 4,
			pkt: []byte{
				1, 2, 3, 4,
			},
			expected: &TrafficEngineeringRouterIDTLV{
				TLVType:   134,
				TLVLength: 4,
				Address:   [4]byte{1, 2, 3, 4},
			},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.pkt)
		tlv, err := readTrafficEngineeringRouterIDTLV(buf, test.tlvType, test.tlvLength)
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

		assert.Equal(t, test.expected, tlv, test.name)
	}
}
