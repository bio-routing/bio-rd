package packet

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPeerUpMsgType(t *testing.T) {
	pd := &PeerUpNotification{
		CommonHeader: &CommonHeader{
			MsgType: 100,
		},
	}

	if pd.MsgType() != 100 {
		t.Errorf("Unexpected result")
	}
}

func TestDecodePeerUp(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		ch       *CommonHeader
		wantFail bool
		expected *PeerUpNotification
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

				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 100,
				0, 200,

				// OPEN Sent
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
				0, 34,
				1,
				4,    // Version
				1, 0, // ASN
				2, 0, // Hold Time
				100, 110, 120, 130, // BGP Identifier
				5, // Opt Parm Len
				1, 2, 3, 4, 5,

				// OPEN Recv
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
				0, 29,
				1,
				4,    // Version
				1, 0, // ASN
				2, 0, // Hold Time
				100, 110, 120, 130, // BGP Identifier
				0, // Opt Parm Len

				120, 140, 160, // Information
			},
			ch: &CommonHeader{
				MsgLength: 126,
			},
			wantFail: false,
			expected: &PeerUpNotification{
				CommonHeader: &CommonHeader{
					MsgLength: 126,
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
				LocalAddress: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
				LocalPort:    100,
				RemotePort:   200,
				SentOpenMsg: []byte{
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					0, 34,
					1,
					4,    // Version
					1, 0, // ASN
					2, 0, // Hold Time
					100, 110, 120, 130, // BGP Identifier
					5, // Opt Parm Len
					1, 2, 3, 4, 5,
				},
				ReceivedOpenMsg: []byte{
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					0, 29,
					1,
					4,    // Version
					1, 0, // ASN
					2, 0, // Hold Time
					100, 110, 120, 130, // BGP Identifier
					0, // Opt Parm Len
				},
				Information: []byte{
					120, 140, 160, // Information
				},
			},
		},
		{
			name: "Full #2",
			input: []byte{
				1,
				2,
				0, 0, 0, 0, 0, 0, 0, 3,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 0, 200, 124,
				0, 0, 0, 123,
				0, 0, 0, 100,
				0, 0, 0, 200,

				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 100,
				0, 200,

				// OPEN Sent
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
				0, 34,
				1,
				4,    // Version
				1, 0, // ASN
				2, 0, // Hold Time
				100, 110, 120, 130, // BGP Identifier
				5, // Opt Parm Len
				1, 2, 3, 4, 5,

				// OPEN Recv
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
				0, 29,
				1,
				4,    // Version
				1, 0, // ASN
				2, 0, // Hold Time
				100, 110, 120, 130, // BGP Identifier
				0, // Opt Parm Len
			},
			ch: &CommonHeader{
				MsgLength: 82,
			},
			wantFail: false,
			expected: &PeerUpNotification{
				CommonHeader: &CommonHeader{
					MsgLength: 82,
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
				LocalAddress: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
				LocalPort:    100,
				RemotePort:   200,
				SentOpenMsg: []byte{
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					0, 34,
					1,
					4,    // Version
					1, 0, // ASN
					2, 0, // Hold Time
					100, 110, 120, 130, // BGP Identifier
					5, // Opt Parm Len
					1, 2, 3, 4, 5,
				},
				ReceivedOpenMsg: []byte{
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					0, 29,
					1,
					4,    // Version
					1, 0, // ASN
					2, 0, // Hold Time
					100, 110, 120, 130, // BGP Identifier
					0, // Opt Parm Len
				},
			},
		},
		{
			name: "Incomplete #0",
			input: []byte{
				1,
				2,
				0, 0, 0, 0, 0, 0, 0, 3,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 0, 200, 124,
				0, 0, 0,
			},
			ch: &CommonHeader{
				MsgLength: 51,
			},
			wantFail: true,
		},
		{
			name: "Incomplete #1",
			input: []byte{
				1,
				2,
				0, 0, 0, 0, 0, 0, 0, 3,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 0, 200, 124,
				0, 0, 0, 123,
				0, 0, 0, 100,
				0, 0, 0, 200,

				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 100,
			},
			ch: &CommonHeader{
				MsgLength: 51,
			},
			wantFail: true,
		},
		{
			name: "Incomplete #2",
			input: []byte{
				1,
				2,
				0, 0, 0, 0, 0, 0, 0, 3,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 0, 200, 124,
				0, 0, 0, 123,
				0, 0, 0, 100,
				0, 0, 0, 200,

				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 100,
				0, 200,

				// OPEN Sent
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
				0, 29,
				1,
				4,    // Version
				1, 0, // ASN
				2, 0, // Hold Time
			},
			ch: &CommonHeader{
				MsgLength: 89,
			},
			wantFail: true,
		},
		{
			name: "Incomplete #3",
			input: []byte{
				1,
				2,
				0, 0, 0, 0, 0, 0, 0, 3,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 0, 200, 124,
				0, 0, 0, 123,
				0, 0, 0, 100,
				0, 0, 0, 200,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 100,
				0, 200,

				// OPEN Sent
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
				0, 29,
				1,
				4,    // Version
				1, 0, // ASN
				2, 0, // Hold Time
			},
			ch: &CommonHeader{
				MsgLength: 88,
			},
			wantFail: true,
		},
		{
			name: "Incomplete #4",
			input: []byte{
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 100,
				0, 200,

				// OPEN Sent
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
				0, 29,
				1,
				4,    // Version
				1, 0, // ASN
				2, 0, // Hold Time
				100, 110, 120, 130, // BGP Identifier
				5, // Opt Parm Len
				1, 2, 3, 4, 5,

				// OPEN Recv
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
				0, 29,
				1,
				4,    // Version
				1, 0, // ASN
				2, 0, // Hold Time
				100, 110, 120, 130, // BGP Identifier
				3, // Opt Parm Len
			},
			ch: &CommonHeader{
				MsgLength: 85,
			},
			wantFail: true,
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		pu, err := decodePeerUpNotification(buf, test.ch)
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

		assert.Equalf(t, test.expected, pu, "Test %q", test.name)
	}
}

func TestSerializePeerUpNotification(t *testing.T) {
	tests := []struct {
		name     string
		pun      *PeerUpNotification
		expected []byte
	}{
		{
			name: "Test #1",
			pun: &PeerUpNotification{
				CommonHeader: &CommonHeader{
					Version: 1,
					MsgType: PeerDownNotificationType,
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
				LocalAddress: [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 1, 2, 3},
				LocalPort:    179,
				RemotePort:   54321,
				SentOpenMsg: []byte{
					0xff, 0xff,
				},
				ReceivedOpenMsg: []byte{
					0xfe, 0xfe,
				},
				Information: []byte{
					0xaa, 0xbb,
				},
			},
			expected: []byte{
				0x1,
				0x0, 0x0, 0x0, 0x4a,
				0x2,
				0x0, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
				0x0, 0x17, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
				0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xa, 0x1,
				0x2, 0x3, 0x0, 0x0, 0x34, 0x17, 0x0, 0x0,
				0x5, 0x39, 0x63, 0x1f, 0x4c, 0x60, 0x0, 0x0,
				0x0, 0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
				0x0, 0x0, 0x0, 0x0, 0x0, 0xa, 0x1, 0x2, 0x3,
				0x0, 0xb3, 0xd4, 0x31,
				0xff, 0xff,
				0xfe, 0xfe,
				0xaa, 0xbb,
			},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		test.pun.Serialize(buf)
		assert.Equal(t, test.expected, buf.Bytes(), test.name)
	}
}
