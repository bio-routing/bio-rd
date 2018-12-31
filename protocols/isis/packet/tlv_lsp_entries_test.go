package packet

import (
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
