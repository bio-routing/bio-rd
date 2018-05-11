package route

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bio-routing/bio-rd/net"
)

func TestNewRoute(t *testing.T) {
	tests := []struct {
		name     string
		pfx      net.Prefix
		path     *Path
		expected *Route
	}{
		{
			name: "BGP Path",
			pfx:  net.NewPfx(strAddr("10.0.0.0"), 8),
			path: &Path{
				Type:    BGPPathType,
				BGPPath: &BGPPath{},
			},
			expected: &Route{
				pfx: net.NewPfx(strAddr("10.0.0.0"), 8),
				paths: []*Path{
					&Path{
						Type:    BGPPathType,
						BGPPath: &BGPPath{},
					},
				},
			},
		},
		{
			name: "Empty Path",
			pfx:  net.NewPfx(strAddr("10.0.0.0"), 8),
			expected: &Route{
				pfx:   net.NewPfx(strAddr("10.0.0.0"), 8),
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
		expected net.Prefix
	}{
		{
			name: "Prefix",
			route: &Route{
				pfx: net.NewPfx(strAddr("10.0.0.0"), 8),
			},
			expected: net.NewPfx(strAddr("10.0.0.0"), 8),
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
		expected uint32
	}{
		{
			name: "Prefix",
			route: &Route{
				pfx: net.NewPfx(strAddr("10.0.0.0"), 8),
			},
			expected: 0xa000000,
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
				pfx: net.NewPfx(strAddr("10.0.0.0"), 8),
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
			route: NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), &Path{
				Type:    BGPPathType,
				BGPPath: &BGPPath{},
			}),
			newPath: &Path{
				Type:    BGPPathType,
				BGPPath: &BGPPath{},
			},
			expected: &Route{
				pfx: net.NewPfx(strAddr("10.0.0.0"), 8),
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
			route: NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), &Path{
				Type:    BGPPathType,
				BGPPath: &BGPPath{},
			}),
			newPath: nil,
			expected: &Route{
				pfx: net.NewPfx(strAddr("10.0.0.0"), 8),
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
				pfx:       net.NewPfx(1000, 8),
				ecmpPaths: 2,
				paths: []*Path{
					{
						Type: 100,
					},
				},
			},
			expected: &Route{
				pfx:       net.NewPfx(1000, 8),
				ecmpPaths: 2,
				paths: []*Path{
					{
						Type: 100,
					},
				},
			},
		},
	}

	for _, test := range tests {
		res := test.route.Copy()
		assert.Equal(t, test.expected, res)
	}
}

func strAddr(s string) uint32 {
	ret, _ := net.StrToAddr(s)
	return ret
}
