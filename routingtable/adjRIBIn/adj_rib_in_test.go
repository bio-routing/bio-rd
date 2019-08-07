package adjRIBIn

import (
	"testing"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/stretchr/testify/assert"
)

func TestAddPath(t *testing.T) {
	routerID := net.IPv4FromOctets(1, 1, 1, 1).ToUint32()
	clusterID := net.IPv4FromOctets(2, 2, 2, 2).ToUint32()

	tests := []struct {
		name       string
		addPath    bool
		routes     []*route.Route
		removePfx  *net.Prefix
		removePath *route.Path
		expected   []*route.Route
	}{
		{
			name: "Add route",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 100,
						},
					},
				}),
			},
			removePfx:  nil,
			removePath: nil,
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 100,
						},
					},
				}),
			},
		},
		{
			name: "Overwrite routes",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 100,
							NextHop:   net.IPv4FromOctets(20, 0, 0, 0),
							Source:    net.IPv4FromOctets(20, 0, 0, 0),
						},
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 100,
							NextHop:   net.IPv4FromOctets(20, 0, 0, 0),
							Source:    net.IPv4FromOctets(20, 0, 0, 0),
						},
					},
				}),
			},
			removePfx: net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8),
			removePath: &route.Path{
				Type: route.BGPPathType,
				BGPPath: &route.BGPPath{
					BGPPathA: &route.BGPPathA{
						LocalPref: 100,
						NextHop:   net.IPv4FromOctets(20, 0, 0, 0),
						Source:    net.IPv4FromOctets(20, 0, 0, 0),
					},
				},
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 100,
							NextHop:   net.IPv4FromOctets(20, 0, 0, 0),
							Source:    net.IPv4FromOctets(20, 0, 0, 0),
						},
					},
				}),
			},
		},
		{
			name: "Add route with our RouterID as OriginatorID",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 32), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref:    111,
							OriginatorID: routerID,
						},
					},
				}),
			},
			expected: []*route.Route{},
		},
		{
			name: "Add route with our ClusterID within ClusterList",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 32), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref:    222,
							OriginatorID: 23,
						},
						ClusterList: &types.ClusterList{
							clusterID,
						},
					},
				}),
			},
			expected: []*route.Route{},
		},
		{
			name:    "Add route (with BGP add path)",
			addPath: true,
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 100,
						},
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 200,
						},
					},
				}),
			},
			removePfx:  nil,
			removePath: nil,
			expected: []*route.Route{
				route.NewRouteAddPath(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), []*route.Path{
					{
						Type: route.BGPPathType,
						BGPPath: &route.BGPPath{
							BGPPathA: &route.BGPPathA{
								LocalPref: 100,
							},
						},
					},
					{
						Type: route.BGPPathType,
						BGPPath: &route.BGPPath{
							BGPPathA: &route.BGPPathA{
								LocalPref: 200,
							},
						},
					},
				}),
			},
		},
	}

	for _, test := range tests {
		adjRIBIn := New(filter.NewAcceptAllFilterChain(), routingtable.NewContributingASNs(), routerID, clusterID, test.addPath)
		mc := routingtable.NewRTMockClient()
		adjRIBIn.clientManager.RegisterWithOptions(mc, routingtable.ClientOptions{BestOnly: true})

		for _, route := range test.routes {
			adjRIBIn.AddPath(route.Prefix(), route.Paths()[0])
		}

		if test.removePath != nil {
			r := mc.Removed()
			assert.Equalf(t, 1, len(r), "Test %q failed: Call to RemovePath did not propagate prefix", test.name)

			removePathParams := r[0]
			if !removePathParams.Pfx.Equal(test.removePfx) {
				t.Errorf("Test %q failed: Call to RemovePath did not propagate prefix properly: Got: %s Want: %s", test.name, removePathParams.Pfx.String(), test.removePfx.String())
			}

			assert.Equal(t, test.removePath.Equal(removePathParams.Path), true, test.name)
		}
		assert.Equal(t, test.expected, adjRIBIn.rt.Dump())
	}
}

