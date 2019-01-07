package net

import (
	gonet "net"
	"testing"

	"github.com/bio-routing/bio-rd/net/api"
	"github.com/stretchr/testify/assert"
)

func TestGetIPNet(t *testing.T) {
	tests := []struct {
		name     string
		pfx      Prefix
		expected *gonet.IPNet
	}{
		{
			name: "Some prefix IPv4",
			pfx:  NewPfx(IPv4FromOctets(127, 0, 0, 0), 8),
			expected: &gonet.IPNet{
				IP:   gonet.IP{127, 0, 0, 0},
				Mask: gonet.IPMask{255, 0, 0, 0},
			},
		},
		{
			name: "Some prefix IPv6",
			pfx:  NewPfx(IPv6FromBlocks(0xffff, 0xffff, 0xffff, 0xffff, 0x0, 0x0, 0x0, 0x0), 64),
			expected: &gonet.IPNet{
				IP:   gonet.IP{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				Mask: gonet.IPMask{255, 255, 255, 255, 255, 255, 255, 255, 0, 0, 0, 0, 0, 0, 0, 0},
			},
		},
	}
	for _, test := range tests {
		res := test.pfx.GetIPNet()
		assert.Equal(t, test.expected, res, test.name)
	}
}
func TestNewPfxFromIPNet(t *testing.T) {
	tests := []struct {
		name     string
		ipNet    *gonet.IPNet
		expected Prefix
	}{
		{
			name: "Some Prefix",
			ipNet: &gonet.IPNet{
				IP:   gonet.IP{127, 0, 0, 0},
				Mask: gonet.IPMask{255, 0, 0, 0},
			},
			expected: NewPfx(IPv4FromOctets(127, 0, 0, 0), 8),
		},
	}
	for _, test := range tests {
		res := NewPfxFromIPNet(test.ipNet)
		assert.Equal(t, test.expected, res, test.name)
	}
}

func TestPrefixToProto(t *testing.T) {
	tests := []struct {
		name     string
		pfx      Prefix
		expected api.Prefix
	}{
		{
			name: "IPv4",
			pfx: Prefix{
				addr: IP{
					lower:    200,
					isLegacy: true,
				},
				pfxlen: 24,
			},
			expected: api.Prefix{
				Address: &api.IP{
					Lower:   200,
					Version: api.IP_IPv4,
				},
				Pfxlen: 24,
			},
		},
		{
			name: "IPv6",
			pfx: Prefix{
				addr: IP{
					higher:   100,
					lower:    200,
					isLegacy: false,
				},
				pfxlen: 64,
			},
			expected: api.Prefix{
				Address: &api.IP{
					Higher:  100,
					Lower:   200,
					Version: api.IP_IPv6,
				},
				Pfxlen: 64,
			},
		},
	}

	for _, test := range tests {
		res := test.pfx.ToProto()
		assert.Equal(t, test.expected, res, test.name)
	}
}

func TestNewPrefixFromProtoPrefix(t *testing.T) {
	tests := []struct {
		name     string
		proto    api.Prefix
		expected Prefix
	}{
		{
			name: "IPv4",
			proto: api.Prefix{
				Address: &api.IP{
					Higher:  0,
					Lower:   2000,
					Version: api.IP_IPv4,
				},
				Pfxlen: 24,
			},
			expected: Prefix{
				addr: IP{
					higher:   0,
					lower:    2000,
					isLegacy: true,
				},
				pfxlen: 24,
			},
		},
		{
			name: "IPv6",
			proto: api.Prefix{
				Address: &api.IP{
					Higher:  1000,
					Lower:   2000,
					Version: api.IP_IPv6,
				},
				Pfxlen: 64,
			},
			expected: Prefix{
				addr: IP{
					higher:   1000,
					lower:    2000,
					isLegacy: false,
				},
				pfxlen: 64,
			},
		},
	}

	for _, test := range tests {
		res := NewPrefixFromProtoPrefix(test.proto)
		assert.Equal(t, test.expected, res, test.name)
	}
}

func TestNewPfx(t *testing.T) {
	p := NewPfx(IPv4(123), 11)
	if p.addr != IPv4(123) || p.pfxlen != 11 {
		t.Errorf("NewPfx() failed: Unexpected values")
	}
}

func TestAddr(t *testing.T) {
	tests := []struct {
		name     string
		pfx      Prefix
		expected IP
	}{
		{
			name:     "Test 1",
			pfx:      NewPfx(IPv4(100), 5),
			expected: IPv4(100),
		},
	}

	for _, test := range tests {
		res := test.pfx.Addr()
		if res != test.expected {
			t.Errorf("Unexpected result for test %s: Got %v Expected %v", test.name, res, test.expected)
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
			pfx:      NewPfx(IPv4(100), 5),
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
			name: "Supernet of 10.0.0.0 and 11.100.123.0 -> 10.0.0.0/7",
			a: Prefix{
				addr:   IPv4FromOctets(10, 0, 0, 0),
				pfxlen: 8,
			},
			b: Prefix{
				addr:   IPv4FromOctets(11, 100, 123, 0),
				pfxlen: 24,
			},
			expected: Prefix{
				addr:   IPv4FromOctets(10, 0, 0, 0),
				pfxlen: 7,
			},
		},
		{
			name: "Supernet of 10.0.0.0 and 192.168.0.0 -> 0.0.0.0/0",
			a: Prefix{
				addr:   IPv4FromOctets(10, 0, 0, 0),
				pfxlen: 8,
			},
			b: Prefix{
				addr:   IPv4FromOctets(192, 168, 0, 0),
				pfxlen: 24,
			},
			expected: Prefix{
				addr:   IPv4(0),
				pfxlen: 0,
			},
		},
		{
			name: "Supernet of 2001:678:1e0:100:23::/64 and 2001:678:1e0:1ff::/64 -> 2001:678:1e0:100::/56",
			a: Prefix{
				addr:   IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0x100, 0x23, 0, 0, 0),
				pfxlen: 64,
			},
			b: Prefix{
				addr:   IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0x1ff, 0, 0, 0, 0),
				pfxlen: 64,
			},
			expected: Prefix{
				addr:   IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0x100, 0, 0, 0, 0),
				pfxlen: 56,
			},
		},
		{
			name: "Supernet of 2001:678:1e0::/128 and 2001:678:1e0::1/128 -> 2001:678:1e0:100::/127",
			a: Prefix{
				addr:   IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0, 0, 0, 0, 0),
				pfxlen: 128,
			},
			b: Prefix{
				addr:   IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0, 0, 0, 0, 1),
				pfxlen: 128,
			},
			expected: Prefix{
				addr:   IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0, 0, 0, 0, 0),
				pfxlen: 127,
			},
		},
		{
			name: "Supernet of all ones and all zeros -> ::/0",
			a: Prefix{
				addr:   IPv6FromBlocks(0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF),
				pfxlen: 128,
			},
			b: Prefix{
				addr:   IPv6(0, 0),
				pfxlen: 128,
			},
			expected: Prefix{
				addr:   IPv6FromBlocks(0, 0, 0, 0, 0, 0, 0, 0),
				pfxlen: 0,
			},
		},
	}

	t.Parallel()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := test.a.GetSupernet(test.b)
			assert.Equal(t, test.expected, s)
		})
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
				addr:   IPv4(0),
				pfxlen: 0,
			},
			b: Prefix{
				addr:   IPv4(100),
				pfxlen: 24,
			},
			expected: true,
		},
		{
			name: "Test 2",
			a: Prefix{
				addr:   IPv4(100),
				pfxlen: 24,
			},
			b: Prefix{
				addr:   IPv4(0),
				pfxlen: 0,
			},
			expected: false,
		},
		{
			name: "Test 3",
			a: Prefix{
				addr:   IPv4(167772160),
				pfxlen: 8,
			},
			b: Prefix{
				addr:   IPv4(167772160),
				pfxlen: 9,
			},
			expected: true,
		},
		{
			name: "Test 4",
			a: Prefix{
				addr:   IPv4(167772160),
				pfxlen: 8,
			},
			b: Prefix{
				addr:   IPv4(174391040),
				pfxlen: 24,
			},
			expected: true,
		},
		{
			name: "Test 5",
			a: Prefix{
				addr:   IPv4(167772160),
				pfxlen: 8,
			},
			b: Prefix{
				addr:   IPv4(184549377),
				pfxlen: 24,
			},
			expected: false,
		},
		{
			name: "Test 6",
			a: Prefix{
				addr:   IPv4(167772160),
				pfxlen: 8,
			},
			b: Prefix{
				addr:   IPv4(191134464),
				pfxlen: 24,
			},
			expected: false,
		},
		{
			name: "Test 7",
			a: Prefix{
				addr:   IPv4FromOctets(169, 0, 0, 0),
				pfxlen: 25,
			},
			b: Prefix{
				addr:   IPv4FromOctets(169, 1, 1, 0),
				pfxlen: 26,
			},
			expected: false,
		},
		{
			name: "IPv6: 2001:678:1e0:100::/56 is subnet of 2001:678:1e0::/48",
			a: Prefix{
				addr:   IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0, 0, 0, 0, 0),
				pfxlen: 48,
			},
			expected: true,
			b: Prefix{
				addr:   IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0x100, 0, 0, 0, 0),
				pfxlen: 56,
			},
		},
		{
			name: "IPv6: 2001:678:1e0:100::/56 is subnet of 2001:678:1e0::/48",
			a: Prefix{
				addr:   IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0x200, 0, 0, 0, 0),
				pfxlen: 56,
			},
			b: Prefix{
				addr:   IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0x100, 0, 0, 0, 0),
				pfxlen: 64,
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
			a:        NewPfx(IPv4(100), 8),
			b:        NewPfx(IPv4(100), 8),
			expected: true,
		},
		{
			name:     "Unequal PFXs",
			a:        NewPfx(IPv4(100), 8),
			b:        NewPfx(IPv4(200), 8),
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
			pfx:      NewPfx(IPv4FromOctets(10, 0, 0, 0), 8),
			expected: "10.0.0.0/8",
		},
		{
			name:     "Test 2",
			pfx:      NewPfx(IPv4FromOctets(10, 0, 0, 0), 16),
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
			name:     "Non numeric",
			input:    "10.10.10.a",
			wantFail: true,
		},
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

func TestEqualOperator(t *testing.T) {
	p1 := NewPfx(IPv4(100), 4)
	p2 := NewPfx(IPv4(100), 4)

	if p1 != p2 {
		assert.Fail(t, "p1 != p2 (even if attributes are equal)")
	}
}
