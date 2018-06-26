package route

import (
	"testing"

	"github.com/bio-routing/bio-rd/protocols/bgp/packet"

	"github.com/stretchr/testify/assert"
)

func TestComputeHash(t *testing.T) {
	p := &BGPPath{
		ASPath: packet.ASPath{
			packet.ASPathSegment{
				ASNs:  []uint32{123, 456},
				Count: 2,
				Type:  packet.ASSequence,
			},
		},
		BGPIdentifier: 1,
		Communities: []uint32{
			123, 456,
		},
		EBGP: false,
		LargeCommunities: []packet.LargeCommunity{
			packet.LargeCommunity{
				DataPart1:           1,
				DataPart2:           2,
				GlobalAdministrator: 3,
			},
		},
		LocalPref:      100,
		MED:            1,
		NextHop:        100,
		PathIdentifier: 5,
		Source:         4,
	}

	assert.Equal(t, "313030093130300931323320343536093009310966616c736509310934095b313233203435365d095b7b33203120327d5d0935e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", p.ComputeHash())

	p.LocalPref = 150

	assert.NotEqual(t, "313030093130300931323320343536093009310966616c736509310934095b313233203435365d095b7b33203120327d5d0935e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", p.ComputeHash())
}
