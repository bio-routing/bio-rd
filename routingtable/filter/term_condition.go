package filter

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
)

type TermCondition struct {
	prefixLists           []*PrefixList
	routeFilters          []*RouteFilter
	communityFilters      []*CommunityFilter
	largeCommunityFilters []*LargeCommunityFilter
}

func NewTermCondition(prefixLists []*PrefixList, routeFilters []*RouteFilter) *TermCondition {
	return &TermCondition{
		prefixLists:  prefixLists,
		routeFilters: routeFilters,
	}
}

func (f *TermCondition) Matches(p net.Prefix, pa *route.Path) bool {
	return f.matchesAnyPrefixList(p) ||
		f.machtchesAnyRouteFilter(p) ||
		f.machtchesAnyLageCommunityFilter(pa) ||
		f.machtchesAnyCommunityFilter(pa)
}

func (t *TermCondition) matchesAnyPrefixList(p net.Prefix) bool {
	for _, l := range t.prefixLists {
		if l.Matches(p) {
			return true
		}
	}

	return false
}

func (t *TermCondition) machtchesAnyRouteFilter(p net.Prefix) bool {
	for _, l := range t.routeFilters {
		if l.Matches(p) {
			return true
		}
	}

	return false
}

func (t *TermCondition) machtchesAnyCommunityFilter(pa *route.Path) bool {
	if pa.BGPPath == nil {
		return false
	}

	for _, l := range t.communityFilters {
		if l.Matches(pa.BGPPath.Communities) {
			return true
		}
	}

	return false
}

func (t *TermCondition) machtchesAnyLageCommunityFilter(pa *route.Path) bool {
	if pa.BGPPath == nil {
		return false
	}

	for _, l := range t.largeCommunityFilters {
		if l.Matches(pa.BGPPath.LargeCommunities) {
			return true
		}
	}

	return false
}
