package packet

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDynamicHostnameTLVSerialize(t *testing.T) {
	tests := []struct {
		name     string
		input    *DynamicHostNameTLV
		expected []byte
	}{
		{
			name: "Full",
			input: &DynamicHostNameTLV{
				TLVType:   137,
				TLVLength: 5,
				Hostname:  []byte{1, 2, 3, 4, 5},
			},
			expected: []byte{137, 5, 1, 2, 3, 4, 5},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		test.input.Serialize(buf)

		assert.Equalf(t, test.expected, buf.Bytes(), "Test %q", test.name)
	}
}

func TestReadDynamicHostnameTLV(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		tlvLen   uint8
		wantFail bool
		expected *DynamicHostNameTLV
	}{
		{
			name: "Full",
			input: []byte{
				1, 2, 3, 4, 5,
			},
			tlvLen: 5,
			expected: &DynamicHostNameTLV{
				TLVType:   137,
				TLVLength: 5,
				Hostname:  []byte{1, 2, 3, 4, 5},
			},
		},
		{
			name: "Incomplete",
			input: []byte{
				1, 2, 3, 4,
			},
			tlvLen:   5,
			wantFail: true,
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		tlv, err := readDynamicHostnameTLV(buf, 137, test.tlvLen)

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

func TestNewDynamicHostnameTLV(t *testing.T) {
	tlv := NewDynamicHostnameTLV([]byte("abcd"))

	expected := &DynamicHostNameTLV{
		TLVType:   137,
		TLVLength: 4,
		Hostname:  []byte("abcd"),
	}

	assert.Equal(t, expected, tlv)
}

func TestDynamicHostnameTLVType(t *testing.T) {
	tlv := &DynamicHostNameTLV{
		TLVType: 100,
	}

	assert.Equal(t, uint8(100), tlv.Type())
}

func TestDynamicHostnameTLVLength(t *testing.T) {
	tlv := &DynamicHostNameTLV{
		TLVLength: 123,
	}

	assert.Equal(t, uint8(123), tlv.Length())
}

func TestDynamicHostnameTLVValue(t *testing.T) {
	tlv := &DynamicHostNameTLV{
		TLVLength: 123,
	}

	assert.Equal(t, tlv, tlv.Value())
}
