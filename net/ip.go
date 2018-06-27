package net

// IP represents an IPv4 or IPv6 address
type IP struct {
	higher uint64
	lower  uint64
}

// IPv4 returns a new `IP` representing an IPv4 address
func IPv4(val uint32) IP {
	return IP{
		lower: uint64(val),
	}
}

// IPv6 returns a new `IP` representing an IPv6 address
func IPv6(higher, lower uint64) IP {
	return IP{
		higher: higher,
		lower:  lower,
	}
}

// ToUint32 returns the uint32 representation of an IP address
func (ip *IP) ToUint32() uint32 {
	return uint32(^uint64(0) >> 32 & ip.lower)
}

// Compare compares two IP addresses (returns 0 if equal, -1 if `ip` is smaller than `other`, 1 if `ip` is greater than `other`)
func (ip *IP) Compare(other *IP) int {
	if ip.higher == other.higher && ip.lower == other.lower {
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
