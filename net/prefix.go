package net

import (
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"

	"github.com/taktv6/tflow2/convert"
)

// Prefix represents an IPv4 prefix
type Prefix struct {
	addr   uint32
	pfxlen uint8
}

// NewPfx creates a new Prefix
func NewPfx(addr uint32, pfxlen uint8) Prefix {
	return Prefix{
		addr:   addr,
		pfxlen: pfxlen,
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
func (pfx Prefix) Addr() uint32 {
	return pfx.addr
}

// Pfxlen returns the length of the prefix
func (pfx Prefix) Pfxlen() uint8 {
	return pfx.pfxlen
}

// String returns a string representation of pfx
func (pfx Prefix) String() string {
	return fmt.Sprintf("%s/%d", net.IP(convert.Uint32Byte(pfx.addr)), pfx.pfxlen)
}

// Contains checks if x is a subnet of or equal to pfx
func (pfx Prefix) Contains(x Prefix) bool {
	if x.pfxlen <= pfx.pfxlen {
		return false
	}

	mask := uint32((math.MaxUint32 << (32 - pfx.pfxlen)))
	return (pfx.addr & mask) == (x.addr & mask)
}

// Equal checks if pfx and x are equal
func (pfx Prefix) Equal(x Prefix) bool {
	return pfx == x
}

// GetSupernet gets the next common supernet of pfx and x
func (pfx Prefix) GetSupernet(x Prefix) Prefix {
	maxPfxLen := min(pfx.pfxlen, x.pfxlen) - 1
	a := pfx.addr >> (32 - maxPfxLen)
	b := x.addr >> (32 - maxPfxLen)

	for i := 0; a != b; i++ {
		a = a >> 1
		b = b >> 1
		maxPfxLen--
	}

	return Prefix{
		addr:   a << (32 - maxPfxLen),
		pfxlen: maxPfxLen,
	}
}

func min(a uint8, b uint8) uint8 {
	if a < b {
		return a
	}
	return b
}
