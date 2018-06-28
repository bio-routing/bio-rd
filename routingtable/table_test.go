package routingtable

import (
	"testing"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/stretchr/testify/assert"
)

func TestAddPath(t *testing.T) {
	tests := []struct {
		name     string
		routes   []*route.Route
		expected *node
	}{
		{
			name: "Insert first node",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
			},
			expected: &node{
				route: route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				skip:  8,
			},
		},
		{
			name: "Insert duplicate node",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
			},
			expected: &node{
				route: route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				skip:  8,
			},
		},
		{
			name: "Insert triangle",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 9), nil),
				route.NewRoute(net.NewPfx(strAddr("10.128.0.0"), 9), nil),
			},
			expected: &node{
				route: route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				skip:  8,
				l: &node{
					route: route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 9), nil),
				},
				h: &node{
					route: route.NewRoute(net.NewPfx(strAddr("10.128.0.0"), 9), nil),
				},
			},
		},
		{
			name: "Insert disjunct prefixes",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				route.NewRoute(net.NewPfx(strAddr("11.100.123.0"), 24), nil),
			},
			expected: &node{
				route: route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 7), nil),
				skip:  7,
				dummy: true,
				l: &node{
					route: route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				},
				h: &node{
					route: route.NewRoute(net.NewPfx(strAddr("11.100.123.0"), 24), nil),
					skip:  16,
				},
			},
		},
		{
			name: "Insert disjunct prefixes plus one child low",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				route.NewRoute(net.NewPfx(strAddr("11.100.123.0"), 24), nil),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 12), nil),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 10), nil),
			},
			expected: &node{
				route: route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 7), nil),
				skip:  7,
				dummy: true,
				l: &node{
					route: route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
					l: &node{
						skip:  1,
						route: route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 10), nil),
						l: &node{
							skip:  1,
							route: route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 12), nil),
						},
					},
				},
				h: &node{
					route: route.NewRoute(net.NewPfx(strAddr("11.100.123.0"), 24), nil),
					skip:  16,
				},
			},
		},
		{
			name: "Insert disjunct prefixes plus one child high",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				route.NewRoute(net.NewPfx(strAddr("11.100.123.0"), 24), nil),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 12), nil),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 10), nil),
				route.NewRoute(net.NewPfx(strAddr("11.100.123.128"), 25), nil),
			},
			expected: &node{
				route: route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 7), nil),
				skip:  7,
				dummy: true,
				l: &node{
					route: route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
					l: &node{
						skip:  1,
						route: route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 10), nil),
						l: &node{
							skip:  1,
							route: route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 12), nil),
						},
					},
				},
				h: &node{
					route: route.NewRoute(net.NewPfx(strAddr("11.100.123.0"), 24), nil),
					skip:  16,
					h: &node{
						route: route.NewRoute(net.NewPfx(strAddr("11.100.123.128"), 25), nil),
					},
				},
			},
		},
	}

	for _, test := range tests {
		rt := NewRoutingTable()
		for _, route := range test.routes {
			rt.AddPath(route.Prefix(), nil)
		}

		assert.Equal(t, test.expected, rt.root)
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name     string
		routes   []*route.Route
		needle   net.Prefix
		expected *route.Route
	}{
		{
			name: "Test 1: Search pfx and dump route + more specifics",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				route.NewRoute(net.NewPfx(strAddr("11.100.123.0"), 24), nil),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 12), nil),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 10), nil),
			},
			needle:   net.NewPfx(strAddr("10.0.0.0"), 8),
			expected: route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
		},
		{
			name: "Test 2: Search pfx and don't dump more specifics",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				route.NewRoute(net.NewPfx(strAddr("11.100.123.0"), 24), nil),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 12), nil),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 10), nil),
			},
			needle:   net.NewPfx(strAddr("10.0.0.0"), 8),
			expected: route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
		},
		{
			name:     "Test 3: Empty table",
			routes:   []*route.Route{},
			needle:   net.NewPfx(strAddr("10.0.0.0"), 32),
			expected: nil,
		},
		{
			name: "Test 4: Get Dummy",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				route.NewRoute(net.NewPfx(strAddr("11.100.123.0"), 24), nil),
			},
			needle:   net.NewPfx(strAddr("10.0.0.0"), 7),
			expected: nil,
		},
		{
			name: "Test 5",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				route.NewRoute(net.NewPfx(strAddr("11.100.123.0"), 24), nil),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 12), nil),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 10), nil),
			},
			needle:   net.NewPfx(strAddr("11.100.123.0"), 24),
			expected: route.NewRoute(net.NewPfx(strAddr("11.100.123.0"), 24), nil),
		},
		{
			name: "Test 4: Get nonexistent #1",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				route.NewRoute(net.NewPfx(strAddr("11.100.123.0"), 24), nil),
			},
			needle:   net.NewPfx(strAddr("10.0.0.0"), 10),
			expected: nil,
		},
		{
			name: "Test 4: Get nonexistent #2",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 12), nil),
			},
			needle:   net.NewPfx(strAddr("10.0.0.0"), 10),
			expected: nil,
		},
	}

	for _, test := range tests {
		rt := NewRoutingTable()
		for _, route := range test.routes {
			rt.AddPath(route.Prefix(), nil)
		}
		p := rt.Get(test.needle)

		if p == nil {
			if test.expected != nil {
				t.Errorf("Unexpected nil result for test %q", test.name)
			}
			continue
		}

		assert.Equal(t, test.expected, p)
	}
}
func TestGetLonger(t *testing.T) {
	tests := []struct {
		name     string
		routes   []*route.Route
		needle   net.Prefix
		expected []*route.Route
	}{
		{
			name: "Test 1: Search pfx and dump route + more specifics",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				route.NewRoute(net.NewPfx(strAddr("11.100.123.0"), 24), nil),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 12), nil),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 10), nil),
			},
			needle: net.NewPfx(strAddr("10.0.0.0"), 8),
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 10), nil),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 12), nil),
			},
		},
		{
			name:     "Test 2: Empty root",
			routes:   nil,
			needle:   net.NewPfx(strAddr("10.0.0.0"), 8),
			expected: []*route.Route{},
		},
	}

	for _, test := range tests {
		rt := NewRoutingTable()
		for _, route := range test.routes {
			rt.AddPath(route.Prefix(), nil)
		}
		p := rt.GetLonger(test.needle)

		if p == nil {
			if test.expected != nil {
				t.Errorf("Unexpected nil result for test %q", test.name)
			}
			continue
		}

		assert.Equal(t, test.expected, p)
	}
}
func TestLPM(t *testing.T) {
	tests := []struct {
		name     string
		routes   []*route.Route
		needle   net.Prefix
		expected []*route.Route
	}{
		{
			name:     "LPM for non-existent route",
			routes:   []*route.Route{},
			needle:   net.NewPfx(strAddr("10.0.0.0"), 32),
			expected: nil,
		},
		{
			name: "Positive LPM test",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				route.NewRoute(net.NewPfx(strAddr("11.100.123.0"), 24), nil),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 12), nil),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 10), nil),
			},
			needle: net.NewPfx(167772160, 32), // 10.0.0.0/32
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 10), nil),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 12), nil),
			},
		},
		{
			name: "Exact match",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				route.NewRoute(net.NewPfx(strAddr("11.100.123.0"), 24), nil),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 12), nil),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 10), nil),
			},
			needle: net.NewPfx(strAddr("10.0.0.0"), 10),
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 10), nil),
			},
		},
	}

	for _, test := range tests {
		rt := NewRoutingTable()
		for _, route := range test.routes {
			rt.AddPath(route.Prefix(), nil)
		}
		assert.Equal(t, test.expected, rt.LPM(test.needle))
	}
}

