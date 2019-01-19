package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIPAddrsEqual(t *testing.T) {
	tests := []struct {
		name     string
		n        *neighbor
		x        []uint32
		expected bool
	}{
		{
			name: "Test #1",
			n: &neighbor{
				ipInterfaceAddresses: []uint32{
					100,
					200,
					300,
				},
			},
			x: []uint32{
				200,
				300,
				100,
			},
			expected: true,
		},
		{
			name: "Test #2",
			n: &neighbor{
				ipInterfaceAddresses: []uint32{
					100,
					200,
					300,
				},
			},
			x: []uint32{
				300,
				100,
			},
			expected: false,
		},
	}

	for _, test := range tests {
		res := test.n.ipAddrsEqual(test.x)
		assert.Equal(t, test.expected, res, test.name)
	}
}
