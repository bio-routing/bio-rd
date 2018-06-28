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

// IPv6 returns a new `IP` representing an IPv6 address
func IPv6(higher, lower uint64) IP {
	return IP{
		higher:    higher,
		lower:     lower,
		ipVersion: 6,
	}
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
	u := ip.ToUint32()

	return fmt.Sprintf("%d.%d.%d.%d",
		u&0xFF000000>>24,
		u&0x00FF0000>>16,
		u&0x0000FF00>>8,
		u&0x000000FF)
}

func (ip IP) Bytes() {

}
