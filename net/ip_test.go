package net

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToUint32(t *testing.T) {
	tests := []struct {
		name     string
		val      uint64
		expected uint32
	}{
		{
			name:     "IP: 172.24.5.1",
			val:      2887255297,
			expected: 2887255297,
		},
		{
			name:     "bigger than IPv4 address",
			val:      2887255295 + 17179869184,
			expected: 2887255295,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ip := IP{
				lower: test.val,
			}
			assert.Equal(t, test.expected, ip.ToUint32())
		})
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		name     string
		ip       *IP
		other    *IP
		expected int
	}{
		{
			name: "equal",
			ip: &IP{
				lower:  100,
				higher: 200,
			},
			other: &IP{
				lower:  100,
				higher: 200,
			},
			expected: 0,
		},
		{
			name: "greater higher word",
			ip: &IP{
				lower:  123,
				higher: 200,
			},
			other: &IP{
				lower:  456,
				higher: 100,
			},
			expected: 1,
		},
		{
			name: "lesser higher word",
			ip: &IP{
				lower:  123,
				higher: 100,
			},
			other: &IP{
				lower:  456,
				higher: 200,
			},
			expected: -1,
		},
		{
			name: "equal higher word but lesser lower word",
			ip: &IP{
				lower:  456,
				higher: 100,
			},
			other: &IP{
				lower:  123,
				higher: 100,
			},
			expected: 1,
		},
		{
			name: "equal higher word but lesser lower word",
			ip: &IP{
				lower:  123,
				higher: 100,
			},
			other: &IP{
				lower:  456,
				higher: 100,
			},
			expected: -1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.ip.Compare(test.other))
		})
	}
}
