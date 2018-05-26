package net

import (
	"testing"

	"net"

	"github.com/stretchr/testify/assert"
)

func TestIPv4ToUint32(t *testing.T) {
	tests := []struct {
		input    []byte
		expected uint32
	}{
		{
			input:    []byte{192, 168, 1, 5},
			expected: 3232235781,
		},
		{
			input:    []byte{10, 0, 0, 0},
			expected: 167772160,
		},
		{
			input:    []byte{172, 24, 5, 1},
			expected: 2887255297,
		},
		{
			input:    net.ParseIP("172.24.5.1"),
			expected: 2887255297,
		},
	}

	for _, test := range tests {
		res := IPv4ToUint32(test.input)
		assert.Equal(t, test.expected, res)
	}
}
