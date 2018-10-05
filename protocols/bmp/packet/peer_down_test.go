package packet

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodePeerDownNotification(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		ch       *CommonHeader
		wantFail bool
		expected *PeerDownNotification
	}{
		{
			name: "Full",
			input: []byte{
				1,
				1, 2, 3,
			},
			ch: &CommonHeader{
				MsgLength: CommonHeaderLen + 4,
			},
			wantFail: false,
			expected: &PeerDownNotification{
				Reason: 1,
				Data: []byte{
					1, 2, 3,
				},
			},
		},
		{
			name: "Full no data",
			input: []byte{
				4,
			},
			ch: &CommonHeader{
				MsgLength: CommonHeaderLen + 4,
			},
			wantFail: false,
			expected: &PeerDownNotification{
				Reason: 4,
				Data:   nil,
			},
		},
		{
			name: "Incomplete data",
			input: []byte{
				1,
				1, 2, 3,
			},
			ch: &CommonHeader{
				MsgLength: CommonHeaderLen + 5,
			},
			wantFail: true,
		},
		{
			name:  "Incomplete",
			input: []byte{},
			ch: &CommonHeader{
				MsgLength: CommonHeaderLen + 5,
			},
			wantFail: true,
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		p, err := decodePeerDownNotification(buf, test.ch)
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

		assert.Equalf(t, test.expected, p, "Test %q", test.name)
	}
}
