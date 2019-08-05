package filter

import "github.com/bio-routing/bio-rd/route"

type CommunityFilter struct {
	community uint32
}

func (f *CommunityFilter) Matches(coms *route.Communities) bool {
	for _, com := range *coms {
		if com == f.community {
			return true
		}
	}

	return false
}
