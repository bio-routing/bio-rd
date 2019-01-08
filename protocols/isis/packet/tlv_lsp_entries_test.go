package packet

import (
	"bytes"
	"testing"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/stretchr/testify/assert"
)

func TestNewLSPEntriesTLV(t *testing.T) {
	tests := []struct {
		name       string
		lspEntries []*LSPEntry
		expected   *LSPEntriesTLV
	}{
		{
			name: "Test #1",
			lspEntries: []*LSPEntry{
				{},
			},
			expected: &LSPEntriesTLV{
				TLVType:   9,
				TLVLength: 16,
				LSPEntries: []*LSPEntry{
					{},
				},
			},
		},
	}

	for _, test := range tests {
		res := NewLSPEntriesTLV(test.lspEntries)
		assert.Equalf(t, test.expected, res, "Test %q", test.name)
	}
}

func TestReadLSPEntriesTLV(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected *LSPEntriesTLV
		wantFail bool
	}{
		{
			name: "Test #1",
			input: []byte{
				2, 0, // Remaining Lifetime
				10, 20, 30, 40, 50, 60, 88, 99, // LSPID
				0, 0, 0, 22, // Sequence Number
				2, 0, // LSPChecksum
			},
			expected: &LSPEntriesTLV{
				TLVType:   9,
				TLVLength: 16,
				LSPEntries: []*LSPEntry{
					{
						RemainingLifetime: 512,
						LSPID: LSPID{
							SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
							PseudonodeID: 88,
							LSPNumber:    99,
						},
						SequenceNumber: 22,
						LSPChecksum:    512,
					},
				},
			},
			wantFail: false,
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		tlv, err := readLSPEntriesTLV(buf, 9, uint8(len(test.input)))
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

		assert.Equalf(t, test.expected, tlv, "Test %q", test.name)
	}
}
