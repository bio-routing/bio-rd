package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseNET(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected *NET
		wantFail bool
	}{
		{
			name: "Simple long valid NET",
			input: []byte{
				0x49, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Area (including AFI)
				0x01, 0x00, 0x00, 0x00, 0x00, 0x01, // SysID
				0x00, // SEL
			},
			expected: &NET{
				AreaID: []byte{
					0x49, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				},
				SystemID: SystemID{
					0x01, 0x00, 0x00, 0x00, 0x00, 0x01,
				},
				SEL: 0x00,
			},
		},
		{
			name: "Simple short valid NET",
			input: []byte{
				0x49,                               // Area (including AFI)
				0x01, 0x00, 0x00, 0x00, 0x00, 0x01, // SysID
				0x00, // SEL
			},
			expected: &NET{
				AreaID: []byte{
					0x49,
				},
				SystemID: SystemID{
					0x01, 0x00, 0x00, 0x00, 0x00, 0x01,
				},
				SEL: 0x00,
			},
		},
		{
			name: "Too long NET",
			input: []byte{
				0x49, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, // Area (including AFI)
				0x01, 0x00, 0x00, 0x00, 0x00, 0x01, // SysID
				0x00, // SEL
			},
			wantFail: true,
		},
		{
			name: "Too short NET",
			input: []byte{
				0x01, 0x00, 0x00, 0x00, 0x00, 0x01, // SysID
				0x00, // SEL
			},
			wantFail: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			n, err := ParseNET([]byte(test.input))
			if err != nil && !test.wantFail {
				t.Errorf("unexpected failure for test %q: %v", test.name, err)
				return
			}

			if err == nil && test.wantFail {
				t.Errorf("unexpected success for test %q", test.name)
				return
			}

			assert.Equal(t, test.expected, n, test.name)
		})
	}
}
