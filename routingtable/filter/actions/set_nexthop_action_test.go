package actions

import (
	"testing"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/stretchr/testify/assert"
)

func TestSetNextHopTest(t *testing.T) {
	tests := []struct {
		name     string
		bgpPath  *route.BGPPath
		expected uint32
	}{
		{
			name: "BGPPath is nil",
		},
		{
			name: "modify path",
			bgpPath: &route.BGPPath{
				NextHop: strAddr("100.64.2.1"),
			},
			expected: strAddr("100.64.2.1"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(te *testing.T) {
			a := NewSetNextHopAction(strAddr("100.64.2.1"))
			p, _ := a.Do(net.NewPfx(strAddr("10.0.0.0"), 8), &route.Path{
				BGPPath: test.bgpPath,
			})

			if test.expected > 0 {
				assert.Equal(te, test.expected, p.BGPPath.NextHop)
			}
		})
	}
}
