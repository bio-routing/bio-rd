package packet

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPfxLen(t *testing.T) {
	tests := []struct {
		name     string
		e        *ExtendedIPReachability
		expected uint8
	}{
		{
			name: "Test #1",
			e: &ExtendedIPReachability{
				UDSubBitPfxLen: 32,
			},
			expected: 32,
		},
		{
			name: "Test #2",
			e: &ExtendedIPReachability{
				UDSubBitPfxLen: 96,
			},
			expected: 32,
		},
		{
			name: "Test #3",
			e: &ExtendedIPReachability{
				UDSubBitPfxLen: 24,
			},
			expected: 24,
		},
	}

	for _, test := range tests {
		res := test.e.PfxLen()
		assert.Equal(t, test.expected, res, test.name)
	}
}

func TestHasSubTLVs(t *testing.T) {
	tests := []struct {
		name     string
		e        *ExtendedIPReachability
		expected bool
	}{
		{
			name: "Test #1",
			e: &ExtendedIPReachability{
				UDSubBitPfxLen: 64,
			},
			expected: true,
		},
		{
			name: "Test #2",
			e: &ExtendedIPReachability{
				UDSubBitPfxLen: 23,
			},
			expected: false,
		},
		{
			name: "Test #3",
			e: &ExtendedIPReachability{
				UDSubBitPfxLen: 88, // /24 with Sub TLVs (+64)
			},
			expected: true,
		},
	}

	for _, test := range tests {
		res := test.e.hasSubTLVs()
		assert.Equal(t, test.expected, res, test.name)
	}
}

func TestReadExtendedIPReachabilityTLV(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantFail bool
		expected *ExtendedIPReachabilityTLV
	}{
		{
			name: "Single entry. No sub TLVs.",
			input: []byte{
				// First Extended IP Reach.
				0, 0, 0, 100, // Metric
				24,             // UDSubBitPfxLen (no sub TLVs)
				10, 20, 30, 40, // Address
			},
			expected: &ExtendedIPReachabilityTLV{
				TLVType:   135,
				TLVLength: 9,
				ExtendedIPReachabilities: []*ExtendedIPReachability{
					{
						Metric:         100,
						UDSubBitPfxLen: 24,
						Address:        169090600,
					},
				},
			},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		tlv, err := readExtendedIPReachabilityTLV(buf, 135, uint8(len(test.input)))
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
