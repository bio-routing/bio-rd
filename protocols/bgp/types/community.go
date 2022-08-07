package types

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	// WellKnownCommunityNoExport is the well known no export BGP community (RFC1997)
	WellKnownCommunityNoExport = 0xFFFFFF01
	// WellKnownCommunityNoAdvertise is the well known no advertise BGP community (RFC1997)
	WellKnownCommunityNoAdvertise = 0xFFFFFF02
)

// CommunityStringForUint32 transforms a community into a human readable representation
func CommunityStringForUint32(v uint32) string {
	e1 := v >> 16
	e2 := v & 0x0000FFFF

	return fmt.Sprintf("(%d,%d)", e1, e2)
}

// ParseCommunityString parses human readable community representation
func ParseCommunityString(s string) (uint32, error) {
	s = strings.Trim(s, "()")
	t := strings.Split(s, ",")

	if len(t) != 2 {
		return 0, fmt.Errorf("can not parse community %s", s)
	}

	v, err := strconv.ParseUint(t[0], 10, 16)
	if err != nil {
		return 0, err
	}
	e1 := uint32(v)

	v, err = strconv.ParseUint(t[1], 10, 16)
	if err != nil {
		return 0, err
	}
	e2 := uint32(v)

	return e1<<16 + e2, nil
}

type Communities []uint32

func (c *Communities) String() string {
	if c == nil {
		return ""
	}

	cStrings := make([]string, len(*c))
	for i, x := range *c {
		cStrings[i] = strconv.Itoa(int(x))
	}

	return strings.Join(cStrings, " ")
}
