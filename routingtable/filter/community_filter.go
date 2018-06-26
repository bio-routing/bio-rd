package filter

type CommunityFilter struct {
	community uint32
}

func (f *CommunityFilter) Matches(coms []uint32) bool {
	for _, com := range coms {
		if com == f.community {
			return true
		}
	}

	return false
}
