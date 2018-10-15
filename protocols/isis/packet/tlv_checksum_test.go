package packet

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChecksumTLV(t *testing.T) {
	tlv := &ChecksumTLV{
		TLVType:   12,
		TLVLength: 2,
		Checksum:  123,
	}

	assert.Equal(t, uint8(12), tlv.Type())
	assert.Equal(t, uint8(2), tlv.Length())
	assert.Equal(t, ChecksumTLV{
		TLVType:   12,
		TLVLength: 2,
		Checksum:  123,
	}, tlv.Value())
}

func TestChecksumSerialize(t *testing.T) {
	tests := []struct {
		name     string
		input    *ChecksumTLV
		expected []byte
	}{
		{
			name: "A",
			input: &ChecksumTLV{
				TLVType:   12,
				TLVLength: 2,
				Checksum:  123,
			},
			expected: []byte{
				12, 2, 0, 123,
			},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		test.input.Serialize(buf)

		assert.Equalf(t, test.expected, buf.Bytes(), "Test %q", test.name)
	}
}

func TestReadChecksumTLV(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantFail bool
		expected *ChecksumTLV
	}{
		{
			name:     "Full",
			input:    []byte{0, 123},
			wantFail: false,
			expected: &ChecksumTLV{
				TLVType:   8,
				TLVLength: 2,
				Checksum:  123,
			},
		},
		{
			name:     "Incomplete",
			input:    []byte{0},
			wantFail: true,
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		tlv, err := readChecksumTLV(buf, 8, uint8(len(test.input)))
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
