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
	routerID := net.IPv4FromOctets(1, 1, 1, 1).Ptr().ToUint32()
	clusterID := net.IPv4FromOctets(2, 2, 2, 2).Ptr().ToUint32()

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
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
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
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
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
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 100,
							NextHop:   net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
							Source:    net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
						},
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 100,
							NextHop:   net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
							Source:    net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
						},
					},
				}),
			},
			removePfx: net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(),
			removePath: &route.Path{
				Type: route.BGPPathType,
				BGPPath: &route.BGPPath{
					BGPPathA: &route.BGPPathA{
						LocalPref: 100,
						NextHop:   net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
						Source:    net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
					},
				},
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref: 100,
							NextHop:   net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
							Source:    net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
						},
					},
				}),
			},
		},
		{
			name: "Add route with our RouterID as OriginatorID",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 32).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref:    111,
							OriginatorID: routerID,
						},
					},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 32).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							LocalPref:    111,
							OriginatorID: routerID,
						},
					},
					HiddenReason: route.HiddenReasonOurOriginatorID,
				}),
			},
		},
		{
			name: "Add route with our ClusterID within ClusterList",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 32).Ptr(), &route.Path{
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
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 32).Ptr(), &route.Path{
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

					HiddenReason: route.HiddenReasonClusterLoop,
				}),
			},
		},
		{
			name:    "Add route (with BGP add path)",
			addPath: true,
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						PathIdentifier: 10,
						BGPPathA: &route.BGPPathA{
							LocalPref: 100,
							NextHop:   net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
							Source:    net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
						},
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						PathIdentifier: 20,
						BGPPathA: &route.BGPPathA{
							LocalPref: 200,
							NextHop:   net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
							Source:    net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
						},
					},
				}),
			},
			removePfx:  nil,
			removePath: nil,
			expected: []*route.Route{
				route.NewRouteAddPath(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), []*route.Path{
					{
						Type: route.BGPPathType,
						BGPPath: &route.BGPPath{
							PathIdentifier: 10,
							BGPPathA: &route.BGPPathA{
								LocalPref: 100,
								NextHop:   net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
								Source:    net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
							},
						},
					},
					{
						Type: route.BGPPathType,
						BGPPath: &route.BGPPath{
							PathIdentifier: 20,
							BGPPathA: &route.BGPPathA{
								LocalPref: 200,
								NextHop:   net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
								Source:    net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
							},
						},
					},
				}),
			},
		},
		{
			name:    "Add route (with BGP add path) twice",
			addPath: false,
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						PathIdentifier: 10,
						BGPPathA: &route.BGPPathA{
							LocalPref: 100,
							NextHop:   net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
							Source:    net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
						},
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						PathIdentifier: 10,
						BGPPathA: &route.BGPPathA{
							LocalPref: 200,
							NextHop:   net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
							Source:    net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
						},
					},
				}),
			},
			removePfx:  nil,
			removePath: nil,
			expected: []*route.Route{
				route.NewRouteAddPath(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), []*route.Path{
					{
						Type: route.BGPPathType,
						BGPPath: &route.BGPPath{
							PathIdentifier: 10,
							BGPPathA: &route.BGPPathA{
								LocalPref: 200,
								NextHop:   net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
								Source:    net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
							},
						},
					},
				}),
			},
		},
	}

	for _, test := range tests {
		sessionAttrs := routingtable.SessionAttrs{
			RouterID:  routerID,
			ClusterID: clusterID,
			AddPathRX: test.addPath,
		}
		adjRIBIn := New(filter.NewAcceptAllFilterChain(), routingtable.NewContributingASNs(), sessionAttrs)
		mc := routingtable.NewRTMockClient()
		adjRIBIn.clientManager.RegisterWithOptions(mc, routingtable.ClientOptions{BestOnly: true})

		for _, route := range test.routes {
			adjRIBIn.AddPath(route.Prefix().Ptr(), route.Paths()[0])
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
		assert.Equal(t, test.expected, adjRIBIn.rt.Dump(), test.name)
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
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						PathIdentifier: 100,
						BGPPathA: &route.BGPPathA{
							NextHop: net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
							Source:  net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
						},
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						PathIdentifier: 200,
						BGPPathA: &route.BGPPathA{
							NextHop: net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
							Source:  net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
						},
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						PathIdentifier: 300,
						BGPPathA: &route.BGPPathA{
							NextHop: net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
							Source:  net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
						},
					},
				}),
			},
			removePfx: net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(),
			removePath: &route.Path{
				Type: route.BGPPathType,
				BGPPath: &route.BGPPath{
					PathIdentifier: 200,
					BGPPathA: &route.BGPPathA{
						NextHop: net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
						Source:  net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
					},
				},
			},
			expected: []*route.Route{
				route.NewRouteAddPath(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), []*route.Path{
					{
						Type: route.BGPPathType,
						BGPPath: &route.BGPPath{
							PathIdentifier: 100,
							BGPPathA: &route.BGPPathA{
								NextHop: net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
								Source:  net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
							},
						},
					},
					{
						Type: route.BGPPathType,
						BGPPath: &route.BGPPath{
							PathIdentifier: 300,
							BGPPathA: &route.BGPPathA{
								NextHop: net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
								Source:  net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
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
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop: net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
							Source:  net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
						},
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 9).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop: net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
							Source:  net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
						},
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 128, 0, 0), 9).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop: net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
							Source:  net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
						},
					},
				}),
			},
			removePfx: net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(),
			removePath: &route.Path{
				Type: route.BGPPathType,
				BGPPath: &route.BGPPath{
					BGPPathA: &route.BGPPathA{
						NextHop: net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
						Source:  net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
					},
				},
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 9).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop: net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
							Source:  net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
						},
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 128, 0, 0), 9).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop: net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
							Source:  net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
						},
					},
				}),
			},
			wantPropagation: true,
		},
		{
			name: "Remove non existing route",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 9).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop: net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
							Source:  net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
						},
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 128, 0, 0), 9).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop: net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
							Source:  net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
						},
					},
				}),
			},
			removePfx: net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(),
			removePath: &route.Path{
				Type: route.BGPPathType,
				BGPPath: &route.BGPPath{
					BGPPathA: &route.BGPPathA{
						NextHop: net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
						Source:  net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
					},
				},
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 9).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop: net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
							Source:  net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
						},
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 128, 0, 0), 9).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop: net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
							Source:  net.IPv4FromOctets(20, 0, 0, 0).Ptr(),
						},
					},
				}),
			},
			wantPropagation: false,
		},
	}

	for _, test := range tests {
		sessionAttrs := routingtable.SessionAttrs{
			RouterID:  1,
			ClusterID: 2,
			AddPathRX: test.addPath,
		}
		adjRIBIn := New(filter.NewAcceptAllFilterChain(), routingtable.NewContributingASNs(), sessionAttrs)
		for _, route := range test.routes {
			adjRIBIn.AddPath(route.Prefix().Ptr(), route.Paths()[0])
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
	adjRIBIn := New(filter.NewAcceptAllFilterChain(), routingtable.NewContributingASNs(), routingtable.SessionAttrs{})
	mc := routingtable.NewRTMockClient()
	adjRIBIn.Register(mc)

	pfxs := []*net.Prefix{
		net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 16).Ptr(),
		net.NewPfx(net.IPv4FromOctets(10, 0, 1, 0), 24).Ptr(),
	}

	paths := []*route.Path{
		{
			BGPPath: &route.BGPPath{
				BGPPathA: &route.BGPPathA{
					NextHop: net.IPv4FromOctets(192, 168, 0, 0).Ptr(),
				},
			},
		},
		{
			BGPPath: &route.BGPPath{
				BGPPathA: &route.BGPPathA{
					NextHop: net.IPv4FromOctets(192, 168, 2, 1).Ptr(),
				},
			},
		},
		{
			BGPPath: &route.BGPPath{
				BGPPathA: &route.BGPPathA{
					NextHop: net.IPv4FromOctets(192, 168, 3, 1).Ptr(),
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
	assert.Equal(t, &routingtable.RemovePathParams{Pfx: pfxs[0], Path: paths[0]}, r[0], "Withdraw 1")
	assert.Equal(t, &routingtable.RemovePathParams{Pfx: pfxs[0], Path: paths[1]}, r[1], "Withdraw 2")
	assert.Equal(t, &routingtable.RemovePathParams{Pfx: pfxs[1], Path: paths[2]}, r[2], "Withdraw 3")
}
