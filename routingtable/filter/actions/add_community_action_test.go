package actions

import (
	"testing"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/stretchr/testify/assert"
)

func TestAddingCommunities(t *testing.T) {
	tests := []struct {
		name        string
		current     []uint32
		communities []uint32
		expected    string
	}{
		{
			name: "add one to empty",
			communities: []uint32{
				65538,
			},
			expected: "(1,2)",
		},
		{
			name: "add one to existing",
			current: []uint32{
				65538,
			},
			communities: []uint32{
				196612,
			},
			expected: "(1,2) (3,4)",
		},
		{
			name: "add two to existing",
			current: []uint32{
				65538,
			},
			communities: []uint32{
				196612, 327686,
			},
			expected: "(1,2) (3,4) (5,6)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := &route.Path{
				BGPPath: &route.BGPPath{
					Communities: test.current,
				},
			}

			a := NewAddCommunityAction(test.communities)
			modPath, _ := a.Do(net.Prefix{}, p)

			assert.Equal(t, test.expected, modPath.BGPPath.CommunitiesString())
		})
	}
}
