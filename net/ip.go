package net

import (
	"fmt"
	"math"
	"net"

	api "github.com/bio-routing/bio-rd/net/api"
	bmath "github.com/bio-routing/bio-rd/util/math"
)

// IP represents an IPv4 or IPv6 address
type IP struct {
	higher   uint64
	lower    uint64
	isLegacy bool
}

// Dedup gets a copy of IP from the cache
func (ip IP) Dedup() *IP {
	return ipc.get(ip)
}

// Ptr returns a pointer to ip
func (ip IP) Ptr() *IP {
	return &ip
}

// IPFromProtoIP creates an IP address from a proto IP
func IPFromProtoIP(addr *api.IP) IP {
	return IP{
		higher:   addr.Higher,
		lower:    addr.Lower,
		isLegacy: addr.Version == api.IP_IPv4,
	}
}

// ToProto converts an IP to a proto IP
func (ip IP) ToProto() *api.IP {
	ver := api.IP_IPv6
	if ip.isLegacy {
		ver = api.IP_IPv4
	}

	return &api.IP{
		Lower:   ip.lower,
		Higher:  ip.higher,
		Version: ver,
	}
}

// Lower gets the lower half of the IP address
func (ip *IP) Lower() uint64 {
	return ip.lower
}

// Higher gets the higher half of the IP address
func (ip *IP) Higher() uint64 {
	return ip.higher
}

func (ip *IP) copy() IP {
	return *ip
}

// IPv4 returns a new `IP` representing an IPv4 address
func IPv4(val uint32) IP {
	return IP{
		lower:    uint64(val),
		isLegacy: true,
	}
}

// IPv4FromOctets returns an IPv4 address for the given 4 octets
func IPv4FromOctets(o1, o2, o3, o4 uint8) IP {
	return IPv4(uint32(o1)<<24 + uint32(o2)<<16 + uint32(o3)<<8 + uint32(o4))
}

// IPv6 returns a new `IP` representing an IPv6 address
func IPv6(higher, lower uint64) IP {
	return IP{
		higher:   higher,
		lower:    lower,
		isLegacy: false,
	}
}

// IPv6FromBlocks returns an IPv6 address for the given 8 blocks
func IPv6FromBlocks(b1, b2, b3, b4, b5, b6, b7, b8 uint16) IP {
	return IPv6(
		uint64(uint64(b1)<<48+uint64(b2)<<32+uint64(b3)<<16+uint64(b4)),
		uint64(uint64(b5)<<48+uint64(b6)<<32+uint64(b7)<<16+uint64(b8)))
}

// IPv4FromBytes creates an IPv4 Address from one or more bytes. Missing bytes are filled with zero bytes.
func IPv4FromBytes(b []byte) IP {
	switch len(b) {
	case 0:
		return IPv4FromOctets(0, 0, 0, 0)
	case 1:
		return IPv4FromOctets(b[0], 0, 0, 0)
	case 2:
		return IPv4FromOctets(b[0], b[1], 0, 0)
	case 3:
		return IPv4FromOctets(b[0], b[1], b[2], 0)
	case 4:
		return IPv4FromOctets(b[0], b[1], b[2], b[3])
	}

	return IP{}
}

// IPFromBytes returns an IP address for a byte slice
func IPFromBytes(b []byte) (IP, error) {
	ip4 := net.IP(b).To4()
	if ip4 != nil {
		return IPv4FromOctets(ip4[0], ip4[1], ip4[2], ip4[3]), nil
	}

	if len(b) == 16 {
		return IPv6FromBlocks(
			uint16(b[0])<<8+uint16(b[1]),
			uint16(b[2])<<8+uint16(b[3]),
			uint16(b[4])<<8+uint16(b[5]),
			uint16(b[6])<<8+uint16(b[7]),
			uint16(b[8])<<8+uint16(b[9]),
			uint16(b[10])<<8+uint16(b[11]),
			uint16(b[12])<<8+uint16(b[13]),
			uint16(b[14])<<8+uint16(b[15])), nil
	}

	return IP{}, fmt.Errorf("byte slice has an invalid length. Expected either 4 (IPv4) or 16 (IPv6) bytes but got: %d", len(b))
}

// IPFromString returns an IP address for a given string
func IPFromString(str string) (IP, error) {
	ip := net.ParseIP(str)
	if ip == nil {
		return IP{}, fmt.Errorf("%s is not a valid IP address", str)
	}

	ip4 := ip.To4()
	if ip4 != nil {
		return IPFromBytes(ip4)
	}

	return IPFromBytes(ip.To16())
}

// Equal returns true if ip is equal to other
func (ip *IP) Equal(other IP) bool {
	return *ip == other
}

// Compare compares two IP addresses (returns 0 if equal, -1 if `ip` is smaller than `other`, 1 if `ip` is greater than `other`)
func (ip *IP) Compare(other *IP) int8 {
	if ip.higher > other.higher {
		return 1
	}

	if ip.higher < other.higher {
		return -1
	}

	if ip.lower > other.lower {
		return 1
	}

	if ip.lower < other.lower {
		return -1
	}

	return 0
}

// String returns string representation of an IP address
func (ip IP) String() string {
	if !ip.isLegacy {
		return ip.stringIPv6()
	}

	return ip.stringIPv4()
}

