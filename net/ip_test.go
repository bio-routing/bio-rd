package net

import (
	"math"
	"net"
	"testing"

	"github.com/bio-routing/bio-rd/net/api"
	"github.com/stretchr/testify/assert"
)

func TestLower(t *testing.T) {
	tests := []struct {
		name     string
		ip       *IP
		expected uint64
	}{
		{
			name:     "Test",
			ip:       &IP{lower: 100},
			expected: 100,
		},
	}

	for _, test := range tests {
		res := test.ip.Lower()
		assert.Equal(t, test.expected, res, test.name)
	}
}

func TestHigher(t *testing.T) {
	tests := []struct {
		name     string
		ip       *IP
		expected uint64
	}{
		{
			name:     "Test",
			ip:       &IP{higher: 200},
			expected: 200,
		},
	}

	for _, test := range tests {
		res := test.ip.Higher()
		assert.Equal(t, test.expected, res, test.name)
	}
}

func TestIPVersion(t *testing.T) {
	tests := []struct {
		name     string
		ip       IP
		expected bool
	}{
		{
			name:     "Test",
			ip:       IPv4(0),
			expected: true,
		},
		{
			name:     "Test",
			ip:       IPv6(0, 0),
			expected: false,
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.ip.isLegacy, test.name)
	}
}
func TestIPToProto(t *testing.T) {
	tests := []struct {
		name     string
		ip       *IP
		expected *api.IP
	}{
		{
			name: "IPv4",
			ip: &IP{
				lower:    255,
				isLegacy: true,
			},
			expected: &api.IP{
				Lower:   255,
				Version: api.IP_IPv4,
			},
		},
		{
			name: "IPv6",
			ip: &IP{
				higher:   1000,
				lower:    255,
				isLegacy: false,
			},
			expected: &api.IP{
				Higher:  1000,
				Lower:   255,
				Version: api.IP_IPv6,
			},
		},
	}

	for _, test := range tests {
		res := test.ip.ToProto()
		assert.Equal(t, test.expected, res, test.name)
	}
}

func TestIPFromProtoIP(t *testing.T) {
	tests := []struct {
		name     string
		proto    api.IP
		expected *IP
	}{
		{
			name: "Test IPv4",
			proto: api.IP{
				Lower:   100,
				Higher:  0,
				Version: api.IP_IPv4,
			},
			expected: &IP{
				lower:    100,
				higher:   0,
				isLegacy: true,
			},
		},
		{
			name: "Test IPv6",
			proto: api.IP{
				Lower:   100,
				Higher:  200,
				Version: api.IP_IPv6,
			},
			expected: &IP{
				lower:    100,
				higher:   200,
				isLegacy: false,
			},
		},
	}

	for i := range tests {
		res := IPFromProtoIP(&tests[i].proto)
		assert.Equal(t, tests[i].expected, res, tests[i].name)
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		name     string
		ip       *IP
		other    *IP
		expected int8
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

func TestIPString(t *testing.T) {
	tests := []struct {
		ip       IP
		expected string
	}{
		{
			ip:       IPv4FromOctets(192, 168, 0, 1),
			expected: "192.168.0.1",
		},
		{
			ip:       IPv4FromOctets(0, 0, 0, 0),
			expected: "0.0.0.0",
		},
		{
			ip:       IPv4FromOctets(255, 255, 255, 255),
			expected: "255.255.255.255",
		},
		{
			ip:       IPv6(0, 0),
			expected: "0:0:0:0:0:0:0:0",
		},
		{
			ip:       IPv6(2306131596687708724, 6230974922281175806),
			expected: "2001:678:1E0:1234:5678:DEAD:BEEF:CAFE",
		},
		{
			ip:       IPv6(^uint64(0), ^uint64(0)),
			expected: "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFF",
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.ip.String())
	}
}

func TestBytes(t *testing.T) {
	tests := []struct {
		name     string
		ip       IP
		expected []byte
	}{
		{
			name:     "IPv4 172.217.16.195",
			ip:       IPv4(2899906755),
			expected: []byte{172, 217, 16, 195},
		},
		{
			name:     "IPv6 2001:678:1E0:1234:5678:DEAD:BEEF:CAFE",
			ip:       IPv6(2306131596687708724, 6230974922281175806),
			expected: []byte{32, 1, 6, 120, 1, 224, 18, 52, 86, 120, 222, 173, 190, 239, 202, 254},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.ip.Bytes())
		})
	}
}

