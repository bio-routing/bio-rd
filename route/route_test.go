package route

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/route/api"
)

func TestPathSelection(t *testing.T) {
	tests := []struct {
		name     string
		r        *Route
		expected []*Path
	}{
		{
			name: "Test Localpref",
			r: &Route{
				paths: []*Path{
					{
						Type: BGPPathType,
						BGPPath: &BGPPath{
							BGPPathA: &BGPPathA{
								LocalPref: 1000,
							},
						},
					},
					{
						Type: BGPPathType,
						BGPPath: &BGPPath{
							BGPPathA: &BGPPathA{
								LocalPref: 100,
							},
						},
					},
				},
			},
			expected: []*Path{
				{
					Type: BGPPathType,
					BGPPath: &BGPPath{
						BGPPathA: &BGPPathA{
							LocalPref: 1000,
						},
					},
				},
				{
					Type: BGPPathType,
					BGPPath: &BGPPath{
						BGPPathA: &BGPPathA{
							LocalPref: 100,
						},
					},
				},
			},
		},
		{
			name: "Test ASPath",
			r: &Route{
				paths: []*Path{
					{
						Type: BGPPathType,
						BGPPath: &BGPPath{
							BGPPathA: &BGPPathA{
								LocalPref: 1000,
							},
							ASPathLen: 3,
						},
					},
					{
						Type: BGPPathType,
						BGPPath: &BGPPath{
							BGPPathA: &BGPPathA{
								LocalPref: 1000,
							},
							ASPathLen: 1,
						},
					},
				},
			},
			expected: []*Path{
				{
					Type: BGPPathType,
					BGPPath: &BGPPath{
						BGPPathA: &BGPPathA{
							LocalPref: 1000,
						},
						ASPathLen: 1,
					},
				},
				{
					Type: BGPPathType,
					BGPPath: &BGPPath{
						BGPPathA: &BGPPathA{
							LocalPref: 1000,
						},
						ASPathLen: 3,
					},
				},
			},
		},
	}

	for _, test := range tests {
		test.r.PathSelection()
		assert.Equal(t, test.expected, test.r.paths, test.name)
	}
}

func TestNewRoute(t *testing.T) {
	tests := []struct {
		name     string
		pfx      *bnet.Prefix
		path     *Path
		expected *Route
	}{
		{
			name: "BGP Path",
			pfx:  bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(),
			path: &Path{
				Type:    BGPPathType,
				BGPPath: &BGPPath{},
			},
			expected: &Route{
				pfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(),
				paths: []*Path{
					{
						Type:    BGPPathType,
						BGPPath: &BGPPath{},
					},
				},
			},
		},
		{
			name: "Empty Path",
			pfx:  bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(),
			expected: &Route{
				pfx:   bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(),
				paths: []*Path{},
			},
		},
	}

	for _, test := range tests {
		res := NewRoute(test.pfx, test.path)
		assert.Equal(t, test.expected, res)
	}
}

func TestNewRouteAddPath(t *testing.T) {
	tests := []struct {
		name     string
		pfx      *bnet.Prefix
		paths    []*Path
		expected *Route
	}{
		{
			name: "BGP Path",
			pfx:  bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(),
			paths: []*Path{
				{
					Type:    BGPPathType,
					BGPPath: &BGPPath{},
				},
				{
					Type:    BGPPathType,
					BGPPath: &BGPPath{},
				},
			},
			expected: &Route{
				pfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(),
				paths: []*Path{
					{
						Type:    BGPPathType,
						BGPPath: &BGPPath{},
					},
					{
						Type:    BGPPathType,
						BGPPath: &BGPPath{},
					},
				},
			},
		},
		{
			name: "Empty Path",
			pfx:  bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(),
			expected: &Route{
				pfx:   bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(),
				paths: []*Path{},
			},
		},
	}

	for _, test := range tests {
		res := NewRouteAddPath(test.pfx, test.paths)
		assert.Equal(t, test.expected, res)
	}
}

