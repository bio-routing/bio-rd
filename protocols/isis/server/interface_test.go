package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGet3ByteMetric(t *testing.T) {
	tests := []struct {
		name     string
		level    *level
		expected [3]byte
	}{
		{
			name: "Test #1",
			level: &level{
				Metric: 512,
			},
			expected: [3]byte{0, 2, 0},
		},
		{
			name: "Test #2",
			level: &level{
				Metric: 513,
			},
			expected: [3]byte{0, 2, 1},
		},
	}

	for _, test := range tests {
		res := test.level.get3ByteMetric()
		assert.Equal(t, test.expected, res, test.name)
	}
}
