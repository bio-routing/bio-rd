package packet

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouteMirrorType(t *testing.T) {
	pd := &RouteMirroringMsg{
		CommonHeader: &CommonHeader{
			MsgType: 100,
		},
	}

	if pd.MsgType() != 100 {
		t.Errorf("Unexpected result")
	}
}
func TestDecodeRouteMirroringMsg(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		ch       *CommonHeader
		wantFail bool
		expected *RouteMirroringMsg
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

				0, 1, 0, 2, 100, 200,
				0, 1, 0, 2, 100, 200,
			},
			ch: &CommonHeader{
				MsgLength: CommonHeaderLen + PerPeerHeaderLen + 12,
			},
			wantFail: false,
			expected: &RouteMirroringMsg{
				CommonHeader: &CommonHeader{
					MsgLength: CommonHeaderLen + PerPeerHeaderLen + 12,
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
						InformationLength: 2,
						Information:       []byte{100, 200},
					},
					{
						InformationType:   1,
						InformationLength: 2,
						Information:       []byte{100, 200},
					},
				},
			},
		},
		{
			name: "Incomplete",
			input: []byte{
				1,
				2,
				0, 0, 0, 0, 0, 0, 0, 3,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 0, 200, 124,
			},
			ch: &CommonHeader{
				MsgLength: CommonHeaderLen + PerPeerHeaderLen + 12,
			},
			wantFail: true,
		},
		{
			name: "Incomplete TLV",
			input: []byte{
				1,
				2,
				0, 0, 0, 0, 0, 0, 0, 3,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 0, 200, 124,
				0, 0, 0, 123,
				0, 0, 0, 100,
				0, 0, 0, 200,

				0, 1, 0, 2, 100, 200,
				0, 1, 0, 2,
			},
			ch: &CommonHeader{
				MsgLength: CommonHeaderLen + PerPeerHeaderLen + 12,
			},
			wantFail: true,
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		r, err := decodeRouteMirroringMsg(buf, test.ch)
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
