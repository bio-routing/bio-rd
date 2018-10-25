package packet

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPeerDownMsgType(t *testing.T) {
	pd := &PeerDownNotification{
		CommonHeader: &CommonHeader{
			MsgType: 100,
		},
	}

	if pd.MsgType() != 100 {
		t.Errorf("Unexpected result")
	}
}

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
				2,
				0, 0, 0, 0, 0, 0, 0, 3,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 0, 200, 124,
				0, 0, 0, 123,
				0, 0, 0, 100,
				0, 0, 0, 200,

				1,
				1, 2, 3,
			},
			ch: &CommonHeader{
				MsgLength: CommonHeaderLen + 4 + PerPeerHeaderLen,
			},
			wantFail: false,
			expected: &PeerDownNotification{
				CommonHeader: &CommonHeader{
					MsgLength: CommonHeaderLen + 4 + PerPeerHeaderLen,
				},
				PerPeerHeader: &PerPeerHeader{
					PeerType:              1,
					PeerFlags:             2,
					PeerDistinguisher:     3,
					PeerAddress:           [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
					PeerAS:                51324,
					PeerBGPID:             123,
					Timestamp:             100,
					TimestampMicroSeconds: 200,
				},
				Reason: 1,
				Data: []byte{
					1, 2, 3,
				},
			},
		},
		{
			name: "Full no data",
			input: []byte{
				1,
				2,
				0, 0, 0, 0, 0, 0, 0, 3,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 0, 200, 124,
				0, 0, 0, 123,
				0, 0, 0, 100,
				0, 0, 0, 200,
				4,
			},
			ch: &CommonHeader{
				MsgLength: CommonHeaderLen + PerPeerHeaderLen + 4,
			},
			wantFail: false,
			expected: &PeerDownNotification{
				CommonHeader: &CommonHeader{
					MsgLength: CommonHeaderLen + PerPeerHeaderLen + 4,
				},
				PerPeerHeader: &PerPeerHeader{
					PeerType:              1,
					PeerFlags:             2,
					PeerDistinguisher:     3,
					PeerAddress:           [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
					PeerAS:                51324,
					PeerBGPID:             123,
					Timestamp:             100,
					TimestampMicroSeconds: 200,
				},
				Reason: 4,
				Data:   nil,
			},
		},
		{
			name: "Incomplete per peer header",
			input: []byte{
				1,
				2,
				0, 0, 0, 0, 0, 0, 0, 3,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 0, 200, 124,
				0, 0, 0,
			},
			ch: &CommonHeader{
				MsgLength: CommonHeaderLen + 5,
			},
			wantFail: true,
		},
		{
			name: "Incomplete data",
			input: []byte{
				1,
				2,
				0, 0, 0, 0, 0, 0, 0, 3,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 0, 200, 124,
				0, 0, 0, 123,
				0, 0, 0, 100,
				0, 0, 0, 200,

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
