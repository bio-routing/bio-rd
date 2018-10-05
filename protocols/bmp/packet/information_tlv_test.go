package packet

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeInformationTLV(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantFail bool
		expected *InformationTLV
	}{
		{
			name: "Full",
			input: []byte{
				0, 10, 0, 5,
				1, 2, 3, 4, 5,
			},
			wantFail: false,
			expected: &InformationTLV{
				InformationType:   10,
				InformationLength: 5,
				Information:       []byte{1, 2, 3, 4, 5},
			},
		},
		{
			name: "Incomplete",
			input: []byte{
				0, 10, 0, 5,
				1, 2, 3, 4,
			},
			wantFail: true,
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		infoTLV, err := decodeInformationTLV(buf)
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

		assert.Equalf(t, test.expected, infoTLV, "Test %q", test.name)
	}
}
