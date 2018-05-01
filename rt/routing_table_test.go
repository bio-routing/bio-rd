package rt

import (
	"testing"

	net "github.com/bio-routing/bio-rd/net"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	l := New(false)
	if l == nil {
		t.Errorf("New() returned nil")
	}
}

func TestRemovePath(t *testing.T) {
	tests := []struct {
		name     string
		routes   []*Route
		remove   []*Route
		expected []*Route
	}{
		{
			name: "Remove a path that is the only one for a prefix",
			routes: []*Route{
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), []*Path{
					{
						Type:    BGPPathType,
						BGPPath: &BGPPath{},
					},
				}),
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 9), []*Path{
					{
						Type:    BGPPathType,
						BGPPath: &BGPPath{},
					},
				}),
				NewRoute(net.NewPfx(strAddr("10.128.0.0"), 9), []*Path{
					{
						Type:    BGPPathType,
						BGPPath: &BGPPath{},
					},
				}),
			},
			remove: []*Route{
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), []*Path{
					{
						Type:    BGPPathType,
						BGPPath: &BGPPath{},
					},
				}),
			},
			expected: []*Route{
				{
					pfx: net.NewPfx(strAddr("10.0.0.0"), 9),
					paths: []*Path{
						{
							Type:    BGPPathType,
							BGPPath: &BGPPath{},
						},
					},
				},
				{
					pfx: net.NewPfx(strAddr("10.128.0.0"), 9),
					paths: []*Path{
						{
							Type:    BGPPathType,
							BGPPath: &BGPPath{},
						},
					},
				},
			},
		},
		/*{
			name: "Remove a path that is one of two for a prefix",
			routes: []*Route{
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), []*Path{
					{
						Type: BGPPathType,
						BGPPath: &BGPPath{
							LocalPref: 1000,
						},
					},
					{
						Type: BGPPathType,
						BGPPath: &BGPPath{
							LocalPref: 2000,
						},
					},
				}),
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 9), []*Path{
					{
						Type:    BGPPathType,
						BGPPath: &BGPPath{},
					},
				}),
				NewRoute(net.NewPfx(strAddr("10.128.0.0"), 9), []*Path{
					{
						Type:    BGPPathType,
						BGPPath: &BGPPath{},
					},
				}),
			},
			remove: []*Route{
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), []*Path{
					{
						Type: BGPPathType,
						BGPPath: &BGPPath{
							LocalPref: 1000,
						},
					},
				}),
			},
			expected: []*Route{
				{
					pfx: net.NewPfx(strAddr("10.0.0.0"), 8),
					paths: []*Path{
						{
							Type: BGPPathType,
							BGPPath: &BGPPath{
								LocalPref: 2000,
							},
						},
					},
					activePaths: []*Path{
						{
							Type: BGPPathType,
							BGPPath: &BGPPath{
								LocalPref: 2000,
							},
						},
					},
				},
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 9), []*Path{
					{
						Type:    BGPPathType,
						BGPPath: &BGPPath{},
					},
				}),
				NewRoute(net.NewPfx(strAddr("10.128.0.0"), 9), []*Path{
					{
						Type:    BGPPathType,
						BGPPath: &BGPPath{},
					},
				}),
			},
		},*/
	}

	for _, test := range tests {
		rt := New(false)
		for _, route := range test.routes {
			rt.AddPath(route)
		}

		for _, route := range test.remove {
			rt.RemovePath(route)
		}

		res := rt.Dump()
		assert.Equal(t, test.expected, res)
	}
}

func TestRemovePfx(t *testing.T) {
	tests := []struct {
		name     string
		routes   []*Route
		remove   []*net.Prefix
		expected []*Route
	}{
		{
			name: "Remove non-existent prefix",
			routes: []*Route{
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), []*Path{
					{
						Type:    BGPPathType,
						BGPPath: &BGPPath{},
					},
				}),
				NewRoute(net.NewPfx(strAddr("100.0.0.0"), 8), []*Path{
					{
						Type:    BGPPathType,
						BGPPath: &BGPPath{},
					},
				}),
			},
			remove: []*net.Prefix{
				net.NewPfx(0, 0),
			},
			expected: []*Route{
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), []*Path{
					{
						Type:    BGPPathType,
						BGPPath: &BGPPath{},
					},
				}),
				NewRoute(net.NewPfx(strAddr("100.0.0.0"), 8), []*Path{
					{
						Type:    BGPPathType,
						BGPPath: &BGPPath{},
					},
				}),
			},
		},
		{
			name: "Remove final prefix",
			routes: []*Route{
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), []*Path{
					{
						Type:    BGPPathType,
						BGPPath: &BGPPath{},
					},
				}),
			},
			remove: []*net.Prefix{
				net.NewPfx(strAddr("10.0.0.0"), 8),
			},
			expected: []*Route{},
		},
	}

	for _, test := range tests {
		lpm := New(false)
		for _, route := range test.routes {
			lpm.AddPath(route)
		}

		for _, pfx := range test.remove {
			lpm.RemoveRoute(pfx)
		}

		res := lpm.Dump()
		assert.Equal(t, test.expected, res)
	}
}

