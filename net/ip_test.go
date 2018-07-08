package net

import (
	"math"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompare(t *testing.T) {
	tests := []struct {
		name     string
		ip       IP
		other    IP
		expected int
	}{
		{
			name: "equal",
			ip: IP{
				lower:  100,
				higher: 200,
			},
			other: IP{
				lower:  100,
				higher: 200,
			},
			expected: 0,
		},
		{
			name: "greater higher word",
			ip: IP{
				lower:  123,
				higher: 200,
			},
			other: IP{
				lower:  456,
				higher: 100,
			},
			expected: 1,
		},
		{
			name: "lesser higher word",
			ip: IP{
				lower:  123,
				higher: 100,
			},
			other: IP{
				lower:  456,
				higher: 200,
			},
			expected: -1,
		},
		{
			name: "equal higher word but lesser lower word",
			ip: IP{
				lower:  456,
				higher: 100,
			},
			other: IP{
				lower:  123,
				higher: 100,
			},
			expected: 1,
		},
		{
			name: "equal higher word but lesser lower word",
			ip: IP{
				lower:  123,
				higher: 100,
			},
			other: IP{
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
				higher:    0,
				lower:     2899906755,
				ipVersion: 4,
			},
		},
		{
			name:   "0.0.0.0",
			octets: []uint8{0, 0, 0, 0},
			expected: IP{
				higher:    0,
				lower:     0,
				ipVersion: 4,
			},
		},
		{
			name:   "255.255.255.255",
			octets: []uint8{255, 255, 255, 255},
			expected: IP{
				higher:    0,
				lower:     math.MaxUint32,
				ipVersion: 4,
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
				higher:    2306131596687708724,
				lower:     6230974922281175806,
				ipVersion: 6,
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
				higher:    0,
				lower:     2899906755,
				ipVersion: 4,
			},
		},
		{
			name:  "IPV6: IPv6 2001:678:1E0:1234:5678:DEAD:BEEF:CAFE",
			bytes: []byte{0x20, 0x01, 0x06, 0x78, 0x01, 0xE0, 0x12, 0x34, 0x56, 0x78, 0xDE, 0xAD, 0xBE, 0xEF, 0xCA, 0xFE},
			expected: IP{
				higher:    2306131596687708724,
				lower:     6230974922281175806,
				ipVersion: 6,
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
	}

	for _, test := range tests {
		b := test.input.BitAtPosition(test.position)
		if b != test.expected {
			t.Errorf("%s: Unexpected failure: Bit %d of %d is %v. Expected %v",
				test.name, test.position, test.input, b, test.expected)
		}
	}
}
