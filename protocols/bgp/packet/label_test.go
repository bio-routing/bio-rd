package packet

import (
	"bytes"
	"testing"

	"github.com/bio-routing/tflow2/convert"
	"github.com/stretchr/testify/assert"
)

func TestLabelIsBottomOfStack(t *testing.T) {
	tests := []struct {
		name     string
		label    LabelStackEntry
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
		expected LabelStackEntry
	}{
		{
			name:     "Test #1",
			input:    []byte{0x49, 0x33, 0x01},
			expected: LabelStackEntry(0x00493301),
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		res, err := decodeLabelStackEntry(buf)
		if err != nil {
			t.Errorf("decodeLabel failed with %v", err)
		}

		assert.Equal(t, test.expected, res, test.name)
	}
}

func TestGetLabel(t *testing.T) {
	tests := []struct {
		name     string
		input    LabelStackEntry
		expected uint32
	}{
		{
			name: "Test #1",
			input: LabelStackEntry(convert.Uint32b([]byte{
				0x00, 0x49, 0x33, 0x01,
			})),
			expected: 299824,
		},
		{
			name: "Test #2",
			input: LabelStackEntry(convert.Uint32b([]byte{
				0x00, 0x49, 0x33, 0x00,
			})),
			expected: 299824,
		},
	}

	for _, test := range tests {
		res := test.input.GetLabel()
		assert.Equal(t, test.expected, res, test.name)
	}
}
