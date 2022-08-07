package types

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bio-routing/bio-rd/route/api"
)

type LargeCommunities []LargeCommunity

func (lc *LargeCommunities) String() string {
	if lc == nil {
		return ""
	}

	lcStrings := make([]string, len(*lc))
	for i, x := range *lc {
		lcStrings[i] = x.String()
	}

	return strings.Join(lcStrings, " ")
}

// LargeCommunity represents a large community (RFC8195)
type LargeCommunity struct {
	GlobalAdministrator uint32
	DataPart1           uint32
	DataPart2           uint32
}

// ToProto converts LargeCommunity to proto LargeCommunity
func (c *LargeCommunity) ToProto() *api.LargeCommunity {
	return &api.LargeCommunity{
		GlobalAdministrator: c.GlobalAdministrator,
		DataPart1:           c.DataPart1,
		DataPart2:           c.DataPart2,
	}
}

// LargeCommunityFromProtoCommunity converts a proto LargeCommunity to LargeCommunity
func LargeCommunityFromProtoCommunity(alc *api.LargeCommunity) LargeCommunity {
	return LargeCommunity{
		GlobalAdministrator: alc.GlobalAdministrator,
		DataPart1:           alc.DataPart1,
		DataPart2:           alc.DataPart2,
	}
}

// String transitions a large community to it's human readable representation
func (c *LargeCommunity) String() string {
	if c == nil {
		return ""
	}

	return fmt.Sprintf("(%d,%d,%d)", c.GlobalAdministrator, c.DataPart1, c.DataPart2)
}

// ParseLargeCommunityString parses a human readable large community representation
func ParseLargeCommunityString(s string) (com LargeCommunity, err error) {
	s = strings.Trim(s, "()")
	t := strings.Split(s, ",")

	if len(t) != 3 {
		return com, fmt.Errorf("can not parse large community %s", s)
	}

	v, err := strconv.ParseUint(t[0], 10, 32)
	if err != nil {
		return com, err
	}
	com.GlobalAdministrator = uint32(v)

	v, err = strconv.ParseUint(t[1], 10, 32)
	if err != nil {
		return com, err
	}
	com.DataPart1 = uint32(v)

	v, err = strconv.ParseUint(t[2], 10, 32)
	if err != nil {
		return com, err
	}
	com.DataPart2 = uint32(v)

	return com, err
}