func (ip IP) stringIPv6() string {
	p := ip.Bytes()
	// Find longest run of zeros.
	e0 := -1
	e1 := -1
	for i := 0; i < int(ip.SizeBytes()); i += 2 {
		j := i
		for j < int(ip.SizeBytes()) && p[j] == 0 && p[j+1] == 0 {
			j += 2
		}
		if j > i && j-i > e1-e0 {
			e0 = i
			e1 = j
			i = j
		}
	}
	// The symbol "::" MUST NOT be used to shorten just one 16 bit 0 field.
	if e1-e0 <= 2 {
		e0 = -1
		e1 = -1
	}

	const maxLen = len("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff")
	b := make([]byte, 0, maxLen)

	// Print with possible :: in place of run of zeros
	for i := 0; i < int(ip.SizeBytes()); i += 2 {
		if i == e0 {
			b = append(b, ':', ':')
			i = e1
			if i >= int(ip.SizeBytes()) {
				break
			}
		} else if i > 0 {
			b = append(b, ':')
		}
		b = appendHex(b, (uint32(p[i])<<8)|uint32(p[i+1]))
	}

	return string(b)
}

func (ip IP) stringIPv4() string {
	b := ip.Bytes()

	return fmt.Sprintf("%d.%d.%d.%d", b[0], b[1], b[2], b[3])
}

// Convert i to a hexadecimal string. Leading zeros are not printed.
func appendHex(dst []byte, i uint32) []byte {
	if i == 0 {
		return append(dst, '0')
	}

	const hexDigit = "0123456789abcdef"
	for j := 7; j >= 0; j-- {
		v := i >> uint(j*4)
		if v > 0 {
			dst = append(dst, hexDigit[v&0xf])
		}
	}
	return dst
}

// Bytes returns the byte representation of an IP address
func (ip IP) Bytes() []byte {
	if !ip.isLegacy {
		return ip.bytesIPv6()
	}

	return ip.bytesIPv4()
}

func (ip IP) To16BytesArray() [16]byte {
	ret := [16]byte{}

	for i, x := range ip.bytesIPv6() {
		ret[i] = x
	}

	return ret
}

func (ip *IP) bytesIPv4() []byte {
	u := ip.ToUint32()
	return []byte{
		byte(u & 0xFF000000 >> 24),
		byte(u & 0x00FF0000 >> 16),
		byte(u & 0x0000FF00 >> 8),
		byte(u & 0x000000FF),
	}
}

// IsIPv4 returns if the `IP` is of address family IPv4
func (ip IP) IsIPv4() bool {
	return ip.isLegacy
}

// SizeBytes returns the number of bytes required to represent the `IP`
func (ip *IP) SizeBytes() uint8 {
	if ip.isLegacy {
		return 4
	}

	return 16
}

// ToUint32 return the rightmost 32 bits of an 'IP'
func (ip *IP) ToUint32() uint32 {
	return uint32(^uint64(0) >> 32 & ip.lower)
}

func (ip *IP) bytesIPv6() []byte {
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

// ToNetIP converts the IP address in a `net.IP`
func (ip *IP) ToNetIP() net.IP {
	return net.IP(ip.Bytes())
}

// BitAtPosition returns the bit at position pos
func (ip IP) BitAtPosition(pos uint8) bool {
	if ip.isLegacy {
		return ip.bitAtPositionIPv4(pos)
	}

	return ip.bitAtPositionIPv6(pos)
}

func (ip IP) bitAtPositionIPv4(pos uint8) bool {
	if pos > 32 {
		return false
	}

	return (ip.ToUint32() & (1 << (32 - pos))) != 0
}

func (ip IP) bitAtPositionIPv6(pos uint8) bool {
	if pos > 128 {
		return false
	}

	if pos <= 64 {
		return (ip.higher & (1 << (64 - pos))) != 0
	}

	return (ip.lower & (1 << (128 - pos))) != 0
}

// Next gets the next ip address
func (ip *IP) Next() IP {
	newIP := ip.copy()
	if ip.isLegacy {
		newIP.lower++
		return newIP
	}

	newIP.lower++
	if newIP.lower == 0 {
		newIP.higher++
	}

	return newIP
}

// MaskLastNBits masks the last n bits of an IP address
func (ip *IP) MaskLastNBits(n uint8) IP {
	if ip.isLegacy {
		return ip.maskLastNBitsIPv4(n)
	}

	return ip.maskLastNBitsIPv6(n)
}

func (ip IP) maskLastNBitsIPv4(n uint8) IP {
	mask := uint64((math.MaxUint64 << (n)))
	ip.lower = ip.lower & mask

	return ip
}

func (ip IP) maskLastNBitsIPv6(n uint8) IP {
	maskBitsLow := uint8(bmath.Min(int(n), 64))
	maskBitsHigh := uint8(bmath.Max(int(n)-64, 0))

	maskLow := uint64((math.MaxUint64 << (maskBitsLow)))
	maskHigh := uint64((math.MaxUint64 << (maskBitsHigh)))

	ip.lower = ip.lower & maskLow
	ip.higher = ip.higher & maskHigh

	return ip
}
