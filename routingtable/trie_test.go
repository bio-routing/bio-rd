package routingtable

import (
	"testing"
)

func TestGetBitUint32(t *testing.T) {
	tests := []struct {
		name     string
		input    uint32
		offset   uint8
		expected bool
	}{
		{
			name:     "test 1",
			input:    167772160, // 10.0.0.0
			offset:   8,
			expected: false,
		},
		{
			name:     "test 2",
			input:    184549376, // 11.0.0.0
			offset:   8,
			expected: true,
		},
	}

	for _, test := range tests {
		b := getBitUint32(test.input, test.offset)
		if b != test.expected {
			t.Errorf("%s: Unexpected failure: Bit %d of %d is %v. Expected %v", test.name, test.offset, test.input, b, test.expected)
		}
	}
}
