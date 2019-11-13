package filter

import (
	"github.com/bio-routing/bio-rd/net"
)

type RouteFilter struct {
	pattern *net.Prefix
	matcher PrefixMatcher
}

func NewRouteFilter(pattern *net.Prefix, matcher PrefixMatcher) *RouteFilter {
	return &RouteFilter{
		pattern: pattern,
		matcher: matcher,
	}
}

func (f *RouteFilter) Matches(prefix *net.Prefix) bool {
	return f.matcher.Match(f.pattern, prefix)
}

func (f *RouteFilter) equal(x *RouteFilter) bool {
	if f.pattern != x.pattern {
		return false
	}

	if !f.matcher.equal(x.matcher) {
		return false
	}

	return true
}
