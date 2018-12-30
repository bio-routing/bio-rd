package server

import (
	"testing"

	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/route"
	"github.com/stretchr/testify/assert"
)

func TestProcessAttributes(t *testing.T) {
	unknown3 := &packet.PathAttribute{
		Transitive: true,
		TypeCode:   100,
		Value:      []byte{1, 2, 3, 4},
		Next:       nil,
	}

	unknown2 := &packet.PathAttribute{
		Transitive: false,
		TypeCode:   150,
		Value:      []byte{20},
		Next:       unknown3,
	}

	unknown1 := &packet.PathAttribute{
		Transitive: true,
		TypeCode:   200,
		Value:      []byte{5, 6},
		Next:       unknown2,
	}

	asPath := &packet.PathAttribute{
		Transitive: true,
		TypeCode:   packet.ASPathAttr,
		Value: types.ASPath{
			types.ASPathSegment{
				Type: types.ASSequence,
				ASNs: []uint32{},
			},
		},
		Next: unknown1,
	}

	f := &fsmAddressFamily{}

	p := &route.Path{
		BGPPath: &route.BGPPath{},
	}
	f.processAttributes(asPath, p)

	expectedCodes := []uint8{200, 100}
	expectedValues := [][]byte{{5, 6}, {1, 2, 3, 4}}

	i := 0
	for _, attr := range p.BGPPath.UnknownAttributes {
		assert.Equal(t, true, attr.Transitive, "Transitive")
		assert.Equal(t, expectedCodes[i], attr.TypeCode, "Code")
		assert.Equal(t, expectedValues[i], attr.Value, "Value")
		i++
	}

	assert.Equal(t, 2, i, "Count")
}
