package rt

import (
	"testing"

	net "github.com/bio-routing/bio-rd/net"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	l := New()
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
				NewRoute(net.NewPfx(167772160, 9), []*Path{
					{
						Type:    BGPPathType,
						BGPPath: &BGPPath{},
					},
				}), // 10.0.0.0/9
				NewRoute(net.NewPfx(176160768, 9), []*Path{
					{
						Type:    BGPPathType,
						BGPPath: &BGPPath{},
					},
				}), // 10.128.0.0/9
			},
			remove: []*Route{
				NewRoute(net.NewPfx(167772160, 8), []*Path{
					{
						Type:    BGPPathType,
						BGPPath: &BGPPath{},
					},
				}), // 10.0.0.0
			},
			expected: []*Route{
				NewRoute(net.NewPfx(167772160, 9), []*Path{
					{
						Type:    BGPPathType,
						BGPPath: &BGPPath{},
					},
				}), // 10.0.0.0
				NewRoute(net.NewPfx(176160768, 9), []*Path{
					{
						Type:    BGPPathType,
						BGPPath: &BGPPath{},
					},
				}), // 10.128.0.0
			},
		},
		{
			name: "Remove a path that is one of two for a prefix",
			routes: []*Route{
				NewRoute(net.NewPfx(167772160, 8), []*Path{
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
				}), // 10.0.0.0/8
				NewRoute(net.NewPfx(167772160, 9), []*Path{
					{
						Type:    BGPPathType,
						BGPPath: &BGPPath{},
					},
				}), // 10.0.0.0/9
				NewRoute(net.NewPfx(176160768, 9), []*Path{
					{
						Type:    BGPPathType,
						BGPPath: &BGPPath{},
					},
				}), // 10.128.0.0/9
			},
			remove: []*Route{
				NewRoute(net.NewPfx(167772160, 8), []*Path{
					{
						Type: BGPPathType,
						BGPPath: &BGPPath{
							LocalPref: 1000,
						},
					},
				}), // 10.0.0.0
			},
			expected: []*Route{
				NewRoute(net.NewPfx(167772160, 8), []*Path{
					{
						Type: BGPPathType,
						BGPPath: &BGPPath{
							LocalPref: 2000,
						},
					},
				}), // 10.0.0.0/8
				NewRoute(net.NewPfx(167772160, 9), []*Path{
					{
						Type:    BGPPathType,
						BGPPath: &BGPPath{},
					},
				}), // 10.0.0.0/9
				NewRoute(net.NewPfx(176160768, 9), []*Path{
					{
						Type:    BGPPathType,
						BGPPath: &BGPPath{},
					},
				}), // 10.128.0.0/9
			},
		},
	}

	for _, test := range tests {
		lpm := New()
		for _, route := range test.routes {
			lpm.Insert(route)
		}

		for _, route := range test.remove {
			lpm.RemovePath(route)
		}

		res := lpm.Dump()
		assert.Equal(t, test.expected, res)
	}
}

func strAddr(s string) uint32 {
	ret, _ := net.StrToAddr(s)
	return ret
}
