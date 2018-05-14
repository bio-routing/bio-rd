package filter

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
)

type TermCondition struct {
	prefixLists  []*PrefixList
	routeFilters []*RouteFilter
}

func (f *TermCondition) Matches(p net.Prefix, pa *route.Path) bool {
	return f.matchesAnyPrefixList(p) || f.machtchesAnyRouteFilter(p)
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