func TestPrefix(t *testing.T) {
	tests := []struct {
		name     string
		route    *Route
		expected *bnet.Prefix
	}{
		{
			name: "Prefix",
			route: &Route{
				pfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(),
			},
			expected: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(),
		},
	}

	for _, test := range tests {
		res := test.route.Prefix()
		assert.Equal(t, test.expected, res)
	}
}

func TestAddr(t *testing.T) {
	tests := []struct {
		name     string
		route    *Route
		expected *bnet.IP
	}{
		{
			name: "Prefix",
			route: &Route{
				pfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(),
			},
			expected: bnet.IPv4(0xa000000).Ptr(),
		},
	}

	for _, test := range tests {
		res := test.route.Addr()
		assert.Equal(t, test.expected, res)
	}
}

func TestPfxlen(t *testing.T) {
	tests := []struct {
		name     string
		route    *Route
		expected uint8
	}{
		{
			name: "Prefix",
			route: &Route{
				pfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(),
			},
			expected: 8,
		},
	}

	for _, test := range tests {
		res := test.route.Pfxlen()
		assert.Equal(t, test.expected, res)
	}
}

func TestAddPath(t *testing.T) {
	tests := []struct {
		name     string
		route    *Route
		newPath  *Path
		expected *Route
	}{
		{
			name: "Regular BGP path",
			route: NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &Path{
				Type:    BGPPathType,
				BGPPath: &BGPPath{},
			}),
			newPath: &Path{
				Type:    BGPPathType,
				BGPPath: &BGPPath{},
			},
			expected: &Route{
				pfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(),
				paths: []*Path{
					{
						Type:    BGPPathType,
						BGPPath: &BGPPath{},
					},
					{
						Type:    BGPPathType,
						BGPPath: &BGPPath{},
					},
				},
			},
		},
		{
			name: "Nil path",
			route: NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &Path{
				Type:    BGPPathType,
				BGPPath: &BGPPath{},
			}),
			newPath: nil,
			expected: &Route{
				pfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(),
				paths: []*Path{
					{
						Type:    BGPPathType,
						BGPPath: &BGPPath{},
					},
				},
			},
		},
	}

	for _, test := range tests {
		test.route.AddPath(test.newPath)
		assert.Equal(t, test.expected, test.route)
	}
}

func TestRemovePath(t *testing.T) {
	tests := []struct {
		name     string
		route    *Route
		remove   *Path
		expected int
	}{
		{
			name: "nil path to remove",
			route: &Route{
				paths: []*Path{
					{
						Type: BGPPathType,
						BGPPath: &BGPPath{
							BGPPathA: &BGPPathA{
								LocalPref: 100,
								Source:    bnet.IPv4(0).Ptr(),
								NextHop:   bnet.IPv4(0).Ptr(),
							},
						},
					},
					{
						Type: BGPPathType,
						BGPPath: &BGPPath{
							BGPPathA: &BGPPathA{
								LocalPref: 200,
								Source:    bnet.IPv4(0).Ptr(),
								NextHop:   bnet.IPv4(0).Ptr(),
							},
						},
					},
					{
						Type: BGPPathType,
						BGPPath: &BGPPath{
							BGPPathA: &BGPPathA{
								LocalPref: 300,
								Source:    bnet.IPv4(0).Ptr(),
								NextHop:   bnet.IPv4(0).Ptr(),
							},
						},
					},
				},
			},
			remove:   nil,
			expected: 3,
		},
		{
			name: "Remove middle",
			route: &Route{
				paths: []*Path{
					{
						Type: BGPPathType,
						BGPPath: &BGPPath{
							BGPPathA: &BGPPathA{
								LocalPref: 100,
								Source:    bnet.IPv4(0).Ptr(),
								NextHop:   bnet.IPv4(0).Ptr(),
							},
						},
					},
					{
						Type: BGPPathType,
						BGPPath: &BGPPath{
							BGPPathA: &BGPPathA{
								LocalPref: 200,
								Source:    bnet.IPv4(0).Ptr(),
								NextHop:   bnet.IPv4(0).Ptr(),
							},
						},
					},
					{
						Type: BGPPathType,
						BGPPath: &BGPPath{
							BGPPathA: &BGPPathA{
								LocalPref: 300,
								Source:    bnet.IPv4(0).Ptr(),
								NextHop:   bnet.IPv4(0).Ptr(),
							},
						},
					},
				},
			},
			remove: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{
						LocalPref: 200,
						Source:    bnet.IPv4(0).Ptr(),
						NextHop:   bnet.IPv4(0).Ptr(),
					},
				},
			},
			expected: 2,
		},
		{
			name: "Remove non-existent",
			route: &Route{
				paths: []*Path{
					{
						Type: BGPPathType,
						BGPPath: &BGPPath{
							BGPPathA: &BGPPathA{
								LocalPref: 10,
								Source:    bnet.IPv4(0).Ptr(),
								NextHop:   bnet.IPv4(0).Ptr(),
							},
						},
					},
					{
						Type: BGPPathType,
						BGPPath: &BGPPath{
							BGPPathA: &BGPPathA{
								LocalPref: 20,
								Source:    bnet.IPv4(0).Ptr(),
								NextHop:   bnet.IPv4(0).Ptr(),
							},
						},
					},
				},
			},
			remove: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{
						LocalPref: 50,
						Source:    bnet.IPv4(0).Ptr(),
						NextHop:   bnet.IPv4(0).Ptr(),
					},
				},
			},
			expected: 2,
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.route.RemovePath(test.remove))
	}
}

