package actions

import (
	"testing"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/stretchr/testify/assert"
)

func TestAppendPath(t *testing.T) {
	tests := []struct {
		name           string
		times          uint16
		bgpPath        *route.BGPPath
		expectedPath   string
		expectedLength uint16
	}{
		{
			name: "BGPPath is nil",
		},
		{
			name:  "append 0",
			times: 0,
			bgpPath: &route.BGPPath{
				ASPath:    "12345 12345",
				ASPathLen: 2,
			},
			expectedPath:   "12345 12345",
			expectedLength: 2,
		},
		{
			name:  "append 3",
			times: 3,
			bgpPath: &route.BGPPath{
				ASPath:    "12345 12345",
				ASPathLen: 2,
			},
			expectedPath:   "12345 12345 12345 12345 12345",
			expectedLength: 5,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(te *testing.T) {
			a := NewASPathPrependAction(12345, test.times)
			p, _ := a.Do(net.NewPfx(strAddr("10.0.0.0"), 8), &route.Path{
				BGPPath: test.bgpPath,
			})

			if test.bgpPath == nil {
				return
			}

			assert.Equal(te, test.expectedPath, p.BGPPath.ASPath, "ASPath")
			assert.Equal(te, test.expectedLength, p.BGPPath.ASPathLen, "ASPathLen")
		})
	}
}
