package net

import (
	"fmt"
	"math"
	gonet "net"
	"strconv"
	"strings"

	"github.com/bio-routing/bio-rd/net/api"
)

// Prefix represents an IPv4 prefix
type Prefix struct {
	addr   IP
	pfxlen uint8
}

// NewPrefixFromProtoPrefix creates a Prefix from a proto Prefix
func NewPrefixFromProtoPrefix(pfx api.Prefix) Prefix {
	return Prefix{
		addr:   IPFromProtoIP(*pfx.Address),
		pfxlen: uint8(pfx.Pfxlen),
	}
}

// ToProto converts prefix to proto prefix
func (pfx Prefix) ToProto() api.Prefix {
	return api.Prefix{
		Address: pfx.addr.ToProto(),
		Pfxlen:  uint32(pfx.pfxlen),
	}
}

// NewPfx creates a new Prefix
func NewPfx(addr IP, pfxlen uint8) Prefix {
	return Prefix{
		addr:   addr,
		pfxlen: pfxlen,
	}
}

// NewPfxFromIPNet creates a Prefix object from an gonet.IPNet object
func NewPfxFromIPNet(ipNet *gonet.IPNet) Prefix {
	ones, _ := ipNet.Mask.Size()
	ip, _ := IPFromBytes(ipNet.IP)

	return Prefix{
		addr:   ip,
		pfxlen: uint8(ones),
	}
}

// StrToAddr converts an IP address string to it's uint32 representation
func StrToAddr(x string) (uint32, error) {
	parts := strings.Split(x, ".")
	if len(parts) != 4 {
		return 0, fmt.Errorf("Invalid format")
	}

	ret := uint32(0)
	for i := 0; i < 4; i++ {
		y, err := strconv.Atoi(parts[i])
		if err != nil {
			return 0, fmt.Errorf("Unable to convert %q to int: %v", parts[i], err)
		}

		if y > 255 {
			return 0, fmt.Errorf("%d is too big for a uint8", y)
		}

		ret += uint32(math.Pow(256, float64(3-i))) * uint32(y)
	}

	return ret, nil
}

// Addr returns the address of the prefix
func (pfx Prefix) Addr() IP {
	return pfx.addr
}

// Pfxlen returns the length of the prefix
func (pfx Prefix) Pfxlen() uint8 {
	return pfx.pfxlen
}

// String returns a string representation of pfx
func (pfx Prefix) String() string {
	return fmt.Sprintf("%s/%d", pfx.addr, pfx.pfxlen)
}

// GetIPNet returns the gonet.IP object for a Prefix object
func (pfx Prefix) GetIPNet() *gonet.IPNet {
	var dstNetwork gonet.IPNet
	dstNetwork.IP = pfx.Addr().Bytes()

	pfxLen := int(pfx.Pfxlen())
	if pfx.Addr().IsIPv4() {
		dstNetwork.Mask = gonet.CIDRMask(pfxLen, 32)
	} else {
		dstNetwork.Mask = gonet.CIDRMask(pfxLen, 128)
	}

	return &dstNetwork
}

// Contains checks if x is a subnet of or equal to pfx
func (pfx Prefix) Contains(x Prefix) bool {
	if x.pfxlen <= pfx.pfxlen {
		return false
	}

	if pfx.addr.isLegacy {
		return pfx.containsIPv4(x)
	}

	return pfx.containsIPv6(x)
}

func (pfx Prefix) containsIPv4(x Prefix) bool {
	mask := uint32((math.MaxUint32 << (32 - pfx.pfxlen)))
	return (pfx.addr.ToUint32() & mask) == (x.addr.ToUint32() & mask)
}

func (pfx Prefix) containsIPv6(x Prefix) bool {
	var maskHigh, maskLow uint64
	if pfx.pfxlen <= 64 {
		maskHigh = math.MaxUint32 << (64 - pfx.pfxlen)
		maskLow = uint64(0)
	} else {
		maskHigh = math.MaxUint32
		maskLow = math.MaxUint32 << (128 - pfx.pfxlen)
	}

	return pfx.addr.higher&maskHigh&maskHigh == x.addr.higher&maskHigh&maskHigh &&
		pfx.addr.lower&maskHigh&maskLow == x.addr.lower&maskHigh&maskLow
}

// Equal checks if pfx and x are equal
func (pfx Prefix) Equal(x Prefix) bool {
	return pfx.addr.Equal(x.addr) && pfx.pfxlen == x.pfxlen
}

// GetSupernet gets the next common supernet of pfx and x
func (pfx Prefix) GetSupernet(x Prefix) Prefix {
	if pfx.addr.isLegacy {
		return pfx.supernetIPv4(x)
	}

	return pfx.supernetIPv6(x)
}

func (pfx Prefix) supernetIPv4(x Prefix) Prefix {
	maxPfxLen := min(pfx.pfxlen, x.pfxlen) - 1
	a := pfx.addr.ToUint32() >> (32 - maxPfxLen)
	b := x.addr.ToUint32() >> (32 - maxPfxLen)

	for i := 0; a != b; i++ {
		a = a >> 1
		b = b >> 1
		maxPfxLen--
	}

	return Prefix{
		addr:   IPv4(a << (32 - maxPfxLen)),
		pfxlen: maxPfxLen,
	}
}

func (pfx Prefix) supernetIPv6(x Prefix) Prefix {
	maxPfxLen := min(pfx.pfxlen, x.pfxlen)

	a := pfx.addr.BitAtPosition(1)
	b := x.addr.BitAtPosition(1)
	pfxLen := uint8(0)
	mask := uint64(0)
	for a == b && pfxLen < maxPfxLen {
		a = pfx.addr.BitAtPosition(pfxLen + 2)
		b = x.addr.BitAtPosition(pfxLen + 2)
		pfxLen++

		if pfxLen == 64 {
			mask = 0
		}

		m := pfxLen % 64
		mask = mask + uint64(1)<<(64-m)
	}

	if pfxLen == 0 {
		return NewPfx(IPv6(0, 0), pfxLen)
	}

	if pfxLen > 64 {
		return NewPfx(IPv6(pfx.addr.higher, pfx.addr.lower&mask), pfxLen)
	}

	return NewPfx(IPv6(pfx.addr.higher&mask, 0), pfxLen)
}

func min(a uint8, b uint8) uint8 {
	if a < b {
		return a
	}
	return b
}