func TestRemovePath(t *testing.T) {
	tests := []struct {
		name            string
		addPath         bool
		routes          []*route.Route
		removePfx       *net.Prefix
		removePath      *route.Path
		expected        []*route.Route
		wantPropagation bool
	}{
		{
			name:    "Remove an a path from existing route",
			addPath: true,
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						PathIdentifier: 100,
						BGPPathA: &route.BGPPathA{
							NextHop: net.IPv4FromOctets(20, 0, 0, 0),
							Source:  net.IPv4FromOctets(20, 0, 0, 0),
						},
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						PathIdentifier: 200,
						BGPPathA: &route.BGPPathA{
							NextHop: net.IPv4FromOctets(20, 0, 0, 0),
							Source:  net.IPv4FromOctets(20, 0, 0, 0),
						},
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						PathIdentifier: 300,
						BGPPathA: &route.BGPPathA{
							NextHop: net.IPv4FromOctets(20, 0, 0, 0),
							Source:  net.IPv4FromOctets(20, 0, 0, 0),
						},
					},
				}),
			},
			removePfx: net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8),
			removePath: &route.Path{
				Type: route.BGPPathType,
				BGPPath: &route.BGPPath{
					PathIdentifier: 200,
					BGPPathA: &route.BGPPathA{
						NextHop: net.IPv4FromOctets(20, 0, 0, 0),
						Source:  net.IPv4FromOctets(20, 0, 0, 0),
					},
				},
			},
			expected: []*route.Route{
				route.NewRouteAddPath(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), []*route.Path{
					{
						Type: route.BGPPathType,
						BGPPath: &route.BGPPath{
							PathIdentifier: 100,
							BGPPathA: &route.BGPPathA{
								NextHop: net.IPv4FromOctets(20, 0, 0, 0),
								Source:  net.IPv4FromOctets(20, 0, 0, 0),
							},
						},
					},
					{
						Type: route.BGPPathType,
						BGPPath: &route.BGPPath{
							PathIdentifier: 300,
							BGPPathA: &route.BGPPathA{
								NextHop: net.IPv4FromOctets(20, 0, 0, 0),
								Source:  net.IPv4FromOctets(20, 0, 0, 0),
							},
						},
					},
				}),
			},
			wantPropagation: true,
		},
		{
			name: "Remove an existing route",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop: net.IPv4FromOctets(20, 0, 0, 0),
							Source:  net.IPv4FromOctets(20, 0, 0, 0),
						},
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 9), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop: net.IPv4FromOctets(20, 0, 0, 0),
							Source:  net.IPv4FromOctets(20, 0, 0, 0),
						},
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 128, 0, 0), 9), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop: net.IPv4FromOctets(20, 0, 0, 0),
							Source:  net.IPv4FromOctets(20, 0, 0, 0),
						},
					},
				}),
			},
			removePfx: net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8),
			removePath: &route.Path{
				Type: route.BGPPathType,
				BGPPath: &route.BGPPath{
					BGPPathA: &route.BGPPathA{
						NextHop: net.IPv4FromOctets(20, 0, 0, 0),
						Source:  net.IPv4FromOctets(20, 0, 0, 0),
					},
				},
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 9), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop: net.IPv4FromOctets(20, 0, 0, 0),
							Source:  net.IPv4FromOctets(20, 0, 0, 0),
						},
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 128, 0, 0), 9), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop: net.IPv4FromOctets(20, 0, 0, 0),
							Source:  net.IPv4FromOctets(20, 0, 0, 0),
						},
					},
				}),
			},
			wantPropagation: true,
		},
		{
			name: "Remove non existing route",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 9), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop: net.IPv4FromOctets(20, 0, 0, 0),
							Source:  net.IPv4FromOctets(20, 0, 0, 0),
						},
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 128, 0, 0), 9), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop: net.IPv4FromOctets(20, 0, 0, 0),
							Source:  net.IPv4FromOctets(20, 0, 0, 0),
						},
					},
				}),
			},
			removePfx: net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8),
			removePath: &route.Path{
				Type: route.BGPPathType,
				BGPPath: &route.BGPPath{
					BGPPathA: &route.BGPPathA{
						NextHop: net.IPv4FromOctets(20, 0, 0, 0),
						Source:  net.IPv4FromOctets(20, 0, 0, 0),
					},
				},
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 9), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop: net.IPv4FromOctets(20, 0, 0, 0),
							Source:  net.IPv4FromOctets(20, 0, 0, 0),
						},
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 128, 0, 0), 9), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop: net.IPv4FromOctets(20, 0, 0, 0),
							Source:  net.IPv4FromOctets(20, 0, 0, 0),
						},
					},
				}),
			},
			wantPropagation: false,
		},
	}

	for _, test := range tests {
		adjRIBIn := New(filter.NewAcceptAllFilterChain(), routingtable.NewContributingASNs(), 1, 2, test.addPath)
		for _, route := range test.routes {
			adjRIBIn.AddPath(route.Prefix(), route.Paths()[0])
		}

		mc := routingtable.NewRTMockClient()
		adjRIBIn.clientManager.RegisterWithOptions(mc, routingtable.ClientOptions{})
		adjRIBIn.RemovePath(test.removePfx, test.removePath)

		if test.wantPropagation {
			r := mc.Removed()
			assert.Equalf(t, 1, len(r), "Test %q failed: Call to RemovePath did not propagate prefix", test.name)

			removePathParams := r[0]
			if removePathParams.Pfx != test.removePfx {
				t.Errorf("Test %q failed: Call to RemovePath did not propagate prefix properly: Got: %s Want: %s", test.name, removePathParams.Pfx.String(), test.removePfx.String())
			}
			assert.Equal(t, test.removePath.Equal(removePathParams.Path), true)
		} else {
			r := mc.Removed()
			assert.Equalf(t, 0, len(r), "Test %q failed: Call to RemovePath propagated unexpectedly", test.name)
		}

		assert.Equal(t, test.expected, adjRIBIn.rt.Dump())
	}
}

