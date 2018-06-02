package routingtable

import (
	"net"
	"testing"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/stretchr/testify/assert"
)

func TestShouldPropagateUpdate(t *testing.T) {
	tests := []struct {
		name        string
		communities string
		neighbor    Neighbor
		expected    bool
	}{
		{
			name:     "arbitrary path",
			expected: true,
		},
		{
			name:        "path was received from this peer before",
			communities: "(1,2)",
			neighbor: Neighbor{
				Type:    route.BGPPathType,
				Address: bnet.IPv4ToUint32(net.ParseIP("192.168.1.1")),
			},
			expected: false,
		},
		{
			name:        "path with no-export community",
			communities: "(1,2) (65535,65281)",
			expected:    false,
		},
		{
			name:        "path with no-export community (iBGP)",
			communities: "(1,2) (65535,65281)",
			neighbor: Neighbor{
				IBGP: true,
			},
			expected: true,
		},
		{
			name:        "path with no-advertise community",
			communities: "(1,2) (65535,65282)",
			expected:    false,
		},
		{
			name:        "path with no-advertise community (iBGP)",
			communities: "(1,2) (65535,65282)",
			neighbor: Neighbor{
				IBGP: true,
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(te *testing.T) {
			pfx := bnet.NewPfx(0, 32)

			pa := &route.Path{
				Type: route.BGPPathType,
				BGPPath: &route.BGPPath{
					Communities: test.communities,
					Source:      bnet.IPv4ToUint32(net.ParseIP("192.168.1.1")),
				},
			}

			res := ShouldPropagateUpdate(pfx, pa, &test.neighbor)
			assert.Equal(te, test.expected, res)
		})
	}
}
