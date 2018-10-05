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
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 100,
				0, 200,

				// OPEN Sent
				4,    // Version
				1, 0, // ASN
				2, 0, // Hold Time
				100, 110, 120, 130, // BGP Identifier
				5, // Opt Parm Len
				1, 2, 3, 4, 5,

				// OPEN Recv
				4,    // Version
				1, 0, // ASN
				2, 0, // Hold Time
				100, 110, 120, 130, // BGP Identifier
				0, // Opt Parm Len

				120, 140, 160, // Information
			},
			ch: &CommonHeader{
				MsgLength: 47,
			},
			wantFail: false,
			expected: &PeerUpNotification{
				LocalAddress: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
				LocalPort:    100,
				RemotePort:   200,
				SentOpenMsg: []byte{
					4,    // Version
					1, 0, // ASN
					2, 0, // Hold Time
					100, 110, 120, 130, // BGP Identifier
					5, // Opt Parm Len
					1, 2, 3, 4, 5,
				},
				ReceivedOpenMsg: []byte{
					// OPEN Recv
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
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 100,
				0, 200,

				// OPEN Sent
				4,    // Version
				1, 0, // ASN
				2, 0, // Hold Time
				100, 110, 120, 130, // BGP Identifier
				5, // Opt Parm Len
				1, 2, 3, 4, 5,

				// OPEN Recv
				4,    // Version
				1, 0, // ASN
				2, 0, // Hold Time
				100, 110, 120, 130, // BGP Identifier
				0, // Opt Parm Len
			},
			ch: &CommonHeader{
				MsgLength: 44,
			},
			wantFail: false,
			expected: &PeerUpNotification{
				LocalAddress: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
				LocalPort:    100,
				RemotePort:   200,
				SentOpenMsg: []byte{
					4,    // Version
					1, 0, // ASN
					2, 0, // Hold Time
					100, 110, 120, 130, // BGP Identifier
					5, // Opt Parm Len
					1, 2, 3, 4, 5,
				},
				ReceivedOpenMsg: []byte{
					// OPEN Recv
					4,    // Version
					1, 0, // ASN
					2, 0, // Hold Time
					100, 110, 120, 130, // BGP Identifier
					0, // Opt Parm Len
				},
			},
		},
		{
			name: "Incomplete #1",
			input: []byte{
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 100,
			},
			ch: &CommonHeader{
				MsgLength: 47,
			},
			wantFail: true,
		},
		{
			name: "Incomplete #2",
			input: []byte{
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 100,
				0, 200,

				// OPEN Sent
				4,    // Version
				1, 0, // ASN
				2, 0, // Hold Time
			},
			ch: &CommonHeader{
				MsgLength: 47,
			},
			wantFail: true,
		},
		{
			name: "Incomplete #3",
			input: []byte{
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 100,
				0, 200,

				// OPEN Sent
				4,    // Version
				1, 0, // ASN
				2, 0, // Hold Time
			},
			ch: &CommonHeader{
				MsgLength: 47,
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
				4,    // Version
				1, 0, // ASN
				2, 0, // Hold Time
				100, 110, 120, 130, // BGP Identifier
				5, // Opt Parm Len
				1, 2, 3, 4, 5,

				// OPEN Recv
				4,    // Version
				1, 0, // ASN
				2, 0, // Hold Time
				100, 110, 120, 130, // BGP Identifier
				3, // Opt Parm Len
			},
			ch: &CommonHeader{
				MsgLength: 47,
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
