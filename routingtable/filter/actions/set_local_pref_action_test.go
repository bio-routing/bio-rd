package actions

import (
	"testing"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/stretchr/testify/assert"
)

func TestSetLocalPref(t *testing.T) {
	tests := []struct {
		name              string
		bgpPath           *route.BGPPath
		expectedLocalPref uint32
	}{
		{
			name:              "BGPPath is nil",
			expectedLocalPref: 0,
		},
		{
			name: "modify path",
			bgpPath: &route.BGPPath{
				LocalPref: 100,
			},
			expectedLocalPref: 150,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(te *testing.T) {
			a := NewSetLocalPrefAction(150)
			p, _ := a.Do(net.NewPfx(strAddr("10.0.0.0"), 8), &route.Path{
				BGPPath: test.bgpPath,
			})

			if test.expectedLocalPref > 0 {
				assert.Equal(te, test.expectedLocalPref, p.BGPPath.LocalPref)
			}
		})
	}
}

func strAddr(s string) uint32 {
	ret, _ := net.StrToAddr(s)
	return ret
}