func TestGetBitUint32(t *testing.T) {
	tests := []struct {
		name     string
		input    uint32
		offset   uint8
		expected bool
	}{
		{
			name:     "test 1",
			input:    167772160, // 10.0.0.0
			offset:   8,
			expected: false,
		},
		{
			name:     "test 2",
			input:    184549376, // 11.0.0.0
			offset:   8,
			expected: true,
		},
	}

	for _, test := range tests {
		b := getBitUint32(test.input, test.offset)
		if b != test.expected {
			t.Errorf("%s: Unexpected failure: Bit %d of %d is %v. Expected %v", test.name, test.offset, test.input, b, test.expected)
		}
	}
}

func TestLPM(t *testing.T) {
	tests := []struct {
		name     string
		routes   []*Route
		needle   *net.Prefix
		expected []*Route
	}{
		{
			name:     "LPM for non-existent route",
			routes:   []*Route{},
			needle:   net.NewPfx(strAddr("10.0.0.0"), 32),
			expected: nil,
		},
		{
			name: "Positive LPM test",
			routes: []*Route{
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				NewRoute(net.NewPfx(strAddr("11.100.123.0"), 24), nil),
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 12), nil),
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 10), nil),
			},
			needle: net.NewPfx(167772160, 32), // 10.0.0.0/32
			expected: []*Route{
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 10), nil),
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 12), nil),
			},
		},
		/*{
			name: "Exact match",
			routes: []*Route{
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				NewRoute(net.NewPfx(strAddr("11.100.123.0"), 24), nil),
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 12), nil),
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 10), nil),
			},
			needle: net.NewPfx(strAddr("10.0.0.0"), 10),
			expected: []*Route{
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 10), nil),
			},
		},*/
	}

	for _, test := range tests {
		rt := New(false)
		for _, route := range test.routes {
			rt.AddPath(route)
		}
		assert.Equal(t, test.expected, rt.LPM(test.needle))
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name          string
		moreSpecifics bool
		routes        []*Route
		needle        *net.Prefix
		expected      []*Route
	}{
		{
			name:          "Test 1: Search pfx and dump route + more specifics",
			moreSpecifics: true,
			routes: []*Route{
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				NewRoute(net.NewPfx(strAddr("11.100.123.0"), 24), nil),
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 12), nil),
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 10), nil),
			},
			needle: net.NewPfx(strAddr("10.0.0.0"), 8),
			expected: []*Route{
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 10), nil),
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 12), nil),
			},
		},
		{
			name: "Test 2: Search pfx and don't dump more specifics",
			routes: []*Route{
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				NewRoute(net.NewPfx(strAddr("11.100.123.0"), 24), nil),
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 12), nil),
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 10), nil),
			},
			needle: net.NewPfx(strAddr("10.0.0.0"), 8),
			expected: []*Route{
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
			},
		},
		{
			name:     "Test 3: Empty table",
			routes:   []*Route{},
			needle:   net.NewPfx(strAddr("10.0.0.0"), 32),
			expected: nil,
		},
		{
			name: "Test 4: Get Dummy",
			routes: []*Route{
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				NewRoute(net.NewPfx(strAddr("11.100.123.0"), 24), nil),
			},
			needle:   net.NewPfx(strAddr("10.0.0.0"), 7),
			expected: nil,
		},
		{
			name: "Test 5",
			routes: []*Route{
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				NewRoute(net.NewPfx(strAddr("11.100.123.0"), 24), nil),
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 12), nil),
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 10), nil),
			},
			needle: net.NewPfx(strAddr("11.100.123.0"), 24),
			expected: []*Route{
				NewRoute(net.NewPfx(strAddr("11.100.123.0"), 24), nil),
			},
		},
		{
			name: "Test 4: Get nonexistent #1",
			routes: []*Route{
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				NewRoute(net.NewPfx(strAddr("11.100.123.0"), 24), nil),
			},
			needle:   net.NewPfx(strAddr("10.0.0.0"), 10),
			expected: nil,
		},
		{
			name: "Test 4: Get nonexistent #2",
			routes: []*Route{
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 12), nil),
			},
			needle:   net.NewPfx(strAddr("10.0.0.0"), 10),
			expected: nil,
		},
	}

	for _, test := range tests {
		rt := New(false)
		for _, route := range test.routes {
			rt.AddPath(route)
		}
		p := rt.Get(test.needle, test.moreSpecifics)

		if p == nil {
			if test.expected != nil {
				t.Errorf("Unexpected nil result for test %q", test.name)
			}
			continue
		}

		assert.Equal(t, test.expected, p)
	}
}

