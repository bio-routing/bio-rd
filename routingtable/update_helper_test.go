package routingtable

import (
	"strings"
	"testing"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/route"
	"github.com/stretchr/testify/assert"
)

func TestShouldPropagateUpdate(t *testing.T) {
	tests := []struct {
		name         string
		communities  string
		sessionAttrs SessionAttrs
		expected     bool
	}{
		{
			name:     "arbitrary path",
			expected: true,
		},
		{
			name:        "path was received from this peer before",
			communities: "(1,2)",
			sessionAttrs: SessionAttrs{
				Type:   route.BGPPathType,
				PeerIP: bnet.IPv4FromOctets(192, 168, 1, 1).Ptr(),
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
			sessionAttrs: SessionAttrs{
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
			sessionAttrs: SessionAttrs{
				IBGP: true,
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			comms := make(types.Communities, 0)
			for _, s := range strings.Split(test.communities, " ") {
				if s == "" {
					continue
				}

				com, err := types.ParseCommunityString(s)
				if err != nil {
					t.Fatalf("test failed: %s", err)
				}
				comms = append(comms, com)
			}

			pfx := bnet.NewPfx(bnet.IPv4(0), 32).Ptr()
			pa := &route.Path{
				Type: route.BGPPathType,
				BGPPath: &route.BGPPath{
					Communities: &comms,
					BGPPathA: &route.BGPPathA{
						Source: bnet.IPv4FromOctets(192, 168, 1, 1).Ptr(),
					},
				},
			}

			res := ShouldPropagateUpdate(pfx, pa, &test.sessionAttrs)
			assert.Equal(t, test.expected, res)
		})
	}
}
