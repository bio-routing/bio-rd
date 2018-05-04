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

func strAddr(s string) uint32 {
	ret, _ := net.StrToAddr(s)
	return ret
}
