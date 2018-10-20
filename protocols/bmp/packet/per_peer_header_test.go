package packet

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPerPeerHeaderSerialize(t *testing.T) {
	tests := []struct {
		name     string
		input    *PerPeerHeader
		expected []byte
	}{
		{
			name: "Test #1",
			input: &PerPeerHeader{
				PeerType:              1,
				PeerFlags:             2,
				PeerDistinguisher:     3,
				PeerAddress:           [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
				PeerAS:                51324,
				PeerBGPID:             123,
				Timestamp:             100,
				TimestampMicroSeconds: 200,
			},
			expected: []byte{
				1,
				2,
				0, 0, 0, 0, 0, 0, 0, 3,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 0, 200, 124,
				0, 0, 0, 123,
				0, 0, 0, 100,
				0, 0, 0, 200,
			},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		test.input.Serialize(buf)
		res := buf.Bytes()

		assert.Equalf(t, test.expected, res, "Test %q", test.name)
	}
}

func TestDecodePerPeerHeader(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantFail bool
		expected *PerPeerHeader
	}{
		{
			name: "Full packet",
			input: []byte{
				1,
				2,
				0, 0, 0, 0, 0, 0, 0, 3,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				0, 0, 200, 124,
				0, 0, 0, 123,
				0, 0, 0, 100,
				0, 0, 0, 200,
			},
			wantFail: false,
			expected: &PerPeerHeader{
				PeerType:              1,
				PeerFlags:             2,
				PeerDistinguisher:     3,
				PeerAddress:           [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
				PeerAS:                51324,
				PeerBGPID:             123,
				Timestamp:             100,
				TimestampMicroSeconds: 200,
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
				0, 0, 0, 123,
				0, 0, 0, 100,
				0, 0, 0,
			},
			wantFail: true,
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		p, err := decodePerPeerHeader(buf)
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