func TestInsert(t *testing.T) {
	tests := []struct {
		name     string
		routes   []*Route
		expected *node
	}{
		{
			name: "Insert first node",
			routes: []*Route{
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
			},
			expected: &node{
				route: &Route{
					pfx: net.NewPfx(strAddr("10.0.0.0"), 8),
				},
				skip: 8,
			},
		},
		{
			name: "Insert duplicate node",
			routes: []*Route{
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
				NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), nil),
			},
			expected: &node{
				route: &Route{
					pfx: net.NewPfx(strAddr("10.0.0.0"), 8),
				},
				skip: 8,
			},
		},
		/*{
			name: "Insert triangle",
			prefixes: []*net.Prefix{
				net.NewPfx(167772160, 8), // 10.0.0.0
				net.NewPfx(167772160, 9), // 10.0.0.0
				net.NewPfx(176160768, 9), // 10.128.0.0
			},
			expected: &node{
				route: &Route{
					pfx: net.NewPfx(167772160, 8), // 10.0.0.0/8
				},
				skip: 8,
				l: &node{
					route: &Route{
						pfx: net.NewPfx(167772160, 9), // 10.0.0.0
					},
				},
				h: &node{
					route: &Route{
						pfx: net.NewPfx(176160768, 9), // 10.128.0.0
					},
				},
			},
		},
		{
			name: "Insert disjunct prefixes",
			prefixes: []*net.Prefix{
				net.NewPfx(167772160, 8),  // 10.0.0.0
				net.NewPfx(191134464, 24), // 11.100.123.0/24
			},
			expected: &node{
				route: &Route{
					pfx: net.NewPfx(167772160, 7), // 10.0.0.0/7
				},
				skip:  7,
				dummy: true,
				l: &node{
					route: &Route{
						pfx: net.NewPfx(167772160, 8), // 10.0.0.0/8
					},
				},
				h: &node{
					route: &Route{
						pfx: net.NewPfx(191134464, 24), // 10.0.0.0/8
					},
					skip: 16,
				},
			},
		},
		{
			name: "Insert disjunct prefixes plus one child low",
			prefixes: []*net.Prefix{
				net.NewPfx(167772160, 8),  // 10.0.0.0
				net.NewPfx(191134464, 24), // 11.100.123.0/24
				net.NewPfx(167772160, 12), // 10.0.0.0
				net.NewPfx(167772160, 10), // 10.0.0.0
			},
			expected: &node{
				route: &Route{
					pfx: net.NewPfx(167772160, 7), // 10.0.0.0/7
				},
				skip:  7,
				dummy: true,
				l: &node{
					route: &Route{
						pfx: net.NewPfx(167772160, 8), // 10.0.0.0/8
					},
					l: &node{
						skip: 1,
						route: &Route{
							pfx: net.NewPfx(167772160, 10), // 10.0.0.0/10
						},
						l: &node{
							skip: 1,
							route: &Route{
								pfx: net.NewPfx(167772160, 12), // 10.0.0.0
							},
						},
					},
				},
				h: &node{
					route: &Route{
						pfx: net.NewPfx(191134464, 24), // 10.0.0.0/8
					},
					skip: 16,
				},
			},
		},
		{
			name: "Insert disjunct prefixes plus one child high",
			prefixes: []*net.Prefix{
				net.NewPfx(167772160, 8),  // 10.0.0.0
				net.NewPfx(191134464, 24), // 11.100.123.0/24
				net.NewPfx(167772160, 12), // 10.0.0.0
				net.NewPfx(167772160, 10), // 10.0.0.0
				net.NewPfx(191134592, 25), // 11.100.123.128/25
			},
			expected: &node{
				route: &Route{
					pfx: net.NewPfx(167772160, 7), // 10.0.0.0/7
				},
				skip:  7,
				dummy: true,
				l: &node{
					route: &Route{
						pfx: net.NewPfx(167772160, 8), // 10.0.0.0/8
					},
					l: &node{
						skip: 1,
						route: &Route{
							pfx: net.NewPfx(167772160, 10), // 10.0.0.0/10
						},
						l: &node{
							skip: 1,
							route: &Route{
								pfx: net.NewPfx(167772160, 12), // 10.0.0.0
							},
						},
					},
				},
				h: &node{
					route: &Route{
						pfx: net.NewPfx(191134464, 24), //11.100.123.0/24
					},
					skip: 16,
					h: &node{
						route: &Route{
							pfx: net.NewPfx(191134592, 25), //11.100.123.128/25
						},
					},
				},
			},
		},*/
	}

	for _, test := range tests {
		rt := New(false)
		for _, route := range test.routes {
			rt.AddPath(route)
		}

		assert.Equal(t, test.expected, rt.root)
	}
}

func strAddr(s string) uint32 {
	ret, _ := net.StrToAddr(s)
	return ret
}
