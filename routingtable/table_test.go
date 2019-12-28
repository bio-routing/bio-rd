package routingtable

import (
	"testing"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/stretchr/testify/assert"
)

func TestAddPath(t *testing.T) {
	tests := []struct {
		name          string
		routes        []*route.Route
		expected      *node
		expectedCount int64
	}{
		{
			name: "10/8 and 11/8",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 0, 0, 0), 8).Ptr(), nil),
			},
			expected: &node{
				route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 7).Ptr(), nil),
				dummy: true,
				skip:  7,
				l: &node{
					route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
					skip:  0,
				},
				h: &node{
					route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 0, 0, 0), 8).Ptr(), nil),
					skip:  0,
				},
			},
			expectedCount: 2,
		},
		{
			name: "Insert first node",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
			},
			expected: &node{
				route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
				skip:  8,
			},
			expectedCount: 1,
		},
		{
			name: "Insert duplicate node",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
			},
			expected: &node{
				route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
				skip:  8,
			},
			expectedCount: 1,
		},
		{
			name: "Insert triangle",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 9).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 128, 0, 0), 9).Ptr(), nil),
			},
			expected: &node{
				route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
				skip:  8,
				l: &node{
					route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 9).Ptr(), nil),
				},
				h: &node{
					route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 128, 0, 0), 9).Ptr(), nil),
				},
			},
			expectedCount: 3,
		},
		{
			name: "Insert disjunct prefixes",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 100, 123, 0), 24).Ptr(), nil),
			},
			expected: &node{
				route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 7).Ptr(), nil),
				skip:  7,
				dummy: true,
				l: &node{
					route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
				},
				h: &node{
					route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 100, 123, 0), 24).Ptr(), nil),
					skip:  16,
				},
			},
			expectedCount: 2,
		},
		{
			name: "Insert disjunct prefixes plus one child low",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 100, 123, 0), 24).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 12).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 10).Ptr(), nil),
			},
			expected: &node{
				route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 7).Ptr(), nil),
				skip:  7,
				dummy: true,
				l: &node{
					route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
					l: &node{
						skip:  1,
						route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 10).Ptr(), nil),
						l: &node{
							skip:  1,
							route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 12).Ptr(), nil),
						},
					},
				},
				h: &node{
					route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 100, 123, 0), 24).Ptr(), nil),
					skip:  16,
				},
			},
			expectedCount: 4,
		},
		{
			name: "Insert disjunct prefixes plus one child high",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 100, 123, 0), 24).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 12).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 10).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 100, 123, 128), 25).Ptr(), nil),
			},
			expected: &node{
				route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 7).Ptr(), nil),
				skip:  7,
				dummy: true,
				l: &node{
					route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
					l: &node{
						skip:  1,
						route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 10).Ptr(), nil),
						l: &node{
							skip:  1,
							route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 12).Ptr(), nil),
						},
					},
				},
				h: &node{
					route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 100, 123, 0), 24).Ptr(), nil),
					skip:  16,
					h: &node{
						route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 100, 123, 128), 25).Ptr(), nil),
					},
				},
			},
			expectedCount: 5,
		},
		{
			name: "Insert disjunct prefixes plus one child high #2",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 12).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 10).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 100, 123, 128), 25).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 100, 123, 0), 24).Ptr(), nil),
			},
			expected: &node{
				route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 7).Ptr(), nil),
				skip:  7,
				dummy: true,
				l: &node{
					route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
					l: &node{
						skip:  1,
						route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 10).Ptr(), nil),
						l: &node{
							skip:  1,
							route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 12).Ptr(), nil),
						},
					},
				},
				h: &node{
					skip:  16,
					route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 100, 123, 0), 24).Ptr(), nil),
					h: &node{
						skip:  0,
						route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 100, 123, 128), 25).Ptr(), nil),
					},
				},
			},
			expectedCount: 5,
		},
		{
			name: "Insert disjunct prefixes plus one child high #3",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 100, 123, 0), 24).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 12).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 10).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 100, 123, 128), 25).Ptr(), nil),
			},
			expected: &node{
				route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 7).Ptr(), nil),
				skip:  7,
				dummy: true,
				l: &node{
					route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
					l: &node{
						skip:  1,
						route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 10).Ptr(), nil),
						l: &node{
							skip:  1,
							route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 12).Ptr(), nil),
						},
					},
				},
				h: &node{
					skip:  16,
					route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 100, 123, 0), 24).Ptr(), nil),
					h: &node{
						skip:  0,
						route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 100, 123, 128), 25).Ptr(), nil),
					},
				},
			},
			expectedCount: 5,
		},
		{
			name: "Insert triangle #2",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 128, 0, 0), 9).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 9).Ptr(), nil),
			},
			expected: &node{
				route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
				skip:  8,
				l: &node{
					route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 9).Ptr(), nil),
				},
				h: &node{
					route: route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 128, 0, 0), 9).Ptr(), nil),
				},
			},
			expectedCount: 3,
		},
	}

	for _, test := range tests {
		rt := NewRoutingTable()
		for _, route := range test.routes {
			rt.AddPath(route.Prefix(), nil)
		}

		assert.Equal(t, test.expected, rt.root, test.name)
		assert.Equal(t, test.expectedCount, rt.GetRouteCount(), test.name)
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name     string
		routes   []*route.Route
		needle   *net.Prefix
		expected *route.Route
	}{
		{
			name: "Test 1: Search pfx and dump route + more specifics",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 100, 123, 0), 24).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 12).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 10).Ptr(), nil),
			},
			needle:   net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(),
			expected: route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
		},
		{
			name: "Test 2: Search pfx and don't dump more specifics",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 100, 123, 0), 24).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 12).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 10).Ptr(), nil),
			},
			needle:   net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(),
			expected: route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
		},
		{
			name:     "Test 3: Empty table",
			routes:   []*route.Route{},
			needle:   net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 32).Ptr(),
			expected: nil,
		},
		{
			name: "Test 4: Get Dummy",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 100, 123, 0), 24).Ptr(), nil),
			},
			needle:   net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 7).Ptr(),
			expected: nil,
		},
		{
			name: "Test 5",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 100, 123, 0), 24).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 12).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 10).Ptr(), nil),
			},
			needle:   net.NewPfx(net.IPv4FromOctets(11, 100, 123, 0), 24).Ptr(),
			expected: route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 100, 123, 0), 24).Ptr(), nil),
		},
		{
			name: "Test 4: Get nonexistent #1",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 100, 123, 0), 24).Ptr(), nil),
			},
			needle:   net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 10).Ptr(),
			expected: nil,
		},
		{
			name: "Test 4: Get nonexistent #2",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 12).Ptr(), nil),
			},
			needle:   net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 10).Ptr(),
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
		needle   *net.Prefix
		expected []*route.Route
	}{
		{
			name: "Test 1: Search pfx and dump route + more specifics",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 100, 123, 0), 24).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 12).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 10).Ptr(), nil),
			},
			needle: net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(),
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 10).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 12).Ptr(), nil),
			},
		},
		{
			name:     "Test 2: Empty root",
			routes:   nil,
			needle:   net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(),
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
		needle   *net.Prefix
		expected []*route.Route
	}{
		{
			name:     "LPM for non-existent route",
			routes:   []*route.Route{},
			needle:   net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 32).Ptr(),
			expected: nil,
		},
		{
			name: "Positive LPM test",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 100, 123, 0), 24).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 12).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 10).Ptr(), nil),
			},
			needle: net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 32).Ptr(),
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 10).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 12).Ptr(), nil),
			},
		},
		{
			name: "Exact match",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 100, 123, 0), 24).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 12).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 10).Ptr(), nil),
			},
			needle: net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 10).Ptr(),
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), nil),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 10).Ptr(), nil),
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
		name          string
		routes        []*route.Route
		removePfx     *net.Prefix
		removePath    *route.Path
		expected      []*route.Route
		expectedCount int64
	}{
		{
			name: "Remove a path that is the only one for a prefix",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: route.NewBGPPathA(),
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 9).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: route.NewBGPPathA(),
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 128, 0, 0), 9).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: route.NewBGPPathA(),
					},
				}),
			},
			removePfx: net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(),
			removePath: &route.Path{
				Type: route.BGPPathType,
				BGPPath: &route.BGPPath{
					BGPPathA: route.NewBGPPathA(),
				},
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 9).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: route.NewBGPPathA(),
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 128, 0, 0), 9).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: route.NewBGPPathA(),
					},
				}),
			},
			expectedCount: 2,
		},
		{
			name: "Remove a path that is one of two for a prefix",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 1000,
							Source:    net.IPv4(0).Ptr(),
							NextHop:   net.IPv4(0).Ptr(),
						},
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 2000,
							Source:    net.IPv4(0).Ptr(),
							NextHop:   net.IPv4(0).Ptr(),
						},
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 9).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: route.NewBGPPathA(),
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 128, 0, 0), 9).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: route.NewBGPPathA(),
					},
				}),
			},
			removePfx: net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(),
			removePath: &route.Path{
				Type: route.BGPPathType,
				BGPPath: &route.BGPPath{
					BGPPathA: &route.BGPPathA{
						LocalPref: 1000,
						Source:    net.IPv4(0).Ptr(),
						NextHop:   net.IPv4(0).Ptr(),
					},
				},
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 2000,
							Source:    net.IPv4(0).Ptr(),
							NextHop:   net.IPv4(0).Ptr(),
						},
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 9).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: route.NewBGPPathA(),
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 128, 0, 0), 9).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: route.NewBGPPathA(),
					},
				}),
			},
			expectedCount: 3,
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
		assert.Equal(t, test.expectedCount, rt.GetRouteCount())
	}
}

