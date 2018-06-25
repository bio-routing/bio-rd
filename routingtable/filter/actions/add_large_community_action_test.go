package actions

import (
	"testing"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/route"
	"github.com/stretchr/testify/assert"
)

func TestAddingLargeCommunities(t *testing.T) {
	tests := []struct {
		name        string
		current     []packet.LargeCommunity
		communities []packet.LargeCommunity
		expected    string
	}{
		{
			name: "add one to empty",
			communities: []packet.LargeCommunity{
				packet.LargeCommunity{
					GlobalAdministrator: 1,
					DataPart1:           2,
					DataPart2:           3,
				},
			},
			expected: "(1,2,3)",
		},
		{
			name: "add one to existing",
			current: []packet.LargeCommunity{
				packet.LargeCommunity{
					GlobalAdministrator: 5,
					DataPart1:           6,
					DataPart2:           7,
				},
			},
			communities: []packet.LargeCommunity{
				packet.LargeCommunity{
					GlobalAdministrator: 1,
					DataPart1:           2,
					DataPart2:           3,
				},
			},
			expected: "(5,6,7) (1,2,3)",
		},
		{
			name: "add two to existing",
			current: []packet.LargeCommunity{
				packet.LargeCommunity{
					GlobalAdministrator: 5,
					DataPart1:           6,
					DataPart2:           7,
				},
			},
			communities: []packet.LargeCommunity{
				packet.LargeCommunity{
					GlobalAdministrator: 1,
					DataPart1:           2,
					DataPart2:           3,
				},
				packet.LargeCommunity{
					GlobalAdministrator: 7,
					DataPart1:           8,
					DataPart2:           9,
				},
			},
			expected: "(5,6,7) (1,2,3) (7,8,9)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(te *testing.T) {
			p := &route.Path{
				BGPPath: &route.BGPPath{
					LargeCommunities: test.current,
				},
			}

			a := NewAddLargeCommunityAction(test.communities)
			modPath, _ := a.Do(net.Prefix{}, p)

			assert.Equal(te, test.expected, modPath.BGPPath.LargeCommunitiesString())
		})
	}
}
