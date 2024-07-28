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
		assert.Equal(t, test.expected, *res, test.name)
	}
}

func TestPrefixToProto(t *testing.T) {
	tests := []struct {
		name     string
		pfx      Prefix
		expected *api.Prefix
	}{
		{
			name: "IPv4",
			pfx: Prefix{
				addr: IP{
					lower:    200,
					isLegacy: true,
				},
				len: 24,
			},
			expected: &api.Prefix{
				Address: &api.IP{
					Lower:   200,
					Version: api.IP_IPv4,
				},
				Length: 24,
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
				len: 64,
			},
			expected: &api.Prefix{
				Address: &api.IP{
					Higher:  100,
					Lower:   200,
					Version: api.IP_IPv6,
				},
				Length: 64,
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
				Length: 24,
			},
			expected: Prefix{
				addr: IP{
					higher:   0,
					lower:    2000,
					isLegacy: true,
				},
				len: 24,
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
				Length: 64,
			},
			expected: Prefix{
				addr: IP{
					higher:   1000,
					lower:    2000,
					isLegacy: false,
				},
				len: 64,
			},
		},
	}

	for i := range tests {
		res := NewPrefixFromProtoPrefix(&tests[i].proto)
		assert.Equal(t, tests[i].expected, *res, tests[i].name)
	}
}

func TestNewPfx(t *testing.T) {
	p := NewPfx(IPv4(123), 11)
	if p.addr != IPv4(123) || p.len != 11 {
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
		assert.Equal(t, res, test.expected, "Unexpected result for test %s", test.name)
	}
}

func TestLength(t *testing.T) {
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
		res := test.pfx.Len()
		assert.Equal(t, res, test.expected, "Unexpected result for test %s", test.name)
	}
}