func TestRemovePath(t *testing.T) {
	tests := []struct {
		name       string
		routes     []*route.Route
		removePfx  net.Prefix
		removePath *route.Path
		expected   []*route.Route
	}{
		{
			name: "Remove a path that is the only one for a prefix",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), &route.Path{
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
				}),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 9), &route.Path{
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
				}),
				route.NewRoute(net.NewPfx(strAddr("10.128.0.0"), 9), &route.Path{
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
				}),
			},
			removePfx: net.NewPfx(strAddr("10.0.0.0"), 8),
			removePath: &route.Path{
				Type:    route.BGPPathType,
				BGPPath: &route.BGPPath{},
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 9), &route.Path{
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
				}),
				route.NewRoute(net.NewPfx(strAddr("10.128.0.0"), 9), &route.Path{
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
				}),
			},
		},
		{
			name: "Remove a path that is one of two for a prefix",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 1000,
					},
				}),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 2000,
					},
				}),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 9), &route.Path{
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
				}),
				route.NewRoute(net.NewPfx(strAddr("10.128.0.0"), 9), &route.Path{
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
				}),
			},
			removePfx: net.NewPfx(strAddr("10.0.0.0"), 8),
			removePath: &route.Path{
				Type: route.BGPPathType,
				BGPPath: &route.BGPPath{
					LocalPref: 1000,
				},
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 2000,
					},
				}),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 9), &route.Path{
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
				}),
				route.NewRoute(net.NewPfx(strAddr("10.128.0.0"), 9), &route.Path{
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
				}),
			},
		},
	}

	for _, test := range tests {
		rt := NewRoutingTable()
		for _, route := range test.routes {
			for _, p := range route.Paths() {
				rt.AddPath(route.Prefix(), p)
			}
		}

		rt.RemovePath(test.removePfx, test.removePath)

		rtExpected := NewRoutingTable()
		for _, route := range test.expected {
			for _, p := range route.Paths() {
				rtExpected.AddPath(route.Prefix(), p)
			}
		}

		assert.Equal(t, rtExpected.Dump(), rt.Dump())
	}
}

