package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMACString(t *testing.T) {
	tests := []struct {
		name     string
		addr     MACAddress
		expected string
	}{
		{
			name:     "Test #1",
			addr:     MACAddress{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			expected: "ff:ff:ff:ff:ff:ff",
		},
		{
			name:     "Test #1",
			addr:     MACAddress{1, 1, 1, 1, 1, 1},
			expected: "01:01:01:01:01:01",
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.addr.String(), test.name)
	}
}
