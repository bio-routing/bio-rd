package packet

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeRouteMonitoringMsg(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		ch       *CommonHeader
		wantFail bool
		expected *RouteMonitoringMsg
	}{
		{
			name: "Full",
			input: []byte{
				1,
				2,
				0, 0, 0, 3,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 0, 200, 124,
				0, 0, 0, 123,
				0, 0, 0, 100,
				0, 0, 0, 200,

				100, 110, 120,
			},
			ch: &CommonHeader{
				MsgLength: CommonHeaderLen + PerPeerHeaderLen + 3,
			},
			wantFail: false,
			expected: &RouteMonitoringMsg{
				CommonHeader: &CommonHeader{
					MsgLength: CommonHeaderLen + PerPeerHeaderLen + 3,
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
				BGPUpdate: []byte{
					100, 110, 120,
				},
			},
		},
		{
			name: "Incomplete per peer header",
			input: []byte{
				1,
				2,
				0, 0, 0, 3,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
			},
			ch: &CommonHeader{
				MsgLength: CommonHeaderLen + PerPeerHeaderLen + 3,
			},
			wantFail: true,
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		r, err := decodeRouteMonitoringMsg(buf, test.ch)
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

		assert.Equalf(t, test.expected, r, "Test %q", test.name)
	}
}
