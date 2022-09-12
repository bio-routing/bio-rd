package packet

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouteMonitoringMsgType(t *testing.T) {
	pd := &RouteMonitoringMsg{
		CommonHeader: &CommonHeader{
			MsgType: 100,
		},
	}

	if pd.MsgType() != 100 {
		t.Errorf("Unexpected result")
	}
}

func TestSerializeRouteMonitoringMsg(t *testing.T) {
	tests := []struct {
		name     string
		rm       *RouteMonitoringMsg
		expected []byte
	}{
		{
			name: "Test case #1",
			rm: &RouteMonitoringMsg{
				CommonHeader: &CommonHeader{
					Version: 1,
					MsgType: RouteMonitoringType,
				},
				PerPeerHeader: &PerPeerHeader{
					PeerType:              0,
					PeerFlags:             0b10000000,
					PeerDistinguisher:     23,
					PeerAddress:           [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 1, 2, 3},
					PeerAS:                13335,
					PeerBGPID:             1337,
					Timestamp:             1662995552,
					TimestampMicroSeconds: 42,
				},
				BGPUpdate: []byte{
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
					0x0, 0x0, // Withdraw length
					0x0, 0x0, // Path attributes length
				},
			},
			expected: []byte{
				// Common header
				0x1,                 // Version
				0x0, 0x0, 0x0, 0x44, // Length
				0x0, // Type

				// Per peer header
				0x0,                                     // peer type
				0x80,                                    // peer flags
				0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x17, // Peer Distinguisher
				0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xa, 0x1, 0x2, 0x3, // Peer address
				0x0, 0x0, 0x34, 0x17, // Peer AS
				0x0, 0x0, 0x5, 0x39, // Peer BGP ID
				0x63, 0x1f, 0x4c, 0x60, // Timestamp seconds
				0x0, 0x0, 0x0, 0x2a, // Timestamp microseconds

				// BGP Update
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0x0, 0x0, 0x0, 0x0,
			},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		test.rm.Serialize(buf)
		assert.Equal(t, test.expected, buf.Bytes(), test.name)
	}
}

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
				0, 0, 0, 0, 0, 0, 0, 3,
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
