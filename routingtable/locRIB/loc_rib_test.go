package locRIB

import (
	"testing"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"

	"github.com/stretchr/testify/assert"
)

type pfxPath struct {
	pfx  bnet.Prefix
	path *route.Path
}

type containsPfxPathTestcase struct {
	in       []pfxPath
	check    pfxPath
	expected bool
}

func TestContainsPfxPath(t *testing.T) {
	testCases := []containsPfxPathTestcase{
		{
			in: []pfxPath{},
			check: pfxPath{
				pfx:  bnet.NewPfx(bnet.IPv4(1), 32),
				path: nil,
			},
			expected: false,
		},
		// Not equal path
		{
			in: []pfxPath{
				{
					pfx: bnet.NewPfx(bnet.IPv4(1), 32),
					path: &route.Path{
						Type: route.StaticPathType,
						StaticPath: &route.StaticPath{
							NextHop: bnet.IPv4(2),
						},
					},
				},
			},
			check: pfxPath{
				pfx:  bnet.NewPfx(bnet.IPv4(1), 32),
				path: nil,
			},
			expected: false,
		},
		// Equal
		{
			in: []pfxPath{
				{
					pfx: bnet.NewPfx(bnet.IPv4(1), 32),
					path: &route.Path{
						Type: route.StaticPathType,
						StaticPath: &route.StaticPath{
							NextHop: bnet.IPv4(2),
						},
					},
				},
			},
			check: pfxPath{
				pfx: bnet.NewPfx(bnet.IPv4(1), 32),
				path: &route.Path{
					Type: route.StaticPathType,
					StaticPath: &route.StaticPath{
						NextHop: bnet.IPv4(2),
					},
				},
			},
			expected: true,
		},
	}
	for i, tc := range testCases {
		rib := New()
		for _, p := range tc.in {
			err := rib.AddPath(p.pfx, p.path)
			assert.Nil(t, err, "could not fill rib in testcase %v", i)
		}
		contains := rib.ContainsPfxPath(tc.check.pfx, tc.check.path)
		assert.Equal(t, tc.expected, contains, "mismatch in testcase %v", i)
	}
}

func TestLocRIB_RemovePathUnknown(t *testing.T) {
	rib := New()
	assert.True(t, rib.RemovePath(bnet.NewPfx(bnet.IPv4(1), 32),
		&route.Path{
			Type: route.StaticPathType,
			StaticPath: &route.StaticPath{
				NextHop: bnet.IPv4(2),
			},
		}))
}

func TestPropagation(t *testing.T) {
	rib := New()
	err := rib.AddPath(bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
		Type:    route.BGPPathType,
		BGPPath: &route.BGPPath{},
	})
	if err != nil {
		t.Errorf("Unexpected failure: %v", err)
	}

	c := routingtable.NewRTMockClient()
	rib.RegisterWithOptions(c, routingtable.ClientOptions{
		MaxPaths: 10,
	})

}
