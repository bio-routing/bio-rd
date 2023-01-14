package actions

import (
	"testing"

	"github.com/bio-routing/bio-rd/route"
	"github.com/stretchr/testify/assert"

	bnet "github.com/bio-routing/bio-rd/net"
)

func TestSetNextHopTest(t *testing.T) {
	tests := []struct {
		name     string
		path     *route.Path
		expected *bnet.IP
	}{
		{
			name: "BGPPath is nil",
		},
		{
			name: "Modify BGP path",
			path: &route.Path{
				Type: route.BGPPathType,
				BGPPath: &route.BGPPath{
					BGPPathA: &route.BGPPathA{
						NextHop: bnet.IPv4FromOctets(192, 168, 1, 1).Ptr(),
					},
				},
			},
			expected: bnet.IPv4FromOctets(100, 64, 2, 1).Ptr(),
		},
		{
			name: "Modify Static path",
			path: &route.Path{
				Type: route.StaticPathType,
				StaticPath: &route.StaticPath{
					NextHop: bnet.IPv4FromOctets(192, 168, 1, 1).Ptr(),
				},
			},
			expected: bnet.IPv4FromOctets(100, 64, 2, 1).Ptr(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := NewSetNextHopAction(bnet.IPv4FromOctets(100, 64, 2, 1).Ptr())
			res := a.Do(bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), test.path)

			if test.path != nil {
				assert.Equal(t, test.expected, res.Path.GetNextHop())
			}
		})
	}
}
