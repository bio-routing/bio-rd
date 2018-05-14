package filter

import (
	"testing"

	"github.com/bio-routing/bio-rd/net"
	"github.com/stretchr/testify/assert"
)

func TestInRange(t *testing.T) {
	tests := []struct {
		name     string
		prefix   net.Prefix
		pattern  net.Prefix
		begin    uint8
		end      uint8
		expected bool
	}{
		{
			name:     "matches and in range (22-24)",
			prefix:   net.NewPfx(strAddr("1.2.1.0"), 23),
			pattern:  net.NewPfx(strAddr("1.2.0.0"), 22),
			begin:    22,
			end:      24,
			expected: true,
		},
		{
			name:     "matches begin of range (22-24)",
			prefix:   net.NewPfx(strAddr("1.2.0.0"), 22),
			pattern:  net.NewPfx(strAddr("1.2.0.0"), 22),
			begin:    22,
			end:      24,
			expected: true,
		},
		{
			name:     "matches end of range (22-24)",
			prefix:   net.NewPfx(strAddr("1.2.128.0"), 24),
			pattern:  net.NewPfx(strAddr("1.2.0.0"), 22),
			begin:    22,
			end:      24,
			expected: true,
		},
		{
			name:     "matches begin and end of range (24-24)",
			prefix:   net.NewPfx(strAddr("1.2.0.0"), 24),
			pattern:  net.NewPfx(strAddr("1.2.0.0"), 24),
			begin:    24,
			end:      24,
			expected: true,
		},
		{
			name:     "smaller (22-24)",
			prefix:   net.NewPfx(strAddr("1.2.0.0"), 16),
			pattern:  net.NewPfx(strAddr("1.2.4.0"), 22),
			begin:    22,
			end:      24,
			expected: false,
		},
		{
			name:     "longer (22-24)",
			prefix:   net.NewPfx(strAddr("1.2.0.128"), 25),
			pattern:  net.NewPfx(strAddr("1.2.0.0"), 22),
			begin:    22,
			end:      24,
			expected: false,
		},
		{
			name:     "does not match",
			prefix:   net.NewPfx(strAddr("2.0.0.0"), 23),
			pattern:  net.NewPfx(strAddr("1.2.0.0"), 22),
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(te *testing.T) {
			f := NewRouteFilter(test.pattern, InRange(test.begin, test.end))
			assert.Equal(te, test.expected, f.Matches(test.prefix))
		})
	}
}

func TestExact(t *testing.T) {
	tests := []struct {
		name     string
		prefix   net.Prefix
		pattern  net.Prefix
		expected bool
	}{
		{
			name:     "matches (0.0.0.0/0)",
			prefix:   net.NewPfx(strAddr("0.0.0.0"), 0),
			pattern:  net.NewPfx(strAddr("0.0.0.0"), 0),
			expected: true,
		},
		{
			name:     "matches (192.168.0.0)",
			prefix:   net.NewPfx(strAddr("192.168.1.1"), 24),
			pattern:  net.NewPfx(strAddr("192.168.1.1"), 24),
			expected: true,
		},
		{
			name:     "does not match",
			prefix:   net.NewPfx(strAddr("1.0.0.0"), 8),
			pattern:  net.NewPfx(strAddr("0.0.0.0"), 0),
			expected: false,
		},
		{
			name:     "longer",
			prefix:   net.NewPfx(strAddr("1.0.0.0"), 8),
			pattern:  net.NewPfx(strAddr("1.0.0.0"), 7),
			expected: false,
		},
		{
			name:     "lesser",
			prefix:   net.NewPfx(strAddr("1.0.0.0"), 7),
			pattern:  net.NewPfx(strAddr("1.0.0.0"), 8),
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(te *testing.T) {
			f := NewRouteFilter(test.pattern, Exact())
			assert.Equal(te, test.expected, f.Matches(test.prefix))
		})
	}
}

func TestOrLonger(t *testing.T) {
	tests := []struct {
		name     string
		prefix   net.Prefix
		pattern  net.Prefix
		expected bool
	}{
		{
			name:     "longer",
			prefix:   net.NewPfx(strAddr("1.2.3.128"), 25),
			pattern:  net.NewPfx(strAddr("1.2.3.0"), 24),
			expected: true,
		},
		{
			name:     "exact",
			prefix:   net.NewPfx(strAddr("1.2.3.0"), 24),
			pattern:  net.NewPfx(strAddr("1.2.3.0"), 24),
			expected: true,
		},
		{
			name:     "lesser",
			prefix:   net.NewPfx(strAddr("1.2.3.0"), 23),
			pattern:  net.NewPfx(strAddr("1.2.3.0"), 24),
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(te *testing.T) {
			f := NewRouteFilter(test.pattern, OrLonger())
			assert.Equal(te, test.expected, f.Matches(test.prefix))
		})
	}
}

func TestLonger(t *testing.T) {
	tests := []struct {
		name     string
		prefix   net.Prefix
		pattern  net.Prefix
		expected bool
	}{
		{
			name:     "longer",
			prefix:   net.NewPfx(strAddr("1.2.3.128"), 25),
			pattern:  net.NewPfx(strAddr("1.2.3.0"), 24),
			expected: true,
		},
		{
			name:     "exact",
			prefix:   net.NewPfx(strAddr("1.2.3.0"), 24),
			pattern:  net.NewPfx(strAddr("1.2.3.0"), 24),
			expected: false,
		},
		{
			name:     "lesser",
			prefix:   net.NewPfx(strAddr("1.2.3.0"), 23),
			pattern:  net.NewPfx(strAddr("1.2.3.0"), 24),
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(te *testing.T) {
			f := NewRouteFilter(test.pattern, Longer())
			assert.Equal(te, test.expected, f.Matches(test.prefix))
		})
	}
}
