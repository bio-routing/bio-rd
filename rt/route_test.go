package rt

import (
	"testing"

	net "github.com/bio-routing/bio-rd/net"
	"github.com/stretchr/testify/assert"
)

func TestNewRoute(t *testing.T) {
	tests := []struct {
		name     string
		pfx      *net.Prefix
		paths    []*Path
		expected *Route
	}{
		{
			name: "Test #1",
			pfx:  net.NewPfx(158798889, 24),
			paths: []*Path{
				{
					Type: 2,
					StaticPath: &StaticPath{
						NextHop: 56963289,
					},
				},
			},
			expected: &Route{
				pfx: net.NewPfx(158798889, 24),
				paths: []*Path{
					{
						Type: 2,
						StaticPath: &StaticPath{
							NextHop: 56963289,
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		res := NewRoute(test.pfx, test.paths)
		assert.Equal(t, test.expected, res)
	}
}

func TestPfxlen(t *testing.T) {
	tests := []struct {
		name     string
		pfx      *net.Prefix
		expected uint8
	}{
		{
			name:     "Test #1",
			pfx:      net.NewPfx(158798889, 24),
			expected: 24,
		},
	}

	for _, test := range tests {
		r := NewRoute(test.pfx, nil)
		res := r.Pfxlen()
		assert.Equal(t, test.expected, res)
	}
}

func TestPrefix(t *testing.T) {
	tests := []struct {
		name     string
		pfx      *net.Prefix
		expected *net.Prefix
	}{
		{
			name:     "Test #1",
			pfx:      net.NewPfx(158798889, 24),
			expected: net.NewPfx(158798889, 24),
		},
	}

	for _, test := range tests {
		r := NewRoute(test.pfx, nil)
		res := r.Prefix()
		assert.Equal(t, test.expected, res)
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

func TestEqual(t *testing.T) {
	tests := []struct {
		name     string
		pathA    *Path
		pathB    *Path
		expected bool
	}{
		{
			name: "Unequal types",
			pathA: &Path{
				Type: 1,
			},
			pathB: &Path{
				Type: 2,
			},
			expected: false,
		},
		{
			name: "Unequal attributes",
			pathA: &Path{
				Type: 2,
				BGPPath: &BGPPath{
					LocalPref: 100,
				},
			},
			pathB: &Path{
				Type: 2,
				BGPPath: &BGPPath{
					LocalPref: 200,
				},
			},
			expected: false,
		},
		{
			name: "Equal",
			pathA: &Path{
				Type: 2,
				BGPPath: &BGPPath{
					LocalPref: 100,
				},
			},
			pathB: &Path{
				Type: 2,
				BGPPath: &BGPPath{
					LocalPref: 100,
				},
			},
			expected: true,
		},
	}

	for _, test := range tests {
		res := test.pathA.Equal(test.pathB)
		assert.Equal(t, test.expected, res)
	}
}

func TestAddPath(t *testing.T) {
	tests := []struct {
		name     string
		route    *Route
		new      *Path
		expected *Route
	}{
		{
			name: "Add a new best path",
			route: &Route{
				paths: []*Path{
					{
						Type: 2,
						BGPPath: &BGPPath{
							LocalPref: 100,
						},
					},
				},
			},
			new: &Path{
				Type: 2,
				BGPPath: &BGPPath{
					LocalPref: 200,
				},
			},
			expected: &Route{
				activePaths: []*Path{
					{
						Type: 2,
						BGPPath: &BGPPath{
							LocalPref: 200,
						},
					},
				},
				paths: []*Path{
					{
						Type: 2,
						BGPPath: &BGPPath{
							LocalPref: 100,
						},
					},
					{
						Type: 2,
						BGPPath: &BGPPath{
							LocalPref: 200,
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		test.route.AddPath(test.new)
		assert.Equal(t, test.expected, test.route)
	}
}

func TestAddPaths(t *testing.T) {
	tests := []struct {
		name     string
		route    *Route
		new      []*Path
		expected *Route
	}{
		{
			name: "Add 2 new paths including a new best path",
			route: &Route{
				paths: []*Path{
					{
						Type: BGPPathType,
						BGPPath: &BGPPath{
							LocalPref: 100,
						},
					},
				},
			},
			new: []*Path{
				{
					Type: BGPPathType,
					BGPPath: &BGPPath{
						LocalPref: 200,
					},
				},
				{
					Type: BGPPathType,
					BGPPath: &BGPPath{
						LocalPref: 50,
					},
				},
			},
			expected: &Route{
				activePaths: []*Path{
					{
						Type: BGPPathType,
						BGPPath: &BGPPath{
							LocalPref: 200,
						},
					},
				},
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
							LocalPref: 50,
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		test.route.AddPaths(test.new)
		assert.Equal(t, test.expected, test.route)
	}
}

func TestGetBestProtocol(t *testing.T) {
	tests := []struct {
		name     string
		input    []*Path
		expected uint8
	}{
		{
			name: "Foo",
			input: []*Path{
				{
					Type: BGPPathType,
				},
				{
					Type: StaticPathType,
				},
			},
			expected: StaticPathType,
		},
	}

	for _, test := range tests {
		res := getBestProtocol(test.input)
		assert.Equal(t, test.expected, res)
	}
}

func TestMissingPaths(t *testing.T) {
	tests := []struct {
		name     string
		a        []*Path
		b        []*Path
		expected []*Path
	}{
		{
			name: "None missing #2",
			a: []*Path{
				{
					Type: 1,
				},
				{
					Type: 2,
				},
			},
			b: []*Path{
				{
					Type: 1,
				},
				{
					Type: 2,
				},
				{
					Type: 3,
				},
			},
			expected: []*Path{},
		},
		{
			name: "None missing",
			a: []*Path{
				{
					Type: 1,
				},
				{
					Type: 2,
				},
			},
			b: []*Path{
				{
					Type: 1,
				},
				{
					Type: 2,
				},
			},
			expected: []*Path{},
		},
		{
			name: "One missing",
			a: []*Path{
				{
					Type: 1,
				},
				{
					Type: 2,
				},
			},
			b: []*Path{
				{
					Type: 1,
				},
			},
			expected: []*Path{
				{
					Type: 2,
				},
			},
		},
	}

	for _, test := range tests {
		res := missingPaths(test.a, test.b)
		assert.Equal(t, test.expected, res)
	}
}