func TestRemovePathInternal(t *testing.T) {
	tests := []struct {
		name     string
		paths    []*Path
		remove   *Path
		expected []*Path
	}{
		{
			name: "Remove middle",
			paths: []*Path{
				{
					Type: BGPPathType,
					BGPPath: &BGPPath{
						BGPPathA: &BGPPathA{
							LocalPref: 100,
							Source:    bnet.IPv4(0).Ptr(),
							NextHop:   bnet.IPv4(0).Ptr(),
						},
					},
				},
				{
					Type: BGPPathType,
					BGPPath: &BGPPath{
						BGPPathA: &BGPPathA{
							LocalPref: 200,
							Source:    bnet.IPv4(0).Ptr(),
							NextHop:   bnet.IPv4(0).Ptr(),
						},
					},
				},
				{
					Type: BGPPathType,
					BGPPath: &BGPPath{
						BGPPathA: &BGPPathA{
							LocalPref: 300,
							Source:    bnet.IPv4(0).Ptr(),
							NextHop:   bnet.IPv4(0).Ptr(),
						},
					},
				},
			},
			remove: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{
						LocalPref: 200,
						Source:    bnet.IPv4(0).Ptr(),
						NextHop:   bnet.IPv4(0).Ptr(),
					},
				},
			},
			expected: []*Path{
				{
					Type: BGPPathType,
					BGPPath: &BGPPath{
						BGPPathA: &BGPPathA{
							LocalPref: 100,
							Source:    bnet.IPv4(0).Ptr(),
							NextHop:   bnet.IPv4(0).Ptr(),
						},
					},
				},
				{
					Type: BGPPathType,
					BGPPath: &BGPPath{
						BGPPathA: &BGPPathA{
							LocalPref: 300,
							Source:    bnet.IPv4(0).Ptr(),
							NextHop:   bnet.IPv4(0).Ptr(),
						},
					},
				},
			},
		},
		{
			name: "Remove non-existent",
			paths: []*Path{
				{
					Type: BGPPathType,
					BGPPath: &BGPPath{
						BGPPathA: &BGPPathA{
							LocalPref: 10,
							Source:    bnet.IPv4(0).Ptr(),
							NextHop:   bnet.IPv4(0).Ptr(),
						},
					},
				},
				{
					Type: BGPPathType,
					BGPPath: &BGPPath{
						BGPPathA: &BGPPathA{
							LocalPref: 20,
							Source:    bnet.IPv4(0).Ptr(),
							NextHop:   bnet.IPv4(0).Ptr(),
						},
					},
				},
			},
			remove: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{
						LocalPref: 50,
						Source:    bnet.IPv4(0).Ptr(),
						NextHop:   bnet.IPv4(0).Ptr(),
					},
				},
			},
			expected: []*Path{
				{
					Type: BGPPathType,
					BGPPath: &BGPPath{
						BGPPathA: &BGPPathA{
							LocalPref: 10,
							Source:    bnet.IPv4(0).Ptr(),
							NextHop:   bnet.IPv4(0).Ptr(),
						},
					},
				},
				{
					Type: BGPPathType,
					BGPPath: &BGPPath{
						BGPPathA: &BGPPathA{
							LocalPref: 20,
							Source:    bnet.IPv4(0).Ptr(),
							NextHop:   bnet.IPv4(0).Ptr(),
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		res := removePath(test.paths, test.remove)
		assert.Equal(t, test.expected, res)
	}
}

func TestReplacePath(t *testing.T) {
	tests := []struct {
		name     string
		route    *Route
		old      *Path
		new      *Path
		expected error
	}{
		{
			name: "Repalce first",
			route: &Route{
				paths: []*Path{
					{
						Type: BGPPathType,
						BGPPath: &BGPPath{
							BGPPathA: &BGPPathA{
								LocalPref: 100,
								Source:    bnet.IPv4(0).Ptr(),
								NextHop:   bnet.IPv4(0).Ptr(),
							},
						},
					},
					{
						Type: BGPPathType,
						BGPPath: &BGPPath{
							BGPPathA: &BGPPathA{
								LocalPref: 200,
								Source:    bnet.IPv4(0).Ptr(),
								NextHop:   bnet.IPv4(0).Ptr(),
							},
						},
					},
					{
						Type: BGPPathType,
						BGPPath: &BGPPath{
							BGPPathA: &BGPPathA{
								LocalPref: 300,
								Source:    bnet.IPv4(0).Ptr(),
								NextHop:   bnet.IPv4(0).Ptr(),
							},
						},
					},
				},
			},
			old: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{
						LocalPref: 100,
						Source:    bnet.IPv4(0).Ptr(),
						NextHop:   bnet.IPv4(0).Ptr(),
					},
				},
			},
			new: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{
						LocalPref: 123,
						Source:    bnet.IPv4(0).Ptr(),
						NextHop:   bnet.IPv4(0).Ptr(),
					},
				},
			},
			expected: nil,
		},
		{
			name: "Replace non-existent",
			route: &Route{
				paths: []*Path{
					{
						Type: BGPPathType,
						BGPPath: &BGPPath{
							BGPPathA: &BGPPathA{
								LocalPref: 10,
								Source:    bnet.IPv4(0).Ptr(),
								NextHop:   bnet.IPv4(0).Ptr(),
							},
						},
					},
					{
						Type: BGPPathType,
						BGPPath: &BGPPath{
							BGPPathA: &BGPPathA{
								LocalPref: 20,
								Source:    bnet.IPv4(0).Ptr(),
								NextHop:   bnet.IPv4(0).Ptr(),
							},
						},
					},
				},
			},
			old: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{
						LocalPref: 50,
						Source:    bnet.IPv4(0).Ptr(),
						NextHop:   bnet.IPv4(0).Ptr(),
					},
				},
			},
			expected: fmt.Errorf("Path not found"),
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.route.ReplacePath(test.old, test.new))
	}

}

