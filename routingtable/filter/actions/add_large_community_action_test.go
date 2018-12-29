package actions

import (
	"testing"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/route"
	"github.com/stretchr/testify/assert"
)

func TestAddingLargeCommunities(t *testing.T) {
	tests := []struct {
		name        string
		current     []types.LargeCommunity
		communities []types.LargeCommunity
		expected    string
	}{
		{
			name: "add one to empty",
			communities: []types.LargeCommunity{
				{
					GlobalAdministrator: 1,
					DataPart1:           2,
					DataPart2:           3,
				},
			},
			expected: "(1,2,3)",
		},
		{
			name: "add one to existing",
			current: []types.LargeCommunity{
				{
					GlobalAdministrator: 5,
					DataPart1:           6,
					DataPart2:           7,
				},
			},
			communities: []types.LargeCommunity{
				{
					GlobalAdministrator: 1,
					DataPart1:           2,
					DataPart2:           3,
				},
			},
			expected: "(5,6,7) (1,2,3)",
		},
		{
			name: "add two to existing",
			current: []types.LargeCommunity{
				{
					GlobalAdministrator: 5,
					DataPart1:           6,
					DataPart2:           7,
				},
			},
			communities: []types.LargeCommunity{
				{
					GlobalAdministrator: 1,
					DataPart1:           2,
					DataPart2:           3,
				},
				{
					GlobalAdministrator: 7,
					DataPart1:           8,
					DataPart2:           9,
				},
			},
			expected: "(5,6,7) (1,2,3) (7,8,9)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := &route.Path{
				BGPPath: &route.BGPPath{
					LargeCommunities: test.current,
				},
			}

			a := NewAddLargeCommunityAction(test.communities)
			res := a.Do(net.Prefix{}, p)

			assert.Equal(t, test.expected, res.Path.BGPPath.LargeCommunitiesString())
		})
	}
}
