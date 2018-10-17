package packet

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatsReportMsgType(t *testing.T) {
	pd := &StatsReport{
		CommonHeader: &CommonHeader{
			MsgType: 100,
		},
	}

	if pd.MsgType() != 100 {
		t.Errorf("Unexpected result")
	}
}

func TestDecodeStatsReport(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantFail bool
		expected *StatsReport
	}{
		{
			name: "Full",
			input: []byte{
				// Per Peer Header
				1,
				2,
				0, 0, 0, 3,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 0, 200, 124,
				0, 0, 0, 123,
				0, 0, 0, 100,
				0, 0, 0, 200,

				// Stats Count
				0, 0, 0, 2,

				0, 1,
				0, 4,
				0, 0, 0, 2,

				0, 2,
				0, 4,
				0, 0, 0, 3,
			},
			wantFail: false,
			expected: &StatsReport{
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
				StatsCount: 2,
				Stats: []*InformationTLV{
					{
						InformationType:   1,
						InformationLength: 4,
						Information:       []byte{0, 0, 0, 2},
					},
					{
						InformationType:   2,
						InformationLength: 4,
						Information:       []byte{0, 0, 0, 3},
					},
				},
			},
		},
		{
			name: "Incomplete per peer header",
			input: []byte{
				// Per Peer Header
				1,
				2,
				0, 0, 0, 3,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 0, 200, 124,
				0, 0, 0, 123,
			},
			wantFail: true,
		},
		{
			name: "Incomplete stats count",
			input: []byte{
				// Per Peer Header
				1,
				2,
				0, 0, 0, 3,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 0, 200, 124,
				0, 0, 0, 123,
				0, 0, 0, 100,
				0, 0, 0, 200,
			},
			wantFail: true,
		},
		{
			name: "Incomplete TLV",
			input: []byte{
				// Per Peer Header
				1,
				2,
				0, 0, 0, 3,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 0, 200, 124,
				0, 0, 0, 123,
				0, 0, 0, 100,
				0, 0, 0, 200,

				// Stats Count
				0, 0, 0, 2,

				0, 1,
				0, 4,
			},
			wantFail: true,
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		sr, err := decodeStatsReport(buf, nil)
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

		assert.Equalf(t, test.expected, sr, "Test %q", test.name)
	}
}