func TestCopy(t *testing.T) {
	tests := []struct {
		name     string
		route    *Route
		expected *Route
	}{
		{
			name: "",
			route: &Route{
				pfx:       bnet.NewPfx(bnet.IPv4(1000), 8).Ptr(),
				ecmpPaths: 2,
				paths: []*Path{
					{
						Type: 100,
					},
				},
			},
			expected: &Route{
				pfx:       bnet.NewPfx(bnet.IPv4(1000), 8).Ptr(),
				ecmpPaths: 2,
				paths: []*Path{
					{
						Type: 100,
					},
				},
			},
		},
		{
			name: "",
		},
	}

	for _, test := range tests {
		res := test.route.Copy()
		assert.Equal(t, test.expected, res)
	}
}

func TestPaths(t *testing.T) {
	tests := []struct {
		name     string
		route    *Route
		path     *Path
		expected []*Path
	}{
		{
			name:     "nil Route",
			route:    nil,
			expected: nil,
		},
		{
			name: "nil Path",
			route: &Route{
				paths: nil,
			},
			expected: nil,
		},
		{
			name: "with path",
			route: &Route{
				paths: []*Path{},
			},
			expected: []*Path{},
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.route.Paths())
	}
}

func TestECMPPathCount(t *testing.T) {
	var r *Route
	assert.Equal(t, uint(0), r.ECMPPathCount())
	r = &Route{}
	assert.Equal(t, uint(0), r.ECMPPathCount())
	r.ecmpPaths = 12
	assert.Equal(t, uint(12), r.ECMPPathCount())
}

