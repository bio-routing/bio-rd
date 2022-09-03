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
	addr IP
	len  uint8
}

// Dedup gets a copy of Prefix from the cache.
// If Prefix is not in the cache it gets added.
func (p Prefix) Dedup() *Prefix {
	return pfxc.get(p)
}

// Ptr returns a pointer to p
func (p Prefix) Ptr() *Prefix {
	return &p
}

// NewPrefixFromProtoPrefix creates a Prefix from a proto Prefix
func NewPrefixFromProtoPrefix(pfx *api.Prefix) *Prefix {
	return &Prefix{
		addr: IPFromProtoIP(pfx.Address),
		len:  uint8(pfx.Length),
	}
}

// PrefixFromString converts prefix from string representation to Prefix
func PrefixFromString(s string) (*Prefix, error) {
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid format: %q", s)
	}

	ip, err := IPFromString(parts[0])
	if err != nil {
		return nil, err
	}

	l, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("unable to convert to int: %w", err)
	}

	return &Prefix{
		addr: ip,
		len:  uint8(l),
	}, nil
}

// ToProto converts prefix to proto prefix
func (p Prefix) ToProto() *api.Prefix {
	return &api.Prefix{
		Address: p.addr.ToProto(),
		Length:  uint32(p.len),
	}
}

// NewPfx creates a new Prefix
func NewPfx(addr IP, len uint8) Prefix {
	return Prefix{
		addr: addr,
		len:  len,
	}
}

// NewPfxFromIPNet creates a Prefix object from an gonet.IPNet object
func NewPfxFromIPNet(ipNet *gonet.IPNet) *Prefix {
	ones, _ := ipNet.Mask.Size()
	ip, _ := IPFromBytes(ipNet.IP)

	return &Prefix{
		addr: ip,
		len:  uint8(ones),
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
			return 0, fmt.Errorf("unable to convert %q to int: %w", parts[i], err)
		}

		if y > 255 {
			return 0, fmt.Errorf("%d is too big for a uint8", y)
		}

		ret += uint32(y) << uint((3-i)*8)
	}

	return ret, nil
}

// Addr returns the address of the prefix
func (pfx *Prefix) Addr() IP {
	return pfx.addr
}

// Pfxlen returns the length of the prefix
func (pfx *Prefix) Len() uint8 {
	return pfx.len
}

// String returns a string representation of pfx
func (pfx *Prefix) String() string {
	return fmt.Sprintf("%s/%d", pfx.addr.String(), pfx.len)
}

// GetIPNet returns the gonet.IP object for a Prefix object
func (pfx *Prefix) GetIPNet() *gonet.IPNet {
	var dstNetwork gonet.IPNet
	dstNetwork.IP = pfx.Addr().Bytes()

	pfxLen := int(pfx.Len())
	if pfx.Addr().IsIPv4() {
		dstNetwork.Mask = gonet.CIDRMask(pfxLen, 32)
	} else {
		dstNetwork.Mask = gonet.CIDRMask(pfxLen, 128)
	}

	return &dstNetwork
}

// Contains checks if x is a subnet of or equal to pfx
func (pfx *Prefix) Contains(x *Prefix) bool {
	if x.len <= pfx.len {
		return false
	}

	if pfx.addr.isLegacy {
		return pfx.containsIPv4(x)
	}

	return pfx.containsIPv6(x)
}

func (pfx *Prefix) containsIPv4(x *Prefix) bool {
	mask := uint32((math.MaxUint32 << (32 - pfx.len)))
	return (pfx.addr.ToUint32() & mask) == (x.addr.ToUint32() & mask)
}

func (pfx *Prefix) containsIPv6(x *Prefix) bool {
	var maskHigh, maskLow uint64
	if pfx.len <= 64 {
		maskHigh = math.MaxUint32 << (64 - pfx.len)
		maskLow = uint64(0)
	} else {
		maskHigh = math.MaxUint32
		maskLow = math.MaxUint32 << (128 - pfx.len)
	}

	return pfx.addr.higher&maskHigh&maskHigh == x.addr.higher&maskHigh&maskHigh &&
		pfx.addr.lower&maskHigh&maskLow == x.addr.lower&maskHigh&maskLow
}

// Equal checks if pfx and x are equal
func (pfx *Prefix) Equal(x *Prefix) bool {
	return pfx.addr.Equal(x.addr) && pfx.len == x.len
}

// GetSupernet gets the next common supernet of pfx and x
func (pfx *Prefix) GetSupernet(x *Prefix) Prefix {
	if pfx.addr.isLegacy {
		return pfx.supernetIPv4(x)
	}

	return pfx.supernetIPv6(x)
}

func (pfx *Prefix) supernetIPv4(x *Prefix) Prefix {
	maxPfxLen := min(pfx.len, x.len) - 1
	a := pfx.addr.ToUint32() >> (32 - maxPfxLen)
	b := x.addr.ToUint32() >> (32 - maxPfxLen)

	for i := 0; a != b; i++ {
		a = a >> 1
		b = b >> 1
		maxPfxLen--
	}

	return Prefix{
		addr: IPv4(a << (32 - maxPfxLen)),
		len:  maxPfxLen,
	}
}

func (pfx *Prefix) supernetIPv6(x *Prefix) Prefix {
	maxPfxLen := min(pfx.len, x.len)

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

// Valid checks if all bits outside of the prefix lengths range are zero (no host bit set)
func (p *Prefix) Valid() bool {
	if p.addr.isLegacy {
		return checkLastNBitsUint32(uint32(p.addr.lower), 32-p.len)
	}

	if p.len <= 64 {
		if p.addr.lower != 0 {
			return false
		}

		return checkLastNBitsUint64(p.addr.higher, 64-p.len)
	}

	return checkLastNBitsUint64(p.addr.lower, 64-(p.len-64))
}

func min(a uint8, b uint8) uint8 {
	if a < b {
		return a
	}
	return b
}

func checkLastNBitsUint32(x uint32, n uint8) bool {
	return x<<(32-n) == 0
}

func checkLastNBitsUint64(x uint64, n uint8) bool {
	return x<<(64-n) == 0
}

// BaseAddr gets the base address of the prefix
func (p *Prefix) BaseAddr() IP {
	if p.addr.isLegacy {
		return p.baseAddr4()
	}

	return p.baseAddr6()
}

func (p *Prefix) baseAddr4() IP {
	addr := p.addr.copy()

	addr.lower = addr.lower >> (32 - p.len)
	addr.lower = addr.lower << (32 - p.len)

	return addr
}

func (p *Prefix) baseAddr6() IP {
	addr := p.addr.copy()

	if p.len <= 64 {
		addr.lower = 0
		addr.higher = addr.higher >> (64 - p.len)
		addr.higher = addr.higher << (64 - p.len)
	} else {
		addr.lower = addr.lower >> (128 - p.len)
		addr.lower = addr.lower << (128 - p.len)
	}

	return addr
}
