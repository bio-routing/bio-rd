package route

import (
	"testing"

	"github.com/stretchr/testify/assert"

	bnet "github.com/bio-routing/bio-rd/net"
)

func TestNewRoute(t *testing.T) {
	tests := []struct {
		name     string
		pfx      bnet.Prefix
		path     *Path
		expected *Route
	}{
		{
			name: "BGP Path",
			pfx:  bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8),
			path: &Path{
				Type:    BGPPathType,
				BGPPath: &BGPPath{},
			},
			expected: &Route{
				pfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8),
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
			pfx:  bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8),
			expected: &Route{
				pfx:   bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8),
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
		expected bnet.Prefix
	}{
		{
			name: "Prefix",
			route: &Route{
				pfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8),
			},
			expected: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8),
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
		expected bnet.IP
	}{
		{
			name: "Prefix",
			route: &Route{
				pfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8),
			},
			expected: bnet.IPv4(0xa000000),
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
				pfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8),
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
			route: NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8), &Path{
				Type:    BGPPathType,
				BGPPath: &BGPPath{},
			}),
			newPath: &Path{
				Type:    BGPPathType,
				BGPPath: &BGPPath{},
			},
			expected: &Route{
				pfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8),
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
			route: NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8), &Path{
				Type:    BGPPathType,
				BGPPath: &BGPPath{},
			}),
			newPath: nil,
			expected: &Route{
				pfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8),
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
						LocalPref: 100,
					},
				},
				{
					Type: BGPPathType,
					BGPPath: &BGPPath{
						LocalPref: 200,
					},
				},
				{
					Type: BGPPathType,
					BGPPath: &BGPPath{
						LocalPref: 300,
					},
				},
			},
			remove: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					LocalPref: 200,
				},
			},
			expected: []*Path{
				{
					Type: BGPPathType,
					BGPPath: &BGPPath{
						LocalPref: 100,
					},
				},
				{
					Type: BGPPathType,
					BGPPath: &BGPPath{
						LocalPref: 300,
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
						LocalPref: 10,
					},
				},
				{
					Type: BGPPathType,
					BGPPath: &BGPPath{
						LocalPref: 20,
					},
				},
			},
			remove: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					LocalPref: 50,
				},
			},
			expected: []*Path{
				{
					Type: BGPPathType,
					BGPPath: &BGPPath{
						LocalPref: 10,
					},
				},
				{
					Type: BGPPathType,
					BGPPath: &BGPPath{
						LocalPref: 20,
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
				pfx:       bnet.NewPfx(bnet.IPv4(1000), 8),
				ecmpPaths: 2,
				paths: []*Path{
					{
						Type: 100,
					},
				},
			},
			expected: &Route{
				pfx:       bnet.NewPfx(bnet.IPv4(1000), 8),
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
							NextHop: bnet.IPv4(32),
						},
					},
				},
			},
			expected: &Path{
				Type: StaticPathType,
				StaticPath: &StaticPath{
					NextHop: bnet.IPv4(32),
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
							NextHop: bnet.IPv4(32),
						},
					},
					{
						Type: StaticPathType,
						StaticPath: &StaticPath{
							NextHop: bnet.IPv4(32),
						},
					},
				},
			},
			expected: []*Path{
				{
					Type: StaticPathType,
					StaticPath: &StaticPath{
						NextHop: bnet.IPv4(32),
					},
				},
				{
					Type: StaticPathType,
					StaticPath: &StaticPath{
						NextHop: bnet.IPv4(32),
					},
				},
			},
		},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.expected, tc.route.ECMPPaths())
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		a     *Route
		b     *Route
		equal bool
	}{
		{
			a: &Route{
				ecmpPaths: 2,
				paths: []*Path{
					{
						Type: StaticPathType,
						StaticPath: &StaticPath{
							NextHop: bnet.IPv4FromOctets(192, 168, 0, 1),
						},
					},
					{
						Type: StaticPathType,
						StaticPath: &StaticPath{
							NextHop: bnet.IPv4FromOctets(192, 168, 1, 1),
						},
					},
				},
			},
			b: &Route{
				ecmpPaths: 2,
				paths: []*Path{
					{
						Type: StaticPathType,
						StaticPath: &StaticPath{
							NextHop: bnet.IPv4FromOctets(192, 168, 0, 1),
						},
					},
					{
						Type: StaticPathType,
						StaticPath: &StaticPath{
							NextHop: bnet.IPv4FromOctets(192, 168, 1, 1),
						},
					},
				},
			},
			equal: true,
		}, {
			a: &Route{
				ecmpPaths: 2,
				paths: []*Path{
					{
						Type: StaticPathType,
						StaticPath: &StaticPath{
							NextHop: bnet.IPv4FromOctets(192, 168, 0, 1),
						},
					},
					{
						Type: StaticPathType,
						StaticPath: &StaticPath{
							NextHop: bnet.IPv4FromOctets(192, 168, 1, 1),
						},
					},
				},
			},
			b: &Route{
				ecmpPaths: 2,
				paths: []*Path{
					{
						Type: StaticPathType,
						StaticPath: &StaticPath{
							NextHop: bnet.IPv4FromOctets(192, 168, 1, 1),
						},
					},
					{
						Type: StaticPathType,
						StaticPath: &StaticPath{
							NextHop: bnet.IPv4FromOctets(192, 168, 2, 1),
						},
					},
				},
			},
			equal: false,
		},
		{
			a: &Route{
				ecmpPaths: 2,
				paths: []*Path{
					{
						Type: StaticPathType,
						StaticPath: &StaticPath{
							NextHop: bnet.IPv4FromOctets(192, 168, 0, 1),
						},
					},
					{
						Type: StaticPathType,
						StaticPath: &StaticPath{
							NextHop: bnet.IPv4FromOctets(192, 168, 1, 1),
						},
					},
				},
			},
			b: &Route{
				ecmpPaths: 2,
				paths: []*Path{
					{
						Type: StaticPathType,
						StaticPath: &StaticPath{
							NextHop: bnet.IPv4FromOctets(192, 168, 1, 1),
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