func TestBestPath(t *testing.T) {
	tests := []struct {
		route    *Route
		expected *Path
	}{
		{
			route:    nil,
			expected: nil,
		},
		{
			route:    &Route{},
			expected: nil,
		},
		{
			route: &Route{
				paths: []*Path{
					{
						Type: StaticPathType,
						StaticPath: &StaticPath{
							NextHop: bnet.IPv4(32).Ptr(),
						},
					},
				},
			},
			expected: &Path{
				Type: StaticPathType,
				StaticPath: &StaticPath{
					NextHop: bnet.IPv4(32).Ptr(),
				},
			},
		},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.expected, tc.route.BestPath())
	}
}

func TestECMPPaths(t *testing.T) {
	tests := []struct {
		route    *Route
		expected []*Path
	}{
		{
			route:    nil,
			expected: nil,
		},
		{
			route:    &Route{},
			expected: nil,
		},
		{
			route: &Route{
				ecmpPaths: 2,
				paths: []*Path{
					{
						Type: StaticPathType,
						StaticPath: &StaticPath{
							NextHop: bnet.IPv4(32).Ptr(),
						},
					},
					{
						Type: StaticPathType,
						StaticPath: &StaticPath{
							NextHop: bnet.IPv4(32).Ptr(),
						},
					},
				},
			},
			expected: []*Path{
				{
					Type: StaticPathType,
					StaticPath: &StaticPath{
						NextHop: bnet.IPv4(32).Ptr(),
					},
				},
				{
					Type: StaticPathType,
					StaticPath: &StaticPath{
						NextHop: bnet.IPv4(32).Ptr(),
					},
				},
			},
		},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.expected, tc.route.ECMPPaths())
	}
}

func TestRouteEqual(t *testing.T) {
	tests := []struct {
		a     *Route
		b     *Route
		equal bool
	}{
		{
			a: &Route{
				pfx:       bnet.NewPfx(bnet.IPv4(0), 0).Ptr(),
				ecmpPaths: 2,
				paths: []*Path{
					{
						Type: StaticPathType,
						StaticPath: &StaticPath{
							NextHop: bnet.IPv4FromOctets(192, 168, 0, 1).Ptr(),
						},
					},
					{
						Type: StaticPathType,
						StaticPath: &StaticPath{
							NextHop: bnet.IPv4FromOctets(192, 168, 1, 1).Ptr(),
						},
					},
				},
			},
			b: &Route{
				pfx:       bnet.NewPfx(bnet.IPv4(0), 0).Ptr(),
				ecmpPaths: 2,
				paths: []*Path{
					{
						Type: StaticPathType,
						StaticPath: &StaticPath{
							NextHop: bnet.IPv4FromOctets(192, 168, 0, 1).Ptr(),
						},
					},
					{
						Type: StaticPathType,
						StaticPath: &StaticPath{
							NextHop: bnet.IPv4FromOctets(192, 168, 1, 1).Ptr(),
						},
					},
				},
			},
			equal: true,
		},
		{
			a: &Route{
				pfx:       bnet.NewPfx(bnet.IPv4(0), 0).Ptr(),
				ecmpPaths: 2,
				paths: []*Path{
					{
						Type: StaticPathType,
						StaticPath: &StaticPath{
							NextHop: bnet.IPv4FromOctets(192, 168, 0, 1).Ptr(),
						},
					},
					{
						Type: StaticPathType,
						StaticPath: &StaticPath{
							NextHop: bnet.IPv4FromOctets(192, 168, 1, 1).Ptr(),
						},
					},
				},
			},
			b: &Route{
				pfx:       bnet.NewPfx(bnet.IPv4(0), 0).Ptr(),
				ecmpPaths: 2,
				paths: []*Path{
					{
						Type: StaticPathType,
						StaticPath: &StaticPath{
							NextHop: bnet.IPv4FromOctets(192, 168, 1, 1).Ptr(),
						},
					},
					{
						Type: StaticPathType,
						StaticPath: &StaticPath{
							NextHop: bnet.IPv4FromOctets(192, 168, 2, 1).Ptr(),
						},
					},
				},
			},
			equal: false,
		},
		{
			a: &Route{
				pfx:       bnet.NewPfx(bnet.IPv4(0), 0).Ptr(),
				ecmpPaths: 2,
				paths: []*Path{
					{
						Type: StaticPathType,
						StaticPath: &StaticPath{
							NextHop: bnet.IPv4FromOctets(192, 168, 0, 1).Ptr(),
						},
					},
					{
						Type: StaticPathType,
						StaticPath: &StaticPath{
							NextHop: bnet.IPv4FromOctets(192, 168, 1, 1).Ptr(),
						},
					},
				},
			},
			b: &Route{
				pfx:       bnet.NewPfx(bnet.IPv4(0), 0).Ptr(),
				ecmpPaths: 2,
				paths: []*Path{
					{
						Type: StaticPathType,
						StaticPath: &StaticPath{
							NextHop: bnet.IPv4FromOctets(192, 168, 1, 1).Ptr(),
						},
					},
				},
			},
			equal: false,
		},
	}

	for _, tc := range tests {
		res := tc.a.Equal(tc.b)
		assert.Equal(t, tc.equal, res)
	}
}