func TestGetSupernet(t *testing.T) {
	tests := []struct {
		name     string
		a        *Prefix
		b        *Prefix
		expected *Prefix
	}{
		{
			name: "Supernet of 10.0.0.0 and 11.100.123.0 -> 10.0.0.0/7",
			a: &Prefix{
				addr: IPv4FromOctets(10, 0, 0, 0),
				len:  8,
			},
			b: &Prefix{
				addr: IPv4FromOctets(11, 100, 123, 0),
				len:  24,
			},
			expected: &Prefix{
				addr: IPv4FromOctets(10, 0, 0, 0),
				len:  7,
			},
		},
		{
			name: "Supernet of 10.0.0.0 and 192.168.0.0 -> 0.0.0.0/0",
			a: &Prefix{
				addr: IPv4FromOctets(10, 0, 0, 0),
				len:  8,
			},
			b: &Prefix{
				addr: IPv4FromOctets(192, 168, 0, 0),
				len:  24,
			},
			expected: &Prefix{
				addr: IPv4(0),
				len:  0,
			},
		},
		{
			name: "Supernet of 2001:678:1e0:100:23::/64 and 2001:678:1e0:1ff::/64 -> 2001:678:1e0:100::/56",
			a: &Prefix{
				addr: IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0x100, 0x23, 0, 0, 0),
				len:  64,
			},
			b: &Prefix{
				addr: IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0x1ff, 0, 0, 0, 0),
				len:  64,
			},
			expected: &Prefix{
				addr: IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0x100, 0, 0, 0, 0),
				len:  56,
			},
		},
		{
			name: "Supernet of 2001:678:1e0::/128 and 2001:678:1e0::1/128 -> 2001:678:1e0:100::/127",
			a: &Prefix{
				addr: IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0, 0, 0, 0, 0),
				len:  128,
			},
			b: &Prefix{
				addr: IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0, 0, 0, 0, 1),
				len:  128,
			},
			expected: &Prefix{
				addr: IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0, 0, 0, 0, 0),
				len:  127,
			},
		},
		{
			name: "Supernet of all ones and all zeros -> ::/0",
			a: &Prefix{
				addr: IPv6FromBlocks(0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF),
				len:  128,
			},
			b: &Prefix{
				addr: IPv6(0, 0),
				len:  128,
			},
			expected: &Prefix{
				addr: IPv6FromBlocks(0, 0, 0, 0, 0, 0, 0, 0),
				len:  0,
			},
		},
	}

	t.Parallel()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := test.a.GetSupernet(test.b)
			assert.Equal(t, *test.expected, s)
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		a        *Prefix
		b        *Prefix
		expected bool
	}{
		{
			a: &Prefix{
				addr: IPv4(0),
				len:  0,
			},
			b: &Prefix{
				addr: IPv4(100),
				len:  24,
			},
			expected: true,
		},
		{
			a: &Prefix{
				addr: IPv4(100),
				len:  24,
			},
			b: &Prefix{
				addr: IPv4(0),
				len:  0,
			},
			expected: false,
		},
		{
			a: &Prefix{
				addr: IPv4(167772160),
				len:  8,
			},
			b: &Prefix{
				addr: IPv4(167772160),
				len:  9,
			},
			expected: true,
		},
		{
			a: &Prefix{
				addr: IPv4(167772160),
				len:  8,
			},
			b: &Prefix{
				addr: IPv4(174391040),
				len:  24,
			},
			expected: true,
		},
		{
			a: &Prefix{
				addr: IPv4(167772160),
				len:  8,
			},
			b: &Prefix{
				addr: IPv4(184549377),
				len:  24,
			},
			expected: false,
		},
		{
			a: &Prefix{
				addr: IPv4(167772160),
				len:  8,
			},
			b: &Prefix{
				addr: IPv4(191134464),
				len:  24,
			},
			expected: false,
		},
		{
			a: &Prefix{
				addr: IPv4FromOctets(169, 0, 0, 0),
				len:  25,
			},
			b: &Prefix{
				addr: IPv4FromOctets(169, 1, 1, 0),
				len:  26,
			},
			expected: false,
		},
		{
			a: &Prefix{
				addr: IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0, 0, 0, 0, 0),
				len:  48,
			},
			expected: true,
			b: &Prefix{
				addr: IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0x100, 0, 0, 0, 0),
				len:  56,
			},
		},
		{
			a: &Prefix{
				addr: IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0x200, 0, 0, 0, 0),
				len:  56,
			},
			b: &Prefix{
				addr: IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0x100, 0, 0, 0, 0),
				len:  64,
			},
			expected: false,
		},
		{
			a: &Prefix{
				addr: IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0x200, 0, 0, 0, 0),
				len:  65,
			},
			b: &Prefix{
				addr: IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0x100, 0, 0, 0, 0),
				len:  64,
			},
			expected: false,
		},
		{
			a: &Prefix{
				addr: IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0x100, 100, 0, 0, 0),
				len:  72,
			},
			b: &Prefix{
				addr: IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0x100, 100, 0, 0, 1),
				len:  127,
			},
			expected: true,
		},
		{
			a: &Prefix{
				addr: IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0x100, 100, 0, 0, 0),
				len:  126,
			},
			b: &Prefix{
				addr: IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0x100, 100, 0, 100, 1),
				len:  127,
			},
			expected: false,
		},
	}

	for _, test := range tests {
		res := test.a.Contains(test.b)
		assert.Equal(t, res, test.expected, "Subnet %s contains %s", test.a, test.b)
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
		assert.Equal(t, res, test.expected, "Unexpected result for test %s", test.name)
	}
}

func TestEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        *Prefix
		b        *Prefix
		expected bool
	}{
		{
			name:     "Equal PFXs",
			a:        NewPfx(IPv4(100), 8).Ptr(),
			b:        NewPfx(IPv4(100), 8).Ptr(),
			expected: true,
		},
		{
			name:     "Unequal PFXs",
			a:        NewPfx(IPv4(100), 8).Ptr(),
			b:        NewPfx(IPv4(200), 8).Ptr(),
			expected: false,
		},
	}

	for _, test := range tests {
		res := test.a.Equal(test.b)
		assert.Equal(t, res, test.expected, "Unexpected result for %q", test.name)
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
		assert.Equal(t, res, test.expected, "Unexpected result for %q")
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

	assert.Equal(t, p1, p2, "p1 != p2 (even if attributes are equal)")
}

