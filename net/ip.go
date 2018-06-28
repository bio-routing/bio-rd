package net

import (
	"fmt"
)

// IP represents an IPv4 or IPv6 address
type IP struct {
	higher    uint64
	lower     uint64
	ipVersion uint8
}

// IPv4 returns a new `IP` representing an IPv4 address
func IPv4(val uint32) IP {
	return IP{
		lower:     uint64(val),
		ipVersion: 4,
	}
}

// IPv4FromOctets returns an IPv4 address for the given 4 octets
func IPv4FromOctets(o1, o2, o3, o4 uint8) IP {
	return IPv4(uint32(o1)<<24 + uint32(o2)<<16 + uint32(o3)<<8 + uint32(o4))
}

// IPv6 returns a new `IP` representing an IPv6 address
func IPv6(higher, lower uint64) IP {
	return IP{
		higher:    higher,
		lower:     lower,
		ipVersion: 6,
	}
}

// IPv6FromBlocks returns an IPv6 address for the given 8 blocks
func IPv6FromBlocks(b1, b2, b3, b4, b5, b6, b7, b8 uint16) IP {
	return IPv6(
		uint64(uint64(b1)<<48+uint64(b2)<<32+uint64(b3)<<16+uint64(b4)),
		uint64(uint64(b5)<<48+uint64(b6)<<32+uint64(b7)<<16+uint64(b8)))
}

// ToUint32 returns the uint32 representation of an IP address
func (ip *IP) ToUint32() uint32 {
	return uint32(^uint64(0) >> 32 & ip.lower)
}

// Compare compares two IP addresses (returns 0 if equal, -1 if `ip` is smaller than `other`, 1 if `ip` is greater than `other`)
func (ip IP) Compare(other IP) int {
	if ip == other {
		return 0
	}

	if ip.higher > other.higher {
		return 1
	}

	if ip.higher < other.higher {
		return -1
	}

	if ip.lower > other.lower {
		return 1
	}

	return -1
}

func (ip IP) String() string {
	if ip.ipVersion == 6 {
		return ip.stringIPv6()
	}

	return ip.stringIPv4()
}

func (ip IP) stringIPv6() string {
	fmt.Println(ip.higher & 0xFFFF000000000000 >> 48)

	return fmt.Sprintf("%X:%X:%X:%X:%X:%X:%X:%X",
		ip.higher&0xFFFF000000000000>>48,
		ip.higher&0x0000FFFF00000000>>32,
		ip.higher&0x00000000FFFF0000>>16,
		ip.higher&0x000000000000FFFF,
		ip.lower&0xFFFF000000000000>>48,
		ip.lower&0x0000FFFF00000000>>32,
		ip.lower&0x00000000FFFF0000>>16,
		ip.lower&0x000000000000FFFF)
}

func (ip IP) stringIPv4() string {
	b := ip.Bytes()

	return fmt.Sprintf("%d.%d.%d.%d", b[0], b[1], b[2], b[3])
}

// Bytes returns the byte representation of an IP address
func (ip IP) Bytes() []byte {
	if ip.ipVersion == 6 {
		return ip.bytesIPv6()
	}

	return ip.bytesIPv4()
}

func (ip IP) bytesIPv4() []byte {
	u := ip.ToUint32()
	return []byte{
		byte(u & 0xFF000000 >> 24),
		byte(u & 0x00FF0000 >> 16),
		byte(u & 0x0000FF00 >> 8),
		byte(u & 0x000000FF),
	}
}

func (ip IP) bytesIPv6() []byte {
	return []byte{
		byte(ip.higher & 0xFF00000000000000 >> 56),
		byte(ip.higher & 0x00FF000000000000 >> 48),
		byte(ip.higher & 0x0000FF0000000000 >> 40),
		byte(ip.higher & 0x000000FF00000000 >> 32),
		byte(ip.higher & 0x00000000FF000000 >> 24),
		byte(ip.higher & 0x0000000000FF0000 >> 16),
		byte(ip.higher & 0x000000000000FF00 >> 8),
		byte(ip.higher & 0x00000000000000FF),
		byte(ip.lower & 0xFF00000000000000 >> 56),
		byte(ip.lower & 0x00FF000000000000 >> 48),
		byte(ip.lower & 0x0000FF0000000000 >> 40),
		byte(ip.lower & 0x000000FF00000000 >> 32),
		byte(ip.lower & 0x00000000FF000000 >> 24),
		byte(ip.lower & 0x0000000000FF0000 >> 16),
		byte(ip.lower & 0x000000000000FF00 >> 8),
		byte(ip.lower & 0x00000000000000FF),
	}
}
