package net

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPfx(t *testing.T) {
	p := NewPfx(123, 11)
	if p.addr != 123 || p.pfxlen != 11 {
		t.Errorf("NewPfx() failed: Unexpected values")
	}
}

func TestAddr(t *testing.T) {
	tests := []struct {
		name     string
		pfx      Prefix
		expected uint32
	}{
		{
			name:     "Test 1",
			pfx:      NewPfx(100, 5),
			expected: 100,
		},
	}

	for _, test := range tests {
		res := test.pfx.Addr()
		if res != test.expected {
			t.Errorf("Unexpected result for test %s: Got %d Expected %d", test.name, res, test.expected)
		}
	}
}

func TestPfxlen(t *testing.T) {
	tests := []struct {
		name     string
		pfx      Prefix
		expected uint8
	}{
		{
			name:     "Test 1",
			pfx:      NewPfx(100, 5),
			expected: 5,
		},
	}

	for _, test := range tests {
		res := test.pfx.Pfxlen()
		if res != test.expected {
			t.Errorf("Unexpected result for test %s: Got %d Expected %d", test.name, res, test.expected)
		}
	}
}

func TestGetSupernet(t *testing.T) {
	tests := []struct {
		name     string
		a        Prefix
		b        Prefix
		expected Prefix
	}{
		{
			name: "Test 1",
			a: Prefix{
				addr:   167772160, // 10.0.0.0/8
				pfxlen: 8,
			},
			b: Prefix{
				addr:   191134464, // 11.100.123.0/24
				pfxlen: 24,
			},
			expected: Prefix{
				addr:   167772160, // 10.0.0.0/7
				pfxlen: 7,
			},
		},
		{
			name: "Test 2",
			a: Prefix{
				addr:   167772160, // 10.0.0.0/8
				pfxlen: 8,
			},
			b: Prefix{
				addr:   3232235520, // 192.168.0.0/24
				pfxlen: 24,
			},
			expected: Prefix{
				addr:   0, // 0.0.0.0/0
				pfxlen: 0,
			},
		},
	}

	for _, test := range tests {
		s := test.a.GetSupernet(test.b)
		assert.Equal(t, s, test.expected)
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		a        Prefix
		b        Prefix
		expected bool
	}{
		{
			name: "Test 1",
			a: Prefix{
				addr:   0,
				pfxlen: 0,
			},
			b: Prefix{
				addr:   100,
				pfxlen: 24,
			},
			expected: true,
		},
		{
			name: "Test 2",
			a: Prefix{
				addr:   100,
				pfxlen: 24,
			},
			b: Prefix{
				addr:   0,
				pfxlen: 0,
			},
			expected: false,
		},
		{
			name: "Test 3",
			a: Prefix{
				addr:   167772160,
				pfxlen: 8,
			},
			b: Prefix{
				addr:   167772160,
				pfxlen: 9,
			},
			expected: true,
		},
		{
			name: "Test 4",
			a: Prefix{
				addr:   167772160,
				pfxlen: 8,
			},
			b: Prefix{
				addr:   174391040,
				pfxlen: 24,
			},
			expected: true,
		},
		{
			name: "Test 5",
			a: Prefix{
				addr:   167772160,
				pfxlen: 8,
			},
			b: Prefix{
				addr:   184549377,
				pfxlen: 24,
			},
			expected: false,
		},
		{
			name: "Test 6",
			a: Prefix{
				addr:   167772160,
				pfxlen: 8,
			},
			b: Prefix{
				addr:   191134464,
				pfxlen: 24,
			},
			expected: false,
		},
		{
			name: "Test 7",
			a: Prefix{
				addr:   strAddr("169.0.0.0"),
				pfxlen: 25,
			},
			b: Prefix{
				addr:   strAddr("169.1.1.0"),
				pfxlen: 26,
			},
			expected: false,
		},
	}

	for _, test := range tests {
		res := test.a.Contains(test.b)
		if res != test.expected {
			t.Errorf("Unexpected result %v for test %s: %s contains %s\n", res, test.name, test.a.String(), test.b.String())
		}
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		name     string
		a        uint8
		b        uint8
		expected uint8
	}{
		{
			name:     "Min 100 200",
			a:        100,
			b:        200,
			expected: 100,
		},
		{
			name:     "Min 200 100",
			a:        200,
			b:        100,
			expected: 100,
		},
		{
			name:     "Min 111 112",
			a:        111,
			b:        112,
			expected: 111,
		},
	}

	for _, test := range tests {
		res := min(test.a, test.b)
		if res != test.expected {
			t.Errorf("Unexpected result for test %s: Got %d Expected %d", test.name, res, test.expected)
		}
	}
}

func TestEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        Prefix
		b        Prefix
		expected bool
	}{
		{
			name:     "Equal PFXs",
			a:        NewPfx(100, 8),
			b:        NewPfx(100, 8),
			expected: true,
		},
		{
			name:     "Unequal PFXs",
			a:        NewPfx(100, 8),
			b:        NewPfx(200, 8),
			expected: false,
		},
	}

	for _, test := range tests {
		res := test.a.Equal(test.b)
		if res != test.expected {
			t.Errorf("Unexpected result for %q: Got %v Expected %v", test.name, res, test.expected)
		}
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name     string
		pfx      Prefix
		expected string
	}{
		{
			name:     "Test 1",
			pfx:      NewPfx(167772160, 8), // 10.0.0.0/8
			expected: "10.0.0.0/8",
		},
		{
			name:     "Test 2",
			pfx:      NewPfx(167772160, 16), // 10.0.0.0/8
			expected: "10.0.0.0/16",
		},
	}

	for _, test := range tests {
		res := test.pfx.String()
		if res != test.expected {
			t.Errorf("Unexpected result for %q: Got %q Expected %q", test.name, res, test.expected)
		}
	}
}

func TestStrToAddr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantFail bool
		expected uint32
	}{
		{
			name:     "Invalid address #1",
			input:    "10.10.10",
			wantFail: true,
		},
		{
			name:     "Invalid address #2",
			input:    "",
			wantFail: true,
		},
		{
			name:     "Invalid address #3",
			input:    "10.10.10.10.10",
			wantFail: true,
		},
		{
			name:     "Invalid address #4",
			input:    "10.256.0.0",
			wantFail: true,
		},
		{
			name:     "Valid address",
			input:    "10.0.0.0",
			wantFail: false,
			expected: 167772160,
		},
	}

	for _, test := range tests {
		res, err := StrToAddr(test.input)
		if err != nil {
			if test.wantFail {
				continue
			}
			t.Errorf("Unexpected failure for test %q: %v", test.name, err)
			continue
		}

		if test.wantFail {
			t.Errorf("Unexpected success for test %q", test.name)
			continue
		}

		assert.Equal(t, test.expected, res)
	}
}

func strAddr(s string) uint32 {
	ret, _ := StrToAddr(s)
	return ret
}
