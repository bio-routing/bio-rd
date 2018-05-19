package filter

import (
	"strings"

	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
)

type LargeCommunityFilter struct {
	community *packet.LargeCommunity
}

func (f *LargeCommunityFilter) Matches(communityString string) bool {
	return strings.Contains(communityString, f.community.String())
}
