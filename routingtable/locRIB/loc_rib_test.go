package locRIB

import (
	"testing"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/stretchr/testify/assert"
)

type pfxPath struct {
	pfx  bnet.Prefix
	path *route.Path
}

type containsPfxPathTestcase struct {
	in       []*pfxPath
	check    *pfxPath
	expected bool
}

func TestContainsPfxPath(t *testing.T) {
	testCases := []containsPfxPathTestcase{
		{
			in: []*pfxPath{},
			check: &pfxPath{
				pfx:  bnet.NewPfx(bnet.IPv4(1).Ptr(), 32),
				path: nil,
			},
			expected: false,
		},
		// Not equal path
		{
			in: []*pfxPath{
				{
					pfx: bnet.NewPfx(bnet.IPv4(1).Ptr(), 32),
					path: &route.Path{
						Type: route.StaticPathType,
						StaticPath: &route.StaticPath{
							NextHop: bnet.IPv4(2).Ptr(),
						},
					},
				},
			},
			check: &pfxPath{
				pfx:  bnet.NewPfx(bnet.IPv4(1).Ptr(), 32),
				path: nil,
			},
			expected: false,
		},
		// Equal
		{
			in: []*pfxPath{
				{
					pfx: bnet.NewPfx(bnet.IPv4(1).Ptr(), 32),
					path: &route.Path{
						Type: route.StaticPathType,
						StaticPath: &route.StaticPath{
							NextHop: bnet.IPv4(2).Ptr(),
						},
					},
				},
			},
			check: &pfxPath{
				pfx: bnet.NewPfx(bnet.IPv4(1).Ptr(), 32),
				path: &route.Path{
					Type: route.StaticPathType,
					StaticPath: &route.StaticPath{
						NextHop: bnet.IPv4(2).Ptr(),
					},
				},
			},
			expected: true,
		},
	}
	for i, tc := range testCases {
		rib := New("inet.0")
		for _, p := range tc.in {
			err := rib.AddPath(&p.pfx, p.path)
			assert.Nil(t, err, "could not fill rib in testcase %v", i)
		}
		contains := rib.ContainsPfxPath(&tc.check.pfx, tc.check.path)
		assert.Equal(t, tc.expected, contains, "mismatch in testcase %v", i)
	}
}

func TestLocRIB_RemovePathUnknown(t *testing.T) {
	rib := New("inet.0")
	assert.True(
		t,
		rib.RemovePath(
			bnet.NewPfx(bnet.IPv4(1).Ptr(), 32).Ptr(),
			&route.Path{
				Type: route.StaticPathType,
				StaticPath: &route.StaticPath{
					NextHop: bnet.IPv4(2).Ptr(),
				},
			}))
}