func TestRouteIsBGPOriginatedBy(t *testing.T) {
	tests := []struct {
		name   string
		r      *Route
		origBy uint32
		isOrig bool
	}{
		{
			name: "Single AS Path, correct originating AS",
			r: &Route{
				pfx: bnet.NewPfx(bnet.IPv4(0), 0).Ptr(),
				paths: []*Path{
					{
						Type: StaticPathType,
						BGPPath: &BGPPath{
							ASPath: &types.ASPath{
								types.ASPathSegment{
									Type: types.ASSequence,
									ASNs: []uint32{65000, 65001, 65002, 65003},
								},
							},
						},
					},
				},
			},
			origBy: 65000,
			isOrig: false,
		},
		{
			name: "Single AS Path, wrong originating AS",
			r: &Route{
				pfx: bnet.NewPfx(bnet.IPv4(0), 0).Ptr(),
				paths: []*Path{
					{
						Type: StaticPathType,
						BGPPath: &BGPPath{
							ASPath: &types.ASPath{
								types.ASPathSegment{
									Type: types.ASSequence,
									ASNs: []uint32{65003, 65002, 65001, 65000},
								},
							},
						},
					},
				},
			},
			origBy: 65000,
			isOrig: true,
		},
		{
			name: "Empty AS Path",
			r: &Route{
				pfx: bnet.NewPfx(bnet.IPv4(0), 0).Ptr(),
				paths: []*Path{
					{
						Type: StaticPathType,
						BGPPath: &BGPPath{
							ASPath: &types.ASPath{
								types.ASPathSegment{
									Type: types.ASSequence,
									ASNs: []uint32{},
								},
							},
						},
					},
				},
			},
			origBy: 65000,
			isOrig: false,
		},
	}

	for _, tc := range tests {
		res := tc.r.IsBGPOriginatedBy(tc.origBy)
		assert.Equal(t, tc.isOrig, res, tc.name)
	}
}

