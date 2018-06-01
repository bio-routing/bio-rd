package filter

import (
	"strings"

	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
)

type CommunityFilter struct {
	community uint32
}

func (f *CommunityFilter) Matches(communityString string) bool {
	return strings.Contains(communityString, packet.CommunityStringForUint32(f.community))
}