func TestReplacePath(t *testing.T) {
	tests := []struct {
		name        string
		routes      []*route.Route
		replacePfx  net.Prefix
		replacePath *route.Path
		expected    []*route.Route
		expectedOld []*route.Path
	}{
		{
			name:       "replace in empty table",
			replacePfx: net.NewPfx(strAddr("10.0.0.0"), 8),
			replacePath: &route.Path{
				Type:    route.BGPPathType,
				BGPPath: &route.BGPPath{},
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), &route.Path{
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
				}),
			},
			expectedOld: nil,
		},
		{
			name: "replace not existing prefix with multiple paths",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("11.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 1001,
						NextHop:   net.IPv4(101),
					},
				}),
				route.NewRoute(net.NewPfx(strAddr("11.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 1002,
						NextHop:   net.IPv4(100),
					},
				}),
			},
			replacePfx: net.NewPfx(strAddr("10.0.0.0"), 8),
			replacePath: &route.Path{
				Type: route.BGPPathType,
				BGPPath: &route.BGPPath{
					LocalPref: 1000,
				},
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 1000,
					},
				}),
				newMultiPathRoute(net.NewPfx(strAddr("11.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 1001,
						NextHop:   net.IPv4(101),
					},
				}, &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 1002,
						NextHop:   net.IPv4(100),
					},
				}),
			},
			expectedOld: []*route.Path{},
		},
		{
			name: "replace existing prefix with multiple paths",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 1,
					},
				}),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 2,
					},
				}),
				route.NewRoute(net.NewPfx(strAddr("11.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 1001,
						NextHop:   net.IPv4(101),
					},
				}),
				route.NewRoute(net.NewPfx(strAddr("11.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 1002,
						NextHop:   net.IPv4(100),
					},
				}),
			},
			replacePfx: net.NewPfx(strAddr("10.0.0.0"), 8),
			replacePath: &route.Path{
				Type: route.BGPPathType,
				BGPPath: &route.BGPPath{
					LocalPref: 1000,
				},
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 1000,
					},
				}),
				newMultiPathRoute(net.NewPfx(strAddr("11.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 1001,
						NextHop:   net.IPv4(101),
					},
				}, &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 1002,
						NextHop:   net.IPv4(100),
					},
				}),
			},
			expectedOld: []*route.Path{
				{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 1,
					},
				},
				{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 2,
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rt := NewRoutingTable()
			for _, route := range test.routes {
				for _, p := range route.Paths() {
					rt.AddPath(route.Prefix(), p)
				}
			}

			old := rt.ReplacePath(test.replacePfx, test.replacePath)
			assert.ElementsMatch(t, test.expectedOld, old)
			assert.ElementsMatch(t, test.expected, rt.Dump())
		})

	}
}

