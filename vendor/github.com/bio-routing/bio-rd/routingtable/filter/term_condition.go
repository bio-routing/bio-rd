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

func NewTermConditionWithRouteFilters(filters ...*RouteFilter) *TermCondition {
	return &TermCondition{
		routeFilters: filters,
	}
}

func NewTermConditionWithPrefixLists(filters ...*PrefixList) *TermCondition {
	return &TermCondition{
		prefixLists: filters,
	}
}

func (f *TermCondition) Matches(p *net.Prefix, pa *route.Path) bool {
	return f.matchesPrefixListFilters(p) &&
		f.matchesRouteFilters(p) &&
		f.matchesCommunityFilters(pa) &&
		f.matchesLargeCommunityFilters(pa)
}

func (t *TermCondition) matchesPrefixListFilters(p *net.Prefix) bool {
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

func (t *TermCondition) matchesRouteFilters(p *net.Prefix) bool {
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

func (t *TermCondition) matchesCommunityFilters(pa *route.Path) bool {
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

func (t *TermCondition) matchesLargeCommunityFilters(pa *route.Path) bool {
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

func (t *TermCondition) equal(x *TermCondition) bool {
	if len(t.routeFilters) != len(x.routeFilters) {
		return false
	}

	if len(t.communityFilters) != len(x.communityFilters) {
		return false
	}

	if len(t.largeCommunityFilters) != len(x.largeCommunityFilters) {
		return false
	}

	for i := range t.routeFilters {
		if !t.routeFilters[i].equal(x.routeFilters[i]) {
			return false
		}
	}

	// TODO: Compare community filters

	// TODO: Compare large community filters

	return true
}
