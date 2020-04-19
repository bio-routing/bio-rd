package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNetStringToByteSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []byte
		wantFail bool
	}{
		{
			name:     "Test #1",
			input:    "49.0001.0100.0000.0001.00",
			expected: []byte{0x49, 0, 1, 1, 0, 0, 0, 0, 1, 0},
		},
		{
			name:     "Test #1",
			input:    "49.000g.0100.0000.0001.00",
			wantFail: true,
		},
	}

	for _, test := range tests {
		res, err := netStringToByteSlice(test.input)
		if test.wantFail {
			if err == nil {
				t.Errorf("Unexpected success for test %s", test.name)
				continue
			}
		} else {
			if err != nil {
				t.Errorf("Expected failure for test %s: %v", test.name, err)
				continue
			}
		}

		assert.Equal(t, test.expected, res, test.name)
	}
}
