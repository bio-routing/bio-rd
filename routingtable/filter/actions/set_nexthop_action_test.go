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
		bgpPath  *route.BGPPath
		expected *bnet.IP
	}{
		{
			name: "BGPPath is nil",
		},
		{
			name: "modify path",
			bgpPath: &route.BGPPath{
				BGPPathA: &route.BGPPathA{
					NextHop: bnet.IPv4FromOctets(192, 168, 1, 1).Ptr(),
				},
			},
			expected: bnet.IPv4FromOctets(100, 64, 2, 1).Ptr(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := NewSetNextHopAction(bnet.IPv4FromOctets(100, 64, 2, 1).Ptr())
			res := a.Do(bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
				BGPPath: test.bgpPath,
			})

			if test.bgpPath != nil {
				assert.Equal(t, test.expected, res.Path.BGPPath.BGPPathA.NextHop)
			}
		})
	}
}
