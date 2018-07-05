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

	assert.Equal(t, "1058916ff3e6a51c7d8a47945d13fc3fcd8ee578a6d376505f46d58979b30fae", p.ComputeHash())

	p.LocalPref = 150

	assert.NotEqual(t, "1058916ff3e6a51c7d8a47945d13fc3fcd8ee578a6d376505f46d58979b30fae", p.ComputeHash())
}