func TestRemovePrefix(t *testing.T) {
	tests := []struct {
		name        string
		routes      []*route.Route
		removePfx   net.Prefix
		expected    []*route.Route
		expectedOld []*route.Path
	}{
		{
			name:        "remove in empty table",
			removePfx:   net.NewPfx(strAddr("10.0.0.0"), 8),
			expected:    []*route.Route{},
			expectedOld: nil,
		},
		{
			name: "remove not exist prefix",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 1,
					},
				}),

				route.NewRoute(net.NewPfx(strAddr("11.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 1002,
						NextHop:   net.IPv4(100),
					},
				}),
			},
			removePfx: net.NewPfx(strAddr("12.0.0.0"), 8),
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 1,
					},
				}),

				route.NewRoute(net.NewPfx(strAddr("11.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 1002,
						NextHop:   net.IPv4(100),
					},
				}),
			},
			expectedOld: nil,
		},
		{
			name: "remove not existing more specific prefix",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 1,
					},
				}),

				route.NewRoute(net.NewPfx(strAddr("11.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 1002,
						NextHop:   net.IPv4(100),
					},
				}),
			},
			removePfx: net.NewPfx(strAddr("10.0.0.0"), 9),
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 1,
					},
				}),

				route.NewRoute(net.NewPfx(strAddr("11.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 1002,
						NextHop:   net.IPv4(100),
					},
				}),
			},
			expectedOld: nil,
		},
		{
			name: "remove not existing more less prefix",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 1,
					},
				}),

				route.NewRoute(net.NewPfx(strAddr("11.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 1002,
						NextHop:   net.IPv4(100),
					},
				}),
			},
			removePfx: net.NewPfx(strAddr("10.0.0.0"), 7),
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 1,
					},
				}),

				route.NewRoute(net.NewPfx(strAddr("11.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 1002,
						NextHop:   net.IPv4(100),
					},
				}),
			},
			expectedOld: nil,
		},
		{
			name: "remove existing prefix",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 1,
					},
				}),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 2,
					},
				}),
				route.NewRoute(net.NewPfx(strAddr("11.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 1002,
						NextHop:   net.IPv4(100),
					},
				}),
			},
			removePfx: net.NewPfx(strAddr("10.0.0.0"), 8),
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("11.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 1002,
						NextHop:   net.IPv4(100),
					},
				}),
			},
			expectedOld: []*route.Path{
				{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 1,
					},
				},
				{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 2,
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rt := NewRoutingTable()
			for _, route := range test.routes {
				for _, p := range route.Paths() {
					rt.AddPath(route.Prefix(), p)
				}
			}

			old := rt.RemovePfx(test.removePfx)
			assert.ElementsMatch(t, test.expectedOld, old)
			assert.ElementsMatch(t, test.expected, rt.Dump())
		})

	}
}

func newMultiPathRoute(pfx net.Prefix, paths ...*route.Path) *route.Route {
	if len(paths) == 0 {
		return route.NewRoute(pfx, nil)
	}
	r := route.NewRoute(pfx, paths[0])
	for i := 1; i < len(paths); i++ {
		r.AddPath(paths[i])
	}
	return r

}

func strAddr(s string) uint32 {
	ret, _ := net.StrToAddr(s)
	return ret
}
