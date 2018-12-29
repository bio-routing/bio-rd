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
			{
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

	assert.Equal(t, "5907ed8960ccc14eed8f1a34a8eb3e6c82a8dd947d6cbf67eb58ca292f4588d5", p.ComputeHash())

	p.LocalPref = 150

	assert.NotEqual(t, "5907ed8960ccc14eed8f1a34a8eb3e6c82a8dd947d6cbf67eb58ca292f4588d5", p.ComputeHash())
}