func TestUnregister(t *testing.T) {
	adjRIBIn := New(filter.NewAcceptAllFilterChain(), routingtable.NewContributingASNs(), 0, 0, false)
	mc := routingtable.NewRTMockClient()
	adjRIBIn.Register(mc)

	pfxs := []*net.Prefix{
		net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 16),
		net.NewPfx(net.IPv4FromOctets(10, 0, 1, 0), 24),
	}

	paths := []*route.Path{
		{
			BGPPath: &route.BGPPath{
				BGPPathA: &route.BGPPathA{
					NextHop: net.IPv4FromOctets(192, 168, 0, 0),
				},
			},
		},
		{
			BGPPath: &route.BGPPath{
				BGPPathA: &route.BGPPathA{
					NextHop: net.IPv4FromOctets(192, 168, 2, 1),
				},
			},
		},
		{
			BGPPath: &route.BGPPath{
				BGPPathA: &route.BGPPathA{
					NextHop: net.IPv4FromOctets(192, 168, 3, 1),
				},
			},
		},
	}

	adjRIBIn.RT().AddPath(pfxs[0], paths[0])
	adjRIBIn.RT().AddPath(pfxs[0], paths[1])
	adjRIBIn.RT().AddPath(pfxs[1], paths[2])

	adjRIBIn.Unregister(mc)

	r := mc.Removed()
	assert.Equalf(t, 3, len(r), "Should have removed 3 paths, but only removed %d", len(r))
	assert.Equal(t, &routingtable.RemovePathParams{pfxs[0], paths[0]}, r[0], "Withdraw 1")
	assert.Equal(t, &routingtable.RemovePathParams{pfxs[0], paths[1]}, r[1], "Withdraw 2")
	assert.Equal(t, &routingtable.RemovePathParams{pfxs[1], paths[2]}, r[2], "Withdraw 3")
}
