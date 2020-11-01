package server

import (
	"testing"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/stretchr/testify/assert"
)

func TestNETsCompatible(t *testing.T) {
	tests := []struct {
		name     string
		input    []*types.NET
		expected bool
	}{
		{
			name: "Test #1",
			input: []*types.NET{
				{
					SystemID: types.SystemID{1, 1, 1, 1, 1, 1},
				},
				{
					SystemID: types.SystemID{1, 1, 1, 1, 1, 1},
				},
			},
			expected: true,
		},
		{
			name: "Test #2",
			input: []*types.NET{
				{
					SystemID: types.SystemID{1, 1, 1, 1, 1, 1},
				},
				{
					SystemID: types.SystemID{1, 1, 1, 1, 1, 2},
				},
			},
			expected: false,
		},
	}

	for _, test := range tests {
		assert.Equalf(t, test.expected, netsCompatible(test.input), test.name)
	}
}
