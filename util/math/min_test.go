package math

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMin(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{
			name:     "A",
			a:        100,
			b:        200,
			expected: 100,
		},
		{
			name:     "A",
			a:        200,
			b:        100,
			expected: 100,
		},
	}

	for _, test := range tests {
		res := Min(test.a, test.b)
		assert.Equalf(t, test.expected, res, "Test %q", test.name)
	}
}
