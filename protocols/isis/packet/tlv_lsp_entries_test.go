package packet

import (
	"bytes"
	"testing"

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
		{},
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

		assert.Equalf(t, test.expected, tlv, "Test %q", err)
	}
}
