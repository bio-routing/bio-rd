package route

import (
	"testing"

	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/stretchr/testify/assert"

	bnet "github.com/bio-routing/bio-rd/net"
)

func TestComputeHash(t *testing.T) {
	p := &BGPPath{
		ASPath: types.ASPath{
			types.ASPathSegment{
				ASNs: []uint32{123, 456},
				Type: types.ASSequence,
			},
		},
		BGPIdentifier: 1,
		Communities: []uint32{
			123, 456,
		},
		EBGP: false,
		LargeCommunities: []types.LargeCommunity{
			types.LargeCommunity{
				DataPart1:           1,
				DataPart2:           2,
				GlobalAdministrator: 3,
			},
		},
		LocalPref:      100,
		MED:            1,
		NextHop:        bnet.IPv4(100),
		PathIdentifier: 5,
		Source:         bnet.IPv4(4),
	}

	assert.Equal(t, "98d68e69d993f8807c561cc7d63de759f7edc732887f88a7ebf42f61b9e54821", p.ComputeHash())

	p.LocalPref = 150

	assert.NotEqual(t, "98d68e69d993f8807c561cc7d63de759f7edc732887f88a7ebf42f61b9e54821", p.ComputeHash())
}
