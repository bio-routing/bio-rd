package filter

import (
	"testing"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/stretchr/testify/assert"
)

func TestMatches(t *testing.T) {
	tests := []struct {
		name         string
		prefix       net.Prefix
		prefixLists  []*PrefixList
		routeFilters []*RouteFilter
		expected     bool
	}{
		{
			name:   "one prefix matches in prefix list, no route filters set",
			prefix: net.NewPfx(strAddr("127.0.0.1"), 8),
			prefixLists: []*PrefixList{
				NewPrefixList(net.NewPfx(strAddr("127.0.0.1"), 8)),
			},
			routeFilters: []*RouteFilter{},
			expected:     true,
		},
		{
			name:   "one prefix in prefix list and no match, no route filters set",
			prefix: net.NewPfx(strAddr("127.0.0.1"), 8),
			prefixLists: []*PrefixList{
				NewPrefixList(net.NewPfx(0, 32)),
			},
			routeFilters: []*RouteFilter{},
			expected:     false,
		},
		{
			name:   "one prefix of 2 matches in prefix list, no route filters set",
			prefix: net.NewPfx(strAddr("127.0.0.1"), 8),
			prefixLists: []*PrefixList{
				NewPrefixList(net.NewPfx(strAddr("10.0.0.0"), 8)),
				NewPrefixList(net.NewPfx(strAddr("127.0.0.1"), 8)),
			},
			routeFilters: []*RouteFilter{},
			expected:     true,
		},
		{
			name:        "no prefixes in prefix list, only route filter matches",
			prefix:      net.NewPfx(strAddr("10.0.0.0"), 24),
			prefixLists: []*PrefixList{},
			routeFilters: []*RouteFilter{
				NewRouteFilter(net.NewPfx(strAddr("10.0.0.0"), 8), Longer()),
			},
			expected: true,
		},
		{
			name:        "no prefixes in prefix list, one route filter matches",
			prefix:      net.NewPfx(strAddr("10.0.0.0"), 24),
			prefixLists: []*PrefixList{},
			routeFilters: []*RouteFilter{
				NewRouteFilter(net.NewPfx(strAddr("8.0.0.0"), 8), Longer()),
				NewRouteFilter(net.NewPfx(strAddr("10.0.0.0"), 8), Longer()),
			},
			expected: true,
		},
		{
			name:        "no prefixes in prefix list, one of many route filters matches",
			prefix:      net.NewPfx(strAddr("127.0.0.1"), 8),
			prefixLists: []*PrefixList{},
			routeFilters: []*RouteFilter{
				NewRouteFilter(net.NewPfx(strAddr("10.0.0.0"), 8), Longer()),
			},
			expected: false,
		},
		{
			name:   "no match in prefix list, no macht in route filter",
			prefix: net.NewPfx(strAddr("9.9.9.0"), 24),
			prefixLists: []*PrefixList{
				NewPrefixList(net.NewPfx(strAddr("8.0.0.0"), 8)),
			},
			routeFilters: []*RouteFilter{
				NewRouteFilter(net.NewPfx(strAddr("10.0.0.0"), 8), Longer()),
			},
			expected: false,
		},
		{
			name:   "one prefix in prefixlist, one route fitler, only prefix list matches",
			prefix: net.NewPfx(strAddr("8.8.8.0"), 24),
			prefixLists: []*PrefixList{
				NewPrefixList(net.NewPfx(strAddr("8.0.0.0"), 8)),
			},
			routeFilters: []*RouteFilter{
				NewRouteFilter(net.NewPfx(strAddr("10.0.0.0"), 8), Longer()),
			},
			expected: true,
		},
		{
			name:   "one prefix in prefixlist, one route fitler, only route filter matches",
			prefix: net.NewPfx(strAddr("10.0.0.0"), 24),
			prefixLists: []*PrefixList{
				NewPrefixList(net.NewPfx(strAddr("8.0.0.0"), 8)),
			},
			routeFilters: []*RouteFilter{
				NewRouteFilter(net.NewPfx(strAddr("10.0.0.0"), 8), Longer()),
			},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(te *testing.T) {
			f := NewTermCondition(
				test.prefixLists,
				test.routeFilters,
			)

			assert.Equal(te, test.expected, f.Matches(test.prefix, &route.Path{}))
		})
	}
}
