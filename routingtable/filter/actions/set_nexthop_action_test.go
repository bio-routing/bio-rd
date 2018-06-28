package actions

import (
	"testing"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/stretchr/testify/assert"

	bnet "github.com/bio-routing/bio-rd/net"
)

func TestSetNextHopTest(t *testing.T) {
	tests := []struct {
		name     string
		bgpPath  *route.BGPPath
		expected net.IP
	}{
		{
			name: "BGPPath is nil",
		},
		{
			name: "modify path",
			bgpPath: &route.BGPPath{
				NextHop: bnet.IPv4FromOctets(192, 168, 1, 1),
			},
			expected: bnet.IPv4FromOctets(100, 64, 2, 1),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := NewSetNextHopAction(bnet.IPv4FromOctets(100, 64, 2, 1))
			p, _ := a.Do(net.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
				BGPPath: test.bgpPath,
			})

			if test.bgpPath != nil {
				assert.Equal(t, test.expected, p.BGPPath.NextHop)
			}
		})
	}
}
