package filter

import (
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
)

type LargeCommunityFilter struct {
	community packet.LargeCommunity
}

func (f *LargeCommunityFilter) Matches(coms []packet.LargeCommunity) bool {
	for _, com := range coms {
		if com == f.community {
			return true
		}
	}

	return false
}
