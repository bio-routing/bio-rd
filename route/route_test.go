package route

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bio-routing/bio-rd/net"
	bnet "github.com/bio-routing/bio-rd/net"
)

func TestNewRoute(t *testing.T) {
	tests := []struct {
		name     string
		pfx      *bnet.Prefix
		path     *Path
		expected *Route
	}{
		{
			name: "BGP Path",
			pfx:  bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0).Ptr(), 8).Ptr(),
			path: &Path{
				Type:    BGPPathType,
				BGPPath: &BGPPath{},
			},
			expected: &Route{
				pfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0).Ptr(), 8).Ptr(),
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
			pfx:  bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0).Ptr(), 8).Ptr(),
			expected: &Route{
				pfx:   bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0).Ptr(), 8).Ptr(),
				paths: []*Path{},
			},
		},
	}

	for _, test := range tests {
		res := NewRoute(test.pfx, test.path)
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
				pfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0).Ptr(), 8).Ptr(),
			},
			expected: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0).Ptr(), 8).Ptr(),
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
				pfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0).Ptr(), 8).Ptr(),
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
				pfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0).Ptr(), 8).Ptr(),
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
			route: NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0).Ptr(), 8).Ptr(), &Path{
				Type:    BGPPathType,
				BGPPath: &BGPPath{},
			}),
			newPath: &Path{
				Type:    BGPPathType,
				BGPPath: &BGPPath{},
			},
			expected: &Route{
				pfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0).Ptr(), 8).Ptr(),
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
			route: NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0).Ptr(), 8).Ptr(), &Path{
				Type:    BGPPathType,
				BGPPath: &BGPPath{},
			}),
			newPath: nil,
			expected: &Route{
				pfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0).Ptr(), 8).Ptr(),
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

func TestRouteRemovePath(t *testing.T) {
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
							Source:    net.IPv4(0).Ptr(),
							NextHop:   net.IPv4(0).Ptr(),
						},
					},
				},
				{
					Type: BGPPathType,
					BGPPath: &BGPPath{
						BGPPathA: &BGPPathA{
							LocalPref: 200,
							Source:    net.IPv4(0).Ptr(),
							NextHop:   net.IPv4(0).Ptr(),
						},
					},
				},
				{
					Type: BGPPathType,
					BGPPath: &BGPPath{
						BGPPathA: &BGPPathA{
							LocalPref: 300,
							Source:    net.IPv4(0).Ptr(),
							NextHop:   net.IPv4(0).Ptr(),
						},
					},
				},
			},
			remove: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{
						LocalPref: 200,
						Source:    net.IPv4(0).Ptr(),
						NextHop:   net.IPv4(0).Ptr(),
					},
				},
			},
			expected: []*Path{
				{
					Type: BGPPathType,
					BGPPath: &BGPPath{
						BGPPathA: &BGPPathA{
							LocalPref: 100,
							Source:    net.IPv4(0).Ptr(),
							NextHop:   net.IPv4(0).Ptr(),
						},
					},
				},
				{
					Type: BGPPathType,
					BGPPath: &BGPPath{
						BGPPathA: &BGPPathA{
							LocalPref: 300,
							Source:    net.IPv4(0).Ptr(),
							NextHop:   net.IPv4(0).Ptr(),
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
							Source:    net.IPv4(0).Ptr(),
							NextHop:   net.IPv4(0).Ptr(),
						},
					},
				},
				{
					Type: BGPPathType,
					BGPPath: &BGPPath{
						BGPPathA: &BGPPathA{
							LocalPref: 20,
							Source:    net.IPv4(0).Ptr(),
							NextHop:   net.IPv4(0).Ptr(),
						},
					},
				},
			},
			remove: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{
						LocalPref: 50,
						Source:    net.IPv4(0).Ptr(),
						NextHop:   net.IPv4(0).Ptr(),
					},
				},
			},
			expected: []*Path{
				{
					Type: BGPPathType,
					BGPPath: &BGPPath{
						BGPPathA: &BGPPathA{
							LocalPref: 10,
							Source:    net.IPv4(0).Ptr(),
							NextHop:   net.IPv4(0).Ptr(),
						},
					},
				},
				{
					Type: BGPPathType,
					BGPPath: &BGPPath{
						BGPPathA: &BGPPathA{
							LocalPref: 20,
							Source:    net.IPv4(0).Ptr(),
							NextHop:   net.IPv4(0).Ptr(),
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

func TestCopy(t *testing.T) {
	tests := []struct {
		name     string
		route    *Route
		expected *Route
	}{
		{
			name: "",
			route: &Route{
				pfx:       bnet.NewPfx(bnet.IPv4(1000).Ptr(), 8).Ptr(),
				ecmpPaths: 2,
				paths: []*Path{
					{
						Type: 100,
					},
				},
			},
			expected: &Route{
				pfx:       bnet.NewPfx(bnet.IPv4(1000).Ptr(), 8).Ptr(),
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
				pfx:       net.NewPfx(net.IPv4(0).Ptr(), 0).Ptr(),
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
				pfx:       net.NewPfx(net.IPv4(0).Ptr(), 0).Ptr(),
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
		}, {
			a: &Route{
				pfx:       net.NewPfx(net.IPv4(0).Ptr(), 0).Ptr(),
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
				pfx:       net.NewPfx(net.IPv4(0).Ptr(), 0).Ptr(),
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
				pfx:       net.NewPfx(net.IPv4(0).Ptr(), 0).Ptr(),
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
				pfx:       net.NewPfx(net.IPv4(0).Ptr(), 0).Ptr(),
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
