package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWireLength(t *testing.T) {
	tests := []struct {
		name     string
		pa       *UnknownPathAttribute
		expected uint16
	}{
		{
			name: "Test #1",
			pa: &UnknownPathAttribute{
				Value: []byte{1, 2, 3},
			},
			expected: 6,
		},
		{
			name: "Non extended length corner case",
			pa: &UnknownPathAttribute{
				Value: []byte{
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255,
				},
			},
			expected: 258,
		},
		{
			name: "Extended length corner case",
			pa: &UnknownPathAttribute{
				Value: []byte{
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255,
				},
			},
			expected: 260,
		},
	}

	for _, test := range tests {
		res := test.pa.WireLength()
		if res != test.expected {
			t.Errorf("Unexpected result for test %q: Expected: %d Got: %d", test.name, test.expected, res)
		}
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		name  string
		u     *UnknownPathAttribute
		c     *UnknownPathAttribute
		equal bool
	}{
		{
			name: "Optional different",
			u: &UnknownPathAttribute{
				Optional: true,
				Value:    []byte{1, 2, 3},
			},
			c: &UnknownPathAttribute{
				Optional: false,
				Value:    []byte{1, 2, 3},
			},
			equal: false,
		},
		{
			name: "Transitive different",
			u: &UnknownPathAttribute{
				Optional:   true,
				Transitive: true,
				Value:      []byte{1, 2, 3},
			},
			c: &UnknownPathAttribute{
				Optional: true,
				Value:    []byte{1, 2, 3},
			},
			equal: false,
		},
		{
			name: "Partial different",
			u: &UnknownPathAttribute{
				Optional:   true,
				Transitive: true,
				Partial:    true,
				Value:      []byte{1, 2, 3},
			},
			c: &UnknownPathAttribute{
				Optional:   true,
				Transitive: true,
				Value:      []byte{1, 2, 3},
			},
			equal: false,
		},
		{
			name: "TypeCode different",
			u: &UnknownPathAttribute{
				Optional:   true,
				Transitive: true,
				Partial:    true,
				TypeCode:   23,
				Value:      []byte{1, 2, 3},
			},
			c: &UnknownPathAttribute{
				Optional:   true,
				Transitive: true,
				Partial:    true,
				TypeCode:   42,
				Value:      []byte{1, 2, 3},
			},
			equal: false,
		},
		{
			name: "Value lenght different",
			u: &UnknownPathAttribute{
				Optional:   true,
				Transitive: true,
				Partial:    true,
				TypeCode:   42,
				Value:      []byte{1, 2, 3},
			},
			c: &UnknownPathAttribute{
				Optional:   true,
				Transitive: true,
				Partial:    true,
				TypeCode:   42,
				Value:      []byte{1, 2},
			},
			equal: false,
		},
		{
			name: "Value different",
			u: &UnknownPathAttribute{
				Optional:   true,
				Transitive: true,
				Partial:    true,
				TypeCode:   42,
				Value:      []byte{1, 2, 3},
			},
			c: &UnknownPathAttribute{
				Optional:   true,
				Transitive: true,
				Partial:    true,
				TypeCode:   42,
				Value:      []byte{1, 2, 4},
			},
			equal: false,
		},
		{
			name: "All equal",
			u: &UnknownPathAttribute{
				Optional:   true,
				Transitive: true,
				Partial:    true,
				TypeCode:   42,
				Value:      []byte{0, 8, 15},
			},
			c: &UnknownPathAttribute{
				Optional:   true,
				Transitive: true,
				Partial:    true,
				TypeCode:   42,
				Value:      []byte{0, 8, 15},
			},
			equal: true,
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.u.Compare(test.c), test.equal, test.name)
	}
}
