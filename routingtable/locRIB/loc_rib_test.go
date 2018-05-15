package locRIB

import (
	"testing"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"

	"github.com/stretchr/testify/assert"
)

type pfxPath struct {
	pfx  net.Prefix
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
				pfx:  net.NewPfx(1, 32),
				path: nil,
			},
			expected: false,
		},
		// Not equal path
		{
			in: []pfxPath{
				{
					pfx: net.NewPfx(1, 32),
					path: &route.Path{
						Type: route.StaticPathType,
						StaticPath: &route.StaticPath{
							NextHop: 2,
						},
					},
				},
			},
			check: pfxPath{
				pfx:  net.NewPfx(1, 32),
				path: nil,
			},
			expected: false,
		},
		// Equal
		{
			in: []pfxPath{
				{
					pfx: net.NewPfx(1, 32),
					path: &route.Path{
						Type: route.StaticPathType,
						StaticPath: &route.StaticPath{
							NextHop: 2,
						},
					},
				},
			},
			check: pfxPath{
				pfx: net.NewPfx(1, 32),
				path: &route.Path{
					Type: route.StaticPathType,
					StaticPath: &route.StaticPath{
						NextHop: 2,
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
