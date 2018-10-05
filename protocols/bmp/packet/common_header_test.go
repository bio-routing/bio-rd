package packet

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommonHeaderSerialize(t *testing.T) {
	tests := []struct {
		name     string
		input    *CommonHeader
		expected []byte
	}{
		{
			name: "Test #1",
			input: &CommonHeader{
				Version:   3,
				MsgLength: 100,
				MsgType:   10,
			},
			expected: []byte{3, 0, 0, 0, 100, 10},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		test.input.Serialize(buf)
		assert.Equalf(t, test.expected, buf.Bytes(), "Test %q", test.name)
	}
}

func TestDecodeCommonHeader(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantFail bool
		expected *CommonHeader
	}{
		{
			name: "Full packet",
			input: []byte{
				3, 0, 0, 0, 100, 10,
			},
			wantFail: false,
			expected: &CommonHeader{
				Version:   3,
				MsgLength: 100,
				MsgType:   10,
			},
		},
		{
			name: "Incomplete",
			input: []byte{
				3, 0, 0, 0, 100,
			},
			wantFail: true,
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		ch, err := decodeCommonHeader(buf)
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

		assert.Equalf(t, test.expected, ch, "Test %q", test.name)
	}
}