func TestRouteToProto(t *testing.T) {
	tests := []struct {
		route  *Route
		result *api.Route
	}{
		{
			route: &Route{
				pfx: bnet.NewPfx(bnet.IPv4(0), 0).Ptr(),
				paths: []*Path{
					{
						Type: StaticPathType,
						StaticPath: &StaticPath{
							NextHop: bnet.IPv4FromOctets(192, 168, 0, 1).Ptr(),
						},
					},
					{
						Type: StaticPathType,
						StaticPath: &StaticPath{
							NextHop: bnet.IPv4FromOctets(192, 168, 1, 1).Ptr(),
						},
					},
				},
			},
			result: &api.Route{
				Pfx: bnet.NewPfx(bnet.IPv4(0), 0).Ptr().ToProto(),
				Paths: []*api.Path{
					{
						Type: api.Path_Static,
						StaticPath: &api.StaticPath{
							NextHop: bnet.IPv4FromOctets(192, 168, 0, 1).Ptr().ToProto(),
						},
					},
					{
						Type: api.Path_Static,
						StaticPath: &api.StaticPath{
							NextHop: bnet.IPv4FromOctets(192, 168, 1, 1).Ptr().ToProto(),
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.result, tc.route.ToProto())
	}
}

func TestRouteFromProtoRoute(t *testing.T) {
	tests := []struct {
		name       string
		protoRoute *api.Route
		result     *Route
	}{
		{
			name: "Static route",
			protoRoute: &api.Route{
				Pfx: bnet.NewPfx(bnet.IPv4(0), 0).Ptr().ToProto(),
				Paths: []*api.Path{
					{
						Type: api.Path_Static,
						StaticPath: &api.StaticPath{
							NextHop: bnet.IPv4FromOctets(192, 168, 0, 1).Ptr().ToProto(),
						},
					},
					{
						Type: api.Path_Static,
						StaticPath: &api.StaticPath{
							NextHop: bnet.IPv4FromOctets(192, 168, 1, 1).Ptr().ToProto(),
						},
					},
				},
			},
			result: &Route{
				pfx:       bnet.NewPfx(bnet.IPv4(0), 0).Ptr(),
				ecmpPaths: 0,
				paths: []*Path{
					{
						Type: StaticPathType,
						StaticPath: &StaticPath{
							NextHop: bnet.IPv4FromOctets(192, 168, 0, 1).Ptr(),
						},
					},
					{
						Type: StaticPathType,
						StaticPath: &StaticPath{
							NextHop: bnet.IPv4FromOctets(192, 168, 1, 1).Ptr(),
						},
					},
				},
			},
		},
		{
			name: "Bare BGP route",
			protoRoute: &api.Route{
				Pfx: bnet.NewPfx(bnet.IPv4(0), 0).Ptr().ToProto(),
				Paths: []*api.Path{
					{
						Type: api.Path_BGP,
						BgpPath: &api.BGPPath{
							NextHop: bnet.IPv4FromOctets(192, 168, 0, 1).Ptr().ToProto(),
							Source:  bnet.IPv4FromOctets(192, 168, 0, 2).Ptr().ToProto(),
						},
					},
				},
			},

			result: &Route{
				pfx: bnet.NewPfx(bnet.IPv4(0), 0).Ptr(),
				paths: []*Path{
					{
						Type: BGPPathType,
						BGPPath: &BGPPath{
							ASPath: &types.ASPath{},
							BGPPathA: &BGPPathA{
								NextHop: bnet.IPv4FromOctets(192, 168, 0, 1).Ptr(),
								Source:  bnet.IPv4FromOctets(192, 168, 0, 2).Ptr(),
							},
						},
					},
				},
			},
		},
		{
			name: "Bare BGP route",
			protoRoute: &api.Route{
				Pfx: bnet.NewPfx(bnet.IPv4(0), 0).Ptr().ToProto(),
				Paths: []*api.Path{
					{
						Type: api.Path_BGP,
						BgpPath: &api.BGPPath{
							NextHop: bnet.IPv4FromOctets(192, 168, 0, 1).Ptr().ToProto(),
							Source:  bnet.IPv4FromOctets(192, 168, 0, 2).Ptr().ToProto(),
						},
					},
				},
			},

			result: &Route{
				pfx: bnet.NewPfx(bnet.IPv4(0), 0).Ptr(),
				paths: []*Path{
					{
						Type: BGPPathType,
						BGPPath: &BGPPath{
							ASPath: &types.ASPath{},
							BGPPathA: &BGPPathA{
								NextHop: bnet.IPv4FromOctets(192, 168, 0, 1).Ptr(),
								Source:  bnet.IPv4FromOctets(192, 168, 0, 2).Ptr(),
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.result, RouteFromProtoRoute(tc.protoRoute, false))
	}
}
