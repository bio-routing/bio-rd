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
	return f.matchesPrefixListFilters(p) &&
		f.machtchesRouteFilters(p) &&
		f.machtchesCommunityFilters(pa) &&
		f.machtchesLageCommunityFilters(pa)
}

func (t *TermCondition) matchesPrefixListFilters(p net.Prefix) bool {
	if len(t.prefixLists) == 0 {
		return true
	}

	for _, l := range t.prefixLists {
		if l.Matches(p) {
			return true
		}
	}

	return false
}

func (t *TermCondition) machtchesRouteFilters(p net.Prefix) bool {
	if len(t.routeFilters) == 0 {
		return true
	}

	for _, l := range t.routeFilters {
		if l.Matches(p) {
			return true
		}
	}

	return false
}

func (t *TermCondition) machtchesCommunityFilters(pa *route.Path) bool {
	if len(t.communityFilters) == 0 {
		return true
	}

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

func (t *TermCondition) machtchesLageCommunityFilters(pa *route.Path) bool {
	if len(t.largeCommunityFilters) == 0 {
		return true
	}

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