func TestReplacePath(t *testing.T) {
	tests := []struct {
		name        string
		routes      []*route.Route
		replacePfx  *net.Prefix
		replacePath *route.Path
		expected    []*route.Route
		expectedOld []*route.Path
	}{
		/*{
			name:       "replace in empty table",
			replacePfx: net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(),
			replacePath: &route.Path{
				Type: route.BGPPathType,
				BGPPath: &route.BGPPath{
					BGPPathA: route.NewBGPPathA(),
				},
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: route.NewBGPPathA(),
					},
				}),
			},
			expectedOld: nil,
		},
		{
			name: "replace not existing prefix with multiple paths",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 1001,
							NextHop:   net.IPv4(101),
							Source:    net.IPv4(0).Ptr(),
						},
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 1002,
							NextHop:   net.IPv4(100),
							Source:    net.IPv4(0).Ptr(),
						},
					},
				}),
			},
			replacePfx: net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(),
			replacePath: &route.Path{
				Type: route.BGPPathType,
				BGPPath: &route.BGPPath{
					BGPPathA: &route.BGPPathA{
						LocalPref: 1000,
						NextHop:   net.IPv4(0).Ptr(),
						Source:    net.IPv4(0).Ptr(),
					},
				},
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 1000,
							NextHop:   net.IPv4(0).Ptr(),
							Source:    net.IPv4(0).Ptr(),
						},
					},
				}),
				newMultiPathRoute(net.NewPfx(net.IPv4FromOctets(11, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 1001,
							NextHop:   net.IPv4(101),
							Source:    net.IPv4(0).Ptr(),
						},
					},
				}, &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 1002,
							NextHop:   net.IPv4(100),
							Source:    net.IPv4(0).Ptr(),
						},
					},
				}),
			},
			expectedOld: []*route.Path{},
		},*/
		{
			name: "replace existing prefix with multiple paths",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 1,
							NextHop:   net.IPv4(0).Ptr(),
							Source:    net.IPv4(0).Ptr(),
						},
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 2,
							NextHop:   net.IPv4(0).Ptr(),
							Source:    net.IPv4(0).Ptr(),
						},
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 1001,
							NextHop:   net.IPv4(101).Ptr(),
							Source:    net.IPv4(0).Ptr(),
						},
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 1002,
							NextHop:   net.IPv4(102).Ptr(),
							Source:    net.IPv4(0).Ptr(),
						},
					},
				}),
			},
			replacePfx: net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(),
			replacePath: &route.Path{
				Type: route.BGPPathType,
				BGPPath: &route.BGPPath{
					BGPPathA: &route.BGPPathA{
						LocalPref: 1000,
						NextHop:   net.IPv4(0).Ptr(),
						Source:    net.IPv4(0).Ptr(),
					},
				},
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 1000,
							NextHop:   net.IPv4(0).Ptr(),
							Source:    net.IPv4(0).Ptr(),
						},
					},
				}),
				newMultiPathRoute(net.NewPfx(net.IPv4FromOctets(11, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 1001,
							NextHop:   net.IPv4(101).Ptr(),
							Source:    net.IPv4(0).Ptr(),
						},
					},
				}, &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 1002,
							NextHop:   net.IPv4(102).Ptr(),
							Source:    net.IPv4(0).Ptr(),
						},
					},
				}),
			},
			expectedOld: []*route.Path{
				{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 1,
							NextHop:   net.IPv4(0).Ptr(),
							Source:    net.IPv4(0).Ptr(),
						},
					},
				},
				{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 2,
							NextHop:   net.IPv4(0).Ptr(),
							Source:    net.IPv4(0).Ptr(),
						},
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
		removePfx   *net.Prefix
		expected    []*route.Route
		expectedOld []*route.Path
	}{
		{
			name:        "remove in empty table",
			removePfx:   net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(),
			expected:    []*route.Route{},
			expectedOld: nil,
		},
		{
			name: "remove not exist prefix",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 1,
						},
					},
				}),

				route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 1002,
							NextHop:   net.IPv4(100).Ptr(),
						},
					},
				}),
			},
			removePfx: net.NewPfx(net.IPv4FromOctets(12, 0, 0, 0), 8).Ptr(),
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 1,
						},
					},
				}),

				route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 1002,
							NextHop:   net.IPv4(100).Ptr(),
						},
					},
				}),
			},
			expectedOld: nil,
		},
		{
			name: "remove not existing more specific prefix",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 1,
						},
					},
				}),

				route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 1002,
							NextHop:   net.IPv4(100).Ptr(),
						},
					},
				}),
			},
			removePfx: net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 9).Ptr(),
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 1,
						},
					},
				}),

				route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 1002,
							NextHop:   net.IPv4(100).Ptr(),
						},
					},
				}),
			},
			expectedOld: nil,
		},
		{
			name: "remove not existing more less prefix",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 1,
						},
					},
				}),

				route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 1002,
							NextHop:   net.IPv4(100).Ptr(),
						},
					},
				}),
			},
			removePfx: net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 7).Ptr(),
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 1,
						},
					},
				}),

				route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 1002,
							NextHop:   net.IPv4(100).Ptr(),
						},
					},
				}),
			},
			expectedOld: nil,
		},
		{
			name: "remove existing prefix",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 1,
							Source:    net.IPv4(0).Ptr(),
							NextHop:   net.IPv4(0).Ptr(),
						},
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 2,
							Source:    net.IPv4(0).Ptr(),
							NextHop:   net.IPv4(0).Ptr(),
						},
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 1002,
							NextHop:   net.IPv4(100).Ptr(),
							Source:    net.IPv4(0).Ptr(),
						},
					},
				}),
			},
			removePfx: net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(),
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(11, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 1002,
							NextHop:   net.IPv4(100).Ptr(),
							Source:    net.IPv4(0).Ptr(),
						},
					},
				}),
			},
			expectedOld: []*route.Path{
				{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 1,
							Source:    net.IPv4(0).Ptr(),
							NextHop:   net.IPv4(0).Ptr(),
						},
					},
				},
				{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 2,
							Source:    net.IPv4(0).Ptr(),
							NextHop:   net.IPv4(0).Ptr(),
						},
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

func newMultiPathRoute(pfx *net.Prefix, paths ...*route.Path) *route.Route {
	if len(paths) == 0 {
		return route.NewRoute(pfx, nil)
	}
	r := route.NewRoute(pfx, paths[0])
	for i := 1; i < len(paths); i++ {
		r.AddPath(paths[i])
	}
	return r

}