func TestValid(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Test #1",
			input:    "169.254.100.0/24",
			expected: true,
		},
		{
			name:     "Test #2",
			input:    "169.254.100.1/24",
			expected: false,
		},
		{
			name:     "Test #2a",
			input:    "169.254.100.111/32",
			expected: true,
		},
		{
			name:     "Test #3",
			input:    "2a05:1234:abcd::/48",
			expected: true,
		},
		{
			name:     "Test #4",
			input:    "2a05:1234:abcd::1337/48",
			expected: false,
		},
		{
			name:     "Test #5",
			input:    "2a05:1234:abcd:face::/64",
			expected: true,
		},
		{
			name:     "Test #6",
			input:    "2a05:1234:abcd:face::a/64",
			expected: false,
		},
		{
			name:     "Test #7",
			input:    "2a05:1234:abcd:face:b00c::/80",
			expected: true,
		},
		{
			name:     "Test #8",
			input:    "2a05:1234:abcd:face:b00c::aa/72",
			expected: false,
		},
		{
			name:     "Test #9",
			input:    "2a05:1234:abcd:face:b00c::aa/128",
			expected: true,
		},
	}

	for _, test := range tests {
		p, _ := PrefixFromString(test.input)
		assert.Equal(t, test.expected, p.Valid(), test.name)
	}
}

func TestBaseAddr(t *testing.T) {
	tests := []struct {
		name     string
		input    *Prefix
		expected IP
	}{
		{
			name:     "Test #1",
			input:    NewPfx(IPv4FromOctets(10, 1, 1, 0), 23).Ptr(),
			expected: IPv4FromOctets(10, 1, 0, 0),
		},
		{
			name:     "Test #2",
			input:    NewPfx(IPv4FromOctets(10, 1, 1, 2), 24).Ptr(),
			expected: IPv4FromOctets(10, 1, 1, 0),
		},
		{
			name:     "Test #3",
			input:    NewPfx(IPv6FromBlocks(10, 10, 20, 20, 1, 0, 0, 1), 64).Ptr(),
			expected: IPv6FromBlocks(10, 10, 20, 20, 0, 0, 0, 0),
		},
		{
			name:     "Test #4",
			input:    NewPfx(IPv6FromBlocks(10, 10, 20, 20, 1, 0, 0, 1), 48).Ptr(),
			expected: IPv6FromBlocks(10, 10, 20, 0, 0, 0, 0, 0),
		},
		{
			name:     "Test #5",
			input:    NewPfx(IPv6FromBlocks(10, 10, 20, 20, 1, 0, 5, 1), 126).Ptr(),
			expected: IPv6FromBlocks(10, 10, 20, 20, 1, 0, 5, 0),
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.input.BaseAddr(), test.name)
	}
}

func TestPrefixFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantFail bool
	}{
		{
			name:     "valid prefix",
			input:    "2a05:1234:abcd:face:b00c::aa/128",
			wantFail: false,
		},
		{
			name:     "No / in prefix",
			input:    "2a05:1234:abcd:face:b00c::aa",
			wantFail: true,
		},
		{
			name:     "invalid IP",
			input:    "2a05:1234:abcd:face:b00c::aax/123",
			wantFail: true,
		},
		{
			name:     "invalid prefix length",
			input:    "2a05:1234:abcd:face:b00c::aax/123foo",
			wantFail: true,
		},
	}

	for _, test := range tests {
		_, err := PrefixFromString(test.input)
		assert.Equal(t, test.wantFail, err != nil, test.name)
	}
}

func TestBytesInAddr(t *testing.T) {
	tests := []struct {
		name     string
		input    uint8
		expected uint8
	}{
		{
			name:     "Test #1",
			input:    24,
			expected: 3,
		},
		{
			name:     "Test #2",
			input:    25,
			expected: 4,
		},
		{
			name:     "Test #3",
			input:    32,
			expected: 4,
		},
		{
			name:     "Test #4",
			input:    0,
			expected: 0,
		},
		{
			name:     "Test #5",
			input:    9,
			expected: 2,
		},
	}

	for _, test := range tests {
		p := &Prefix{
			len: test.input,
		}
		res := p.BytesInPrefix()
		if res != test.expected {
			t.Errorf("Unexpected result for test %q: %d", test.name, res)
		}
	}
}
