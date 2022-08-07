package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClusterListString(t *testing.T) {
	tests := []struct {
		name     string
		value    *ClusterList
		expected string
	}{
		{
			name:     "nil",
			expected: "",
			value:    nil,
		},
		{
			name:     "one element",
			value:    &ClusterList{23},
			expected: "23",
		},
		{
			name:     "two values",
			value:    &ClusterList{23, 42},
			expected: "23 42",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(te *testing.T) {
			assert.Equal(te, test.expected, test.value.String())
		})
	}
}
