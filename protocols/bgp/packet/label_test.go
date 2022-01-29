package packet

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLabelIsBottomOfStack(t *testing.T) {
	tests := []struct {
		name     string
		label    Label
		expected bool
	}{
		{
			name:     "Test #1",
			label:    0x123401,
			expected: true,
		},
		{
			name:     "Test #2",
			label:    0x123402,
			expected: false,
		},
		{
			name:     "Test #3",
			label:    0x123404,
			expected: false,
		},
	}

	for _, test := range tests {
		res := test.label.isBottomOfStack()
		assert.Equal(t, test.expected, res, test.name)
	}
}

func TestDecodeLabel(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected Label
	}{
		{
			name:     "Test #1",
			input:    []byte{0x49, 0x33, 0x01},
			expected: Label(0x00493301),
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		res, err := decodeLabel(buf)
		if err != nil {
			t.Errorf("decodeLabel failed with %v", err)
		}

		assert.Equal(t, test.expected, res, test.name)
	}
}
