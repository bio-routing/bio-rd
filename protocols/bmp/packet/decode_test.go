package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantFail bool
		expected Msg
	}{
		{
			name:     "incomplete common header",
			input:    []byte{1, 2},
			wantFail: true,
		},
		{
			name:     "Invalid version",
			input:    []byte{0, 0, 0, 0, 6, 5},
			wantFail: true,
		},
		{
			name: "Route monitoring ok",
			input: []byte{
				3, 0, 0, 0, 6 + PerPeerHeaderLen + 4, 0,

				1,
				2,
				0, 0, 0, 0, 0, 0, 0, 3,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 0, 200, 124,
				0, 0, 0, 123,
				0, 0, 0, 100,
				0, 0, 0, 200,

				1, 2, 3, 4,
			},
			wantFail: false,
			expected: &RouteMonitoringMsg{
				CommonHeader: &CommonHeader{
					Version:   3,
					MsgLength: 6 + PerPeerHeaderLen + 4,
					MsgType:   0,
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
				BGPUpdate: []byte{1, 2, 3, 4},
			},
		},
		{
			name: "Route monitoring nok",
			input: []byte{
				3, 0, 0, 0, 6 + PerPeerHeaderLen + 4, 0,

				1,
				2,
				0, 0, 0, 0, 0, 0, 0, 3,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 0, 200, 124,
				0, 0, 0, 123,
				0, 0, 0, 100,
				0, 0, 0, 200,

				1, 2,
			},
			wantFail: true,
		},
		{
			name: "Statistic report ok",
			input: []byte{
				3, 0, 0, 0, 6 + 9 + PerPeerHeaderLen, 1,

				1,
				2,
				0, 0, 0, 0, 0, 0, 0, 3,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 0, 200, 124,
				0, 0, 0, 123,
				0, 0, 0, 100,
				0, 0, 0, 200,

				0, 0, 0, 1,
				0, 1, 0, 1, 1,
			},
			wantFail: false,
			expected: &StatsReport{
				CommonHeader: &CommonHeader{
					Version:   3,
					MsgLength: 6 + 9 + PerPeerHeaderLen,
					MsgType:   1,
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
				StatsCount: 1,
				Stats: []*InformationTLV{
					{
						InformationType:   1,
						InformationLength: 1,
						Information:       []byte{1},
					},
				},
			},
		},
		{
			name: "Statistic report nok",
			input: []byte{
				3, 0, 0, 0, 6 + 9 + PerPeerHeaderLen, 1,

				1,
				2,
				0, 0, 0, 0, 0, 0, 0, 3,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 0, 200, 124,
			},
			wantFail: true,
		},
		{
			name: "peer down ok",
			input: []byte{
				3, 0, 0, 0, 6 + 9 + PerPeerHeaderLen, 1,

				1,
				2,
				0, 0, 0, 0, 0, 0, 0, 3,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 0, 200, 124,
				0, 0, 0, 123,
				0, 0, 0, 100,
				0, 0, 0, 200,

				0, 0, 0, 1,
				0, 1, 0, 1, 1,
			},
			wantFail: false,
			expected: &StatsReport{
				CommonHeader: &CommonHeader{
					Version:   3,
					MsgLength: 6 + 9 + PerPeerHeaderLen,
					MsgType:   1,
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
				StatsCount: 1,
				Stats: []*InformationTLV{
					{
						InformationType:   1,
						InformationLength: 1,
						Information:       []byte{1},
					},
				},
			},
		},
		{
			name: "peer down nok",
			input: []byte{
				3, 0, 0, 0, 6 + 9 + PerPeerHeaderLen, 1,

				1,
				2,
				0, 0, 0, 0, 0, 0, 0, 3,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 0, 200, 124,
				0, 0, 0, 123,
				0, 0, 0, 100,
				0, 0, 0, 200,

				0, 0, 0, 1,
				0, 1, 0, 1,
			},
			wantFail: true,
		},
		{
			name: "peer up ok",
			input: []byte{
				3, 0, 0, 0, 54, 3,

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
			wantFail: false,
			expected: &PeerUpNotification{
				CommonHeader: &CommonHeader{
					Version:   3,
					MsgLength: 54,
					MsgType:   3,
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
			name: "peer up nok",
			input: []byte{
				3, 0, 0, 0, 54, 3,

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
				4,    // Version
				1, 0, // ASN
				2, 0, // Hold Time
				100, 110, 120, 130, // BGP Identifier
				5, // Opt Parm Len
				1, 2, 3, 4, 5,

				// OPEN Recv
				4, // Version
				1,
			},
			wantFail: true,
		},
		{
			name: "initiation message ok",
			input: []byte{
				3, 0, 0, 0, 11, 4,

				0, 1, 0, 1, 5,
			},
			wantFail: false,
			expected: &InitiationMessage{
				CommonHeader: &CommonHeader{
					Version:   3,
					MsgLength: 11,
					MsgType:   4,
				},
				TLVs: []*InformationTLV{
					{
						InformationType:   1,
						InformationLength: 1,
						Information:       []byte{5},
					},
				},
			},
		},
		{
			name: "initiation message nok",
			input: []byte{
				3, 0, 0, 0,
			},
			wantFail: true,
		},
		{
			name: "termination message ok",
			input: []byte{
				3, 0, 0, 0, 11, 5,

				0, 1, 0, 1, 5,
			},
			wantFail: false,
			expected: &TerminationMessage{
				CommonHeader: &CommonHeader{
					Version:   3,
					MsgLength: 11,
					MsgType:   5,
				},
				TLVs: []*InformationTLV{
					{
						InformationType:   1,
						InformationLength: 1,
						Information:       []byte{5},
					},
				},
			},
		},
		{
			name: "termination message nok",
			input: []byte{
				3, 0, 0, 0, 11, 5,

				0, 1, 0,
			},
			wantFail: true,
		},
		{
			name: "route mirror message ok",
			input: []byte{
				3, 0, 0, 0, 49, 6,

				1,
				2,
				0, 0, 0, 0, 0, 0, 0, 3,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 0, 200, 124,
				0, 0, 0, 123,
				0, 0, 0, 100,
				0, 0, 0, 200,

				0, 1, 0, 1, 5,
			},
			wantFail: false,
			expected: &RouteMirroringMsg{
				CommonHeader: &CommonHeader{
					Version:   3,
					MsgLength: 49,
					MsgType:   6,
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
				TLVs: []*InformationTLV{
					{
						InformationType:   1,
						InformationLength: 1,
						Information:       []byte{5},
					},
				},
			},
		},
		{
			name: "route mirror message nok",
			input: []byte{
				3, 0, 0, 0, 49, 6,

				1,
				2,
				0, 0, 0, 0, 0, 0, 0, 3,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 0, 200, 124,
				0, 0,
			},
			wantFail: true,
		},
		{
			name: "invalid msg type",
			input: []byte{
				3, 0, 0, 0, 49, 7,
			},
			wantFail: true,
		},
	}

	for _, test := range tests {
		m, err := Decode(test.input)
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

		assert.Equalf(t, test.expected, m, "Test %q", test.name)
	}
}
