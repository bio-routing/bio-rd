package filter

import (
	"testing"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/route"
	"github.com/stretchr/testify/assert"
)

func TestMatches(t *testing.T) {
	tests := []struct {
		name                  string
		prefix                net.Prefix
		bgpPath               *route.BGPPath
		prefixLists           []*PrefixList
		routeFilters          []*RouteFilter
		communityFilters      []*CommunityFilter
		largeCommunityFilters []*LargeCommunityFilter
		expected              bool
	}{
		{
			name:   "one prefix matches in prefix list, no route filters set",
			prefix: net.NewPfx(net.IPv4FromOctets(127, 0, 0, 1), 8),
			prefixLists: []*PrefixList{
				NewPrefixList(net.NewPfx(net.IPv4FromOctets(127, 0, 0, 1), 8)),
			},
			expected: true,
		},
		{
			name:   "one prefix in prefix list and no match, no route filters set",
			prefix: net.NewPfx(net.IPv4FromOctets(127, 0, 0, 1), 8),
			prefixLists: []*PrefixList{
				NewPrefixList(net.NewPfx(net.IPv4(0), 32)),
			},
			expected: false,
		},
		{
			name:   "one prefix of 2 matches in prefix list, no route filters set",
			prefix: net.NewPfx(net.IPv4FromOctets(127, 0, 0, 1), 8),
			prefixLists: []*PrefixList{
				NewPrefixList(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8)),
				NewPrefixList(net.NewPfx(net.IPv4FromOctets(127, 0, 0, 1), 8)),
			},
			expected: true,
		},
		{
			name:   "no prefixes in prefix list, only route filter matches",
			prefix: net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 24),
			routeFilters: []*RouteFilter{
				NewRouteFilter(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), NewLongerMatcher()),
			},
			expected: true,
		},
		{
			name:   "no prefixes in prefix list, one route filter matches",
			prefix: net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 24),
			routeFilters: []*RouteFilter{
				NewRouteFilter(net.NewPfx(net.IPv4FromOctets(8, 0, 0, 0), 8), NewLongerMatcher()),
				NewRouteFilter(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), NewLongerMatcher()),
			},
			expected: true,
		},
		{
			name:   "no prefixes in prefix list, one of many route filters matches",
			prefix: net.NewPfx(net.IPv4FromOctets(127, 0, 0, 1), 8),
			routeFilters: []*RouteFilter{
				NewRouteFilter(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), NewLongerMatcher()),
			},
			expected: false,
		},
		{
			name:   "no match in prefix list, no macht in route filter",
			prefix: net.NewPfx(net.IPv4FromOctets(9, 9, 9, 0), 24),
			prefixLists: []*PrefixList{
				NewPrefixList(net.NewPfx(net.IPv4FromOctets(8, 0, 0, 0), 8)),
			},
			routeFilters: []*RouteFilter{
				NewRouteFilter(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), NewLongerMatcher()),
			},
			expected: false,
		},
		{
			name:   "one prefix in prefixlist, one route filter, only prefix list matches",
			prefix: net.NewPfx(net.IPv4FromOctets(8, 8, 8, 0), 24),
			prefixLists: []*PrefixList{
				NewPrefixList(net.NewPfx(net.IPv4FromOctets(8, 0, 0, 0), 8)),
			},
			routeFilters: []*RouteFilter{
				NewRouteFilter(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), NewLongerMatcher()),
			},
			expected: false,
		},
		{
			name:   "one prefix in prefixlist, one route filter, only route filter matches",
			prefix: net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 24),
			prefixLists: []*PrefixList{
				NewPrefixList(net.NewPfx(net.IPv4FromOctets(8, 0, 0, 0), 8)),
			},
			routeFilters: []*RouteFilter{
				NewRouteFilter(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), NewLongerMatcher()),
			},
			expected: false,
		},
		{
			name:   "community matches",
			prefix: net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 24),
			bgpPath: &route.BGPPath{
				Communities: []uint32{65538, 196612, 327686}, // (1,2) (3,4) (5,6)
			},
			communityFilters: []*CommunityFilter{
				{196612}, // (3,4)
			},
			expected: true,
		},
		{
			name:   "community does not match",
			prefix: net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 24),
			bgpPath: &route.BGPPath{
				Communities: []uint32{65538, 196612, 327686}, // (1,2) (3,4) (5,6)
			},
			communityFilters: []*CommunityFilter{
				{196608}, // (3,0)
			},
			expected: false,
		},
		{
			name:   "community filter, bgp path is nil",
			prefix: net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 24),
			communityFilters: []*CommunityFilter{
				{196608}, // (3,0)
			},
			expected: false,
		},
		{
			name:   "large community matches",
			prefix: net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 24),
			bgpPath: &route.BGPPath{
				LargeCommunities: []types.LargeCommunity{
					{
						GlobalAdministrator: 1,
						DataPart1:           2,
						DataPart2:           3,
					},
					{
						GlobalAdministrator: 1,
						DataPart1:           2,
						DataPart2:           0,
					},
				},
			},
			largeCommunityFilters: []*LargeCommunityFilter{
				{
					types.LargeCommunity{
						GlobalAdministrator: 1,
						DataPart1:           2,
						DataPart2:           3,
					},
				},
			},
			expected: true,
		},
		{
			name:    "large community does not match",
			prefix:  net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 24),
			bgpPath: &route.BGPPath{},
			largeCommunityFilters: []*LargeCommunityFilter{
				{
					types.LargeCommunity{
						GlobalAdministrator: 1,
						DataPart1:           2,
						DataPart2:           3,
					},
				},
			},
			expected: false,
		},
		{
			name:   "large community filter, bgp path is nil",
			prefix: net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 24),
			largeCommunityFilters: []*LargeCommunityFilter{
				{
					types.LargeCommunity{
						GlobalAdministrator: 1,
						DataPart1:           2,
						DataPart2:           3,
					},
				},
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(te *testing.T) {
			f := NewTermCondition(test.prefixLists, test.routeFilters)
			f.communityFilters = test.communityFilters
			f.largeCommunityFilters = test.largeCommunityFilters

			pa := &route.Path{
				BGPPath: test.bgpPath,
			}

			assert.Equal(te, test.expected, f.Matches(test.prefix, pa))
		})
	}
}
