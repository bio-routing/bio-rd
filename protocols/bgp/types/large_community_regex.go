package types

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bio-routing/bio-rd/route/api"
)

type LargeCommunitiesRegex []LargeCommunityRegex

func (lcr *LargeCommunitiesRegex) String() string {
	if lcr == nil {
		return ""
	}

	lcrStrings := make([]string, len(*lcr))
	for i, x := range *lcr {
		lcrStrings[i] = x.String()
	}

	return strings.Join(lcrStrings, " ")
}

type LargeCommunityRegex struct {
	GlobalAdministrator uint32
	DataPart1           string
	DataPart2           string
}

// ToProto converts LargeCommunityRegex to proto LargeCommunityRegex
func (c *LargeCommunityRegex) ToProto() *api.LargeCommunityRegex {
	return &api.LargeCommunityRegex{
		GlobalAdministrator: c.GlobalAdministrator,
		DataPart1:           c.DataPart1,
		DataPart2:           c.DataPart2,
	}
}

// LargeCommunityRegexFromProtoCommunity converts a proto LargeCommunityRegex to LargeCommunityRegex
func LargeCommunityRegexFromProtoCommunity(alcr *api.LargeCommunityRegex) LargeCommunityRegex {
	return LargeCommunityRegex{
		GlobalAdministrator: alcr.GlobalAdministrator,
		DataPart1:           alcr.DataPart1,
		DataPart2:           alcr.DataPart2,
	}
}

// String transitions a large community to it's human readable representation
func (c *LargeCommunityRegex) String() string {
	if c == nil {
		return ""
	}

	return fmt.Sprintf("(%d,%d,%d)", c.GlobalAdministrator, c.DataPart1, c.DataPart2)
}

// ParseLargeCommunityRegexString parses a human readable large community representation
func ParseLargeCommunityRegexString(s string) (com LargeCommunityRegex, err error) {
	s = strings.Trim(s, "()")
	t := strings.Split(s, ",")

	if len(t) != 3 {
		return com, fmt.Errorf("can not parse large community regex %s", s)
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
	com.DataPart1 = string(v)

	v, err = strconv.ParseUint(t[2], 10, 32)
	if err != nil {
		return com, err
	}
	com.DataPart2 = string(v)

	return com, err
}
