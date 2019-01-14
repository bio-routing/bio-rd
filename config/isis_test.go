package config

import (
	"testing"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/stretchr/testify/assert"
)

func TestParseNET(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantFail bool
		expected *NET
	}{
		{
			name:     "Too short",
			input:    []byte{49, 1, 2, 3, 4, 5, 0},
			wantFail: true,
		},
		{
			name:     "Too long",
			input:    []byte{0x49, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 1, 2, 3, 4, 5, 6, 0, 0},
			wantFail: true,
		},
		{
			name:  "Max area ID length",
			input: []byte{0x49, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 1, 2, 3, 4, 5, 6, 0},
			expected: &NET{
				AFI:      0x49,
				AreaID:   []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
				SystemID: types.SystemID{1, 2, 3, 4, 5, 6},
				SEL:      0x00,
			},
		},
		{
			name:  "No Area ID",
			input: []byte{0x49, 1, 2, 3, 4, 5, 6, 0},
			expected: &NET{
				AFI:      0x49,
				AreaID:   []byte{},
				SystemID: types.SystemID{1, 2, 3, 4, 5, 6},
				SEL:      0x00,
			},
		},
	}

	for _, test := range tests {
		NET, err := parseNET(test.input)
		if err != nil {
			if test.wantFail {
				continue
			}

			t.Errorf("Unexpected error for test %q: %v", test.name, err)
			continue
		}

		if test.wantFail {
			t.Errorf("Unexpected success for test %q", test.name)
		}

		assert.Equalf(t, test.expected, NET, "Test: %q", test.name)

	}
}
