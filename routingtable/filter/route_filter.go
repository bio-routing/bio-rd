package filter

import (
	"github.com/bio-routing/bio-rd/net"
)

type RouteFilter struct {
	pattern net.Prefix
	matcher PrefixMatcher
}

func NewRouteFilter(pattern net.Prefix, matcher PrefixMatcher) *RouteFilter {
	return &RouteFilter{
		pattern: pattern,
		matcher: matcher,
	}
}

func (f *RouteFilter) Matches(prefix net.Prefix) bool {
	return f.matcher(f.pattern, prefix)
}
