package filter

import (
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
)

// LargeCommunityFilter represents a filter for large communities
type LargeCommunityFilter struct {
	community types.LargeCommunity
}

// Matches checks if a community f.community is on the filter list
func (f *LargeCommunityFilter) Matches(coms *types.LargeCommunities) bool {
	if coms == nil {
		return false
	}

	for _, com := range *coms {
		if com == f.community {
			return true
		}
	}

	return false
}
