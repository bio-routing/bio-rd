package packet

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var largeCommunityRegEx = regexp.MustCompile("^\\(?\\d+,\\d+,\\d+\\)?$")

type LargeCommunity struct {
	GlobalAdministrator uint32
	DataPart1           uint32
	DataPart2           uint32
}

func (c LargeCommunity) String() string {
	return fmt.Sprintf("(%d,%d,%d)", c.GlobalAdministrator, c.DataPart1, c.DataPart2)
}

func ParseLargeCommunityString(s string) (com LargeCommunity, err error) {
	if !largeCommunityRegEx.Match([]byte(s)) {
		return com, fmt.Errorf("malformed large community string %s", s)
	}
	s = strings.Trim(s, "()")
	t := strings.Split(s, ",")

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