func TestIPv4FromOctets(t *testing.T) {
	tests := []struct {
		name     string
		octets   []uint8
		expected IP
	}{
		{
			name:   "172.217.16.195",
			octets: []uint8{172, 217, 16, 195},
			expected: IP{
				higher:   0,
				lower:    2899906755,
				isLegacy: true,
			},
		},
		{
			name:   "0.0.0.0",
			octets: []uint8{0, 0, 0, 0},
			expected: IP{
				higher:   0,
				lower:    0,
				isLegacy: true,
			},
		},
		{
			name:   "255.255.255.255",
			octets: []uint8{255, 255, 255, 255},
			expected: IP{
				higher:   0,
				lower:    math.MaxUint32,
				isLegacy: true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, IPv4FromOctets(test.octets[0], test.octets[1], test.octets[2], test.octets[3]))
		})
	}
}

func TestIPv6FromBlocks(t *testing.T) {
	tests := []struct {
		name     string
		blocks   []uint16
		expected IP
	}{
		{
			name: "IPv6 2001:678:1E0:1234:5678:DEAD:BEEF:CAFE",
			blocks: []uint16{
				0x2001,
				0x678,
				0x1e0,
				0x1234,
				0x5678,
				0xdead,
				0xbeef,
				0xcafe,
			},
			expected: IP{
				higher: 2306131596687708724,
				lower:  6230974922281175806,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, IPv6FromBlocks(
				test.blocks[0],
				test.blocks[1],
				test.blocks[2],
				test.blocks[3],
				test.blocks[4],
				test.blocks[5],
				test.blocks[6],
				test.blocks[7]))
		})
	}
}

func TestIPFromBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    []byte
		expected IP
		wantFail bool
	}{
		{
			name:  "IPV4: 172.217.16.195",
			bytes: []byte{172, 217, 16, 195},
			expected: IP{
				higher:   0,
				lower:    2899906755,
				isLegacy: true,
			},
		},
		{
			name:  "IPV6: IPv6 2001:678:1E0:1234:5678:DEAD:BEEF:CAFE",
			bytes: []byte{0x20, 0x01, 0x06, 0x78, 0x01, 0xE0, 0x12, 0x34, 0x56, 0x78, 0xDE, 0xAD, 0xBE, 0xEF, 0xCA, 0xFE},
			expected: IP{
				higher: 2306131596687708724,
				lower:  6230974922281175806,
			},
		},
		{
			name:     "invalid length",
			bytes:    []byte{172, 217, 123, 231, 95},
			wantFail: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ip, err := IPFromBytes(test.bytes)
			if err == nil && test.wantFail {
				t.Fatalf("Expected test to fail, but did not")
			}

			if test.wantFail {
				if err == nil {
					t.Fatalf("Unexpected success")
				}
				return
			}

			assert.Equal(t, test.expected, ip)
		})
	}
}

func TestToNetIP(t *testing.T) {
	tests := []struct {
		name     string
		ip       IP
		expected net.IP
	}{
		{
			name:     "IPv4",
			ip:       IPv4FromOctets(192, 168, 1, 1),
			expected: net.IP{192, 168, 1, 1},
		},
		{
			name: "IPv6",
			ip: IPv6FromBlocks(
				0x2001,
				0x678,
				0x1e0,
				0x1234,
				0x5678,
				0xdead,
				0xbeef,
				0xcafe),
			expected: net.IP{32, 1, 6, 120, 1, 224, 18, 52, 86, 120, 222, 173, 190, 239, 202, 254},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.ip.ToNetIP())
		})
	}
}

func TestBitAtPosition(t *testing.T) {
	tests := []struct {
		name     string
		input    IP
		position uint8
		expected bool
	}{
		{
			name:     "IPv4: all ones -> 0",
			input:    IPv4FromOctets(255, 255, 255, 255),
			position: 1,
			expected: true,
		},
		{
			name:     "IPv4: Bit 8 from 1.0.0.0 -> 0",
			input:    IPv4FromOctets(10, 0, 0, 0),
			position: 8,
			expected: false,
		},
		{
			name:     "IPv4: Bit 8 from 11.0.0.0 -> 1",
			input:    IPv4FromOctets(11, 0, 0, 0),
			position: 8,
			expected: true,
		},
		{
			name:     "IPv6: Bit 16 from 2001:678:1e0:: -> 1",
			input:    IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0, 0, 0, 0, 0),
			position: 16,
			expected: true,
		},
		{
			name:     "IPv6: Bit 17 from 2001:678:1e0:: -> 0",
			input:    IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0, 0, 0, 0, 0),
			position: 17,
			expected: false,
		},
		{
			name:     "IPv6: Bit 113 from 2001:678:1e0::cafe -> 1",
			input:    IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0, 0, 0, 0, 0xcafe),
			position: 113,
			expected: true,
		},
		{
			name:     "IPv6: Bit 115 from 2001:678:1e0::cafe -> 0",
			input:    IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0, 0, 0, 0, 0xcafe),
			position: 115,
			expected: false,
		},
		{
			name:     "IPv6: all ones -> 1",
			input:    IPv6FromBlocks(0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF),
			position: 1,
			expected: true,
		},
		{
			name:     "IPv4: invalid position",
			input:    IPv4(0),
			position: 33,
			expected: false,
		},
		{
			name:     "IPv6: invalid position",
			input:    IPv6(0, 0),
			position: 129,
			expected: false,
		},
	}

	for _, test := range tests {
		b := test.input.BitAtPosition(test.position)
		assert.Equal(t, test.expected, b, test.name)
	}
}

func TestIPFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected IP
		wantFail bool
	}{
		{
			name:     "ipv4",
			input:    "192.168.1.234",
			expected: IPv4FromOctets(192, 168, 1, 234),
		},
		{
			name:     "ipv6",
			input:    "2001:678:1e0::cafe",
			expected: IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0, 0, 0, 0, 0xcafe),
		},
		{
			name:     "invalid",
			input:    "foo",
			wantFail: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ip, err := IPFromString(test.input)
			if err == nil && test.wantFail {
				t.Fatal("expected error but got nil")
			}
			if err != nil {
				if test.wantFail {
					return
				}

				t.Fatal(err)
			}

			assert.Equal(t, test.expected, ip)
		})
	}
}

func TestSizeBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    IP
		expected uint8
	}{
		{
			name:     "IPv4",
			input:    IPv4(0),
			expected: 4,
		},
		{
			name:     "IPv6",
			input:    IPv6(0, 0),
			expected: 16,
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.input.SizeBytes(), test.name)
	}
}

func TestNext(t *testing.T) {
	tests := []struct {
		name     string
		input    *IP
		expected *IP
	}{
		{
			name:     "Test #1",
			input:    IPv4FromOctets(10, 0, 0, 1).Dedup(),
			expected: IPv4FromOctets(10, 0, 0, 2).Dedup(),
		},
		{
			name:     "Test #2",
			input:    IPv6FromBlocks(10, 20, 30, 40, 50, 60, 70, 80).Dedup(),
			expected: IPv6FromBlocks(10, 20, 30, 40, 50, 60, 70, 81).Dedup(),
		},
		{
			name:     "Test #3",
			input:    IPv6FromBlocks(10, 20, 30, 40, 0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF).Dedup(),
			expected: IPv6FromBlocks(10, 20, 30, 41, 0, 0, 0, 0).Dedup(),
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.input.Next(), test.name)
	}
}

func TestMaskLastNBits(t *testing.T) {
	tests := []struct {
		name     string
		input    *IP
		maskBits uint8
		expected *IP
	}{
		{
			name:     "Test #1",
			input:    IPv4FromOctets(10, 1, 1, 1).Dedup(),
			maskBits: 8,
			expected: IPv4FromOctets(10, 1, 1, 0).Dedup(),
		},
		{
			name:     "Test #2",
			input:    IPv4FromOctets(185, 65, 241, 123).Dedup(),
			maskBits: 9,
			expected: IPv4FromOctets(185, 65, 240, 0).Dedup(),
		},
		{
			name:     "Test #3",
			input:    IPv4FromOctets(185, 65, 241, 123).Dedup(),
			maskBits: 32,
			expected: IPv4FromOctets(0, 0, 0, 0).Dedup(),
		},
		{
			name:     "Test #4",
			input:    IPv6FromBlocks(0x2001, 0xaaaa, 0x1234, 0x2222, 0x1111, 0x3333, 0xbbbb, 0xacab).Dedup(),
			maskBits: 16,
			expected: IPv6FromBlocks(0x2001, 0xaaaa, 0x1234, 0x2222, 0x1111, 0x3333, 0xbbbb, 0x0000).Dedup(),
		},
		{
			name:     "Test #5",
			input:    IPv6FromBlocks(0x2001, 0xaaaa, 0x1234, 0x2222, 0x1111, 0x3333, 0xbbbb, 0xacab).Dedup(),
			maskBits: 64,
			expected: IPv6FromBlocks(0x2001, 0xaaaa, 0x1234, 0x2222, 0x0000, 0x0000, 0x0000, 0x0000).Dedup(),
		},
		{
			name:     "Test #6",
			input:    IPv6FromBlocks(0x2001, 0xaaaa, 0x1234, 0x2222, 0x1111, 0x3333, 0xbbbb, 0xacab).Dedup(),
			maskBits: 80,
			expected: IPv6FromBlocks(0x2001, 0xaaaa, 0x1234, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000).Dedup(),
		},
		{
			name:     "Test #7",
			input:    IPv6FromBlocks(0x2001, 0xaaaa, 0x1234, 0x2222, 0x1111, 0x3333, 0xbbbb, 0xacab).Dedup(),
			maskBits: 128,
			expected: IPv6FromBlocks(0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000).Dedup(),
		},
	}

	for _, test := range tests {
		res := test.input.MaskLastNBits(test.maskBits)
		assert.Equal(t, test.expected, res, test.name)
	}
}
