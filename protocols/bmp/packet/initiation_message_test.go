package packet

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitiationMessageMsgType(t *testing.T) {
	pd := &InitiationMessage{
		CommonHeader: &CommonHeader{
			MsgType: 100,
		},
	}

	if pd.MsgType() != 100 {
		t.Errorf("Unexpected result")
	}
}
func TestDecodeInitiationMessage(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		ch       *CommonHeader
		wantFail bool
		expected *InitiationMessage
	}{
		{
			name: "Full",
			input: []byte{
				0, 1, // sysDescr
				0, 4, // Length
				42, 42, 42, 42, // AAAA
				0, 2, //sysName
				0, 5, // Length
				43, 43, 43, 43, 43, // BBBBB
			},
			ch: &CommonHeader{
				MsgLength: CommonHeaderLen + 17,
			},
			wantFail: false,
			expected: &InitiationMessage{
				TLVs: []*InformationTLV{
					{
						InformationType:   1,
						InformationLength: 4,
						Information:       []byte{42, 42, 42, 42},
					},
					{
						InformationType:   2,
						InformationLength: 5,
						Information:       []byte{43, 43, 43, 43, 43},
					},
				},
			},
		},
		{
			name: "Incomplete",
			input: []byte{
				0, 1, // sysDescr
				0, 4, // Length
				42, 42, 42, 42, // AAAA
				0, 2, //sysName
				0, 5, // Length
				43, 43, 43, 43, // BBBB
			},
			ch: &CommonHeader{
				MsgLength: CommonHeaderLen + 17,
			},
			wantFail: true,
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		im, err := decodeInitiationMessage(buf, test.ch)
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

		assert.Equalf(t, test.expected, im, "Test %q", test.name)
	}
}
