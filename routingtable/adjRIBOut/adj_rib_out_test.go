package adjRIBOut

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/routingtable/filter"

	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
)

func TestBestPathOnlyEBGP(t *testing.T) {
	neighborBestOnlyEBGP := &routingtable.Neighbor{
		Type:              route.BGPPathType,
		LocalAddress:      net.IPv4FromOctets(127, 0, 0, 1),
		Address:           net.IPv4FromOctets(127, 0, 0, 2),
		IBGP:              false,
		LocalASN:          41981,
		RouteServerClient: false,
	}

	adjRIBOut := New(nil, neighborBestOnlyEBGP, filter.NewAcceptAllFilterChain(), false)

	tests := []struct {
		name          string
		routesAdd     []*route.Route
		routesRemove  []*route.Route
		expected      []*route.Route
		expectedCount int64
	}{
		{
			name: "Add a valid route",
			routesAdd: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							Source: net.IPv4(0),
						},
						ASPath: &types.ASPath{},
					},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop:   neighborBestOnlyEBGP.LocalAddress,
							Origin:    0,
							MED:       0,
							EBGP:      false,
							LocalPref: 0,
							Source:    net.IPv4(0),
						},
						ASPath: &types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									neighborBestOnlyEBGP.LocalASN,
								},
							},
						},
						ASPathLen:         1,
						UnknownAttributes: nil,
						PathIdentifier:    0,
					},
				}),
			},
			expectedCount: 1,
		},
		{
			name: "Try to remove unpresent route",
			routesRemove: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop:   neighborBestOnlyEBGP.LocalAddress,
							Origin:    0,
							MED:       1,
							EBGP:      false,
							LocalPref: 0,
							Source:    net.IPv4(0),
						},
						ASPath: &types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									neighborBestOnlyEBGP.LocalASN,
								},
							},
						},
						ASPathLen:         1,
						Communities:       &types.Communities{},
						LargeCommunities:  &types.LargeCommunities{},
						UnknownAttributes: nil,
						PathIdentifier:    0,
					},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop:   neighborBestOnlyEBGP.LocalAddress,
							Origin:    0,
							MED:       0,
							EBGP:      false,
							LocalPref: 0,
							Source:    net.IPv4(0),
						},
						ASPath: &types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									neighborBestOnlyEBGP.LocalASN,
								},
							},
						},
						ASPathLen:         1,
						UnknownAttributes: nil,
						PathIdentifier:    0,
					},
				}),
			},
			expectedCount: 1,
		},
		{
			name: "Remove route added in first step",
			routesRemove: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop:   neighborBestOnlyEBGP.LocalAddress,
							Origin:    0,
							MED:       0,
							EBGP:      false,
							LocalPref: 0,
							Source:    net.IPv4(0),
						},
						ASPath: &types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									neighborBestOnlyEBGP.LocalASN,
								},
							},
						},
						ASPathLen:         1,
						UnknownAttributes: nil,
						PathIdentifier:    0,
					},
				}),
			},
			expected:      []*route.Route{},
			expectedCount: 0,
		},
		{
			name: "Try to add route with NO_EXPORT community set",
			routesAdd: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							Source:  net.IPv4(0),
							NextHop: net.IPv4(0),
						},
						Communities: &types.Communities{
							types.WellKnownCommunityNoExport,
						},
					},
				}),
			},
			expected: []*route.Route{},
		},
		{
			name: "Try to add route with NO_ADVERTISE community set",
			routesAdd: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							Source:  net.IPv4(0),
							NextHop: net.IPv4(0),
						},
						Communities: &types.Communities{
							types.WellKnownCommunityNoAdvertise,
						},
					},
				}),
			},
			expected:      []*route.Route{},
			expectedCount: 0,
		},
		{
			name: "Re-add valid route again",
			routesAdd: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							Source:  net.IPv4(0),
							NextHop: net.IPv4(0),
						},
						ASPath: &types.ASPath{},
					},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop:   neighborBestOnlyEBGP.LocalAddress,
							Origin:    0,
							MED:       0,
							EBGP:      false,
							LocalPref: 0,
							Source:    net.IPv4(0),
						},
						ASPath: &types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									neighborBestOnlyEBGP.LocalASN,
								},
							},
						},
						ASPathLen:         1,
						UnknownAttributes: nil,
						PathIdentifier:    0,
					},
				}),
			},
			expectedCount: 1,
		},
		{
			name: "Try to remove route with NO_EXPORT community set",
			routesRemove: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop:   neighborBestOnlyEBGP.LocalAddress,
							Origin:    0,
							MED:       0,
							EBGP:      false,
							LocalPref: 0,
							Source:    net.IPv4(0),
						},
						ASPath: &types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									neighborBestOnlyEBGP.LocalASN,
								},
							},
						},
						ASPathLen: 1,
						Communities: &types.Communities{
							types.WellKnownCommunityNoExport,
						},
						LargeCommunities:  &types.LargeCommunities{},
						UnknownAttributes: nil,
						PathIdentifier:    0,
					},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop:   neighborBestOnlyEBGP.LocalAddress,
							Origin:    0,
							MED:       0,
							EBGP:      false,
							LocalPref: 0,
							Source:    net.IPv4(0),
						},
						ASPath: &types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									neighborBestOnlyEBGP.LocalASN,
								},
							},
						},
						ASPathLen:         1,
						UnknownAttributes: nil,
						PathIdentifier:    0,
					},
				}),
			},
			expectedCount: 1,
		},
		{
			name: "Try to remove non-existent prefix",
			routesRemove: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 23, 42, 0), 24), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop: net.IPv4(0),
							Source:  net.IPv4(0),
						},
					},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop:   neighborBestOnlyEBGP.LocalAddress,
							Origin:    0,
							MED:       0,
							EBGP:      false,
							LocalPref: 0,
							Source:    net.IPv4(0),
						},
						ASPath: &types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									neighborBestOnlyEBGP.LocalASN,
								},
							},
						},
						ASPathLen:         1,
						UnknownAttributes: nil,
						PathIdentifier:    0,
					},
				}),
			},
			expectedCount: 1,
		},
	}

	for i, test := range tests {
		fmt.Printf("Running eBGP best only test #%d: %s\n", i+1, test.name)
		for _, route := range test.routesAdd {
			adjRIBOut.AddPath(route.Prefix(), route.Paths()[0])
		}

		for _, route := range test.routesRemove {
			adjRIBOut.RemovePath(route.Prefix(), route.Paths()[0])
		}

		assert.Equal(t, test.expected, adjRIBOut.rt.Dump())

		actualCount := adjRIBOut.RouteCount()
		if test.expectedCount != actualCount {
			t.Errorf("Expected route count %d differs from actual route count %d!\n", test.expectedCount, actualCount)
		}
	}
}

func TestBestPathOnlyIBGP(t *testing.T) {
	neighborBestOnlyEBGP := &routingtable.Neighbor{
		Type:              route.BGPPathType,
		LocalAddress:      net.IPv4FromOctets(127, 0, 0, 1),
		Address:           net.IPv4FromOctets(127, 0, 0, 2),
		IBGP:              true,
		LocalASN:          41981,
		RouteServerClient: false,
	}

	adjRIBOut := New(nil, neighborBestOnlyEBGP, filter.NewAcceptAllFilterChain(), false)

	testSteps := []struct {
		name          string
		routesAdd     []*route.Route
		routesRemove  []*route.Route
		expected      []*route.Route
		expectedCount int64
	}{
		{
			name: "Add an iBGP route (without success)",
			routesAdd: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							Source:  net.IPv4(0),
							NextHop: net.IPv4(0),
						},
					},
				}),
			},
			expected:      []*route.Route{},
			expectedCount: 0,
		},
		{
			name: "Add an eBGP route (with success)",
			routesAdd: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							EBGP:    true,
							NextHop: net.IPv4FromOctets(1, 2, 3, 4),
							Source:  net.IPv4(0),
						},
						ASPath: &types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen: 1,
					},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop:   net.IPv4FromOctets(1, 2, 3, 4),
							Origin:    0,
							MED:       0,
							EBGP:      true,
							LocalPref: 0,
							Source:    net.IPv4(0),
						},
						ASPath: &types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen:         1,
						UnknownAttributes: nil,
						PathIdentifier:    0,
					},
				}),
			},
			expectedCount: 1,
		},
		{
			name: "Try to remove slightly different route",
			routesRemove: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop:   net.IPv4FromOctets(1, 2, 3, 4),
							Origin:    0,
							MED:       1,
							EBGP:      true,
							LocalPref: 0,
							Source:    net.IPv4(0),
						},
						ASPath: &types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen:         1,
						Communities:       &types.Communities{},
						LargeCommunities:  &types.LargeCommunities{},
						UnknownAttributes: nil,
						PathIdentifier:    0,
					},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop:   net.IPv4FromOctets(1, 2, 3, 4),
							Origin:    0,
							MED:       0,
							EBGP:      true,
							LocalPref: 0,
							Source:    net.IPv4(0),
						},
						ASPath: &types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen:         1,
						UnknownAttributes: nil,
						PathIdentifier:    0,
					},
				}),
			},
			expectedCount: 1,
		},
		{
			name: "Remove route added in 2nd step",
			routesRemove: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							EBGP:    true,
							NextHop: net.IPv4FromOctets(1, 2, 3, 4),
							Source:  net.IPv4(0),
						},
						ASPath: &types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen: 1,
					},
				}),
			},
			expected:      []*route.Route{},
			expectedCount: 0,
		},
		{
			name: "Try to add route with NO_EXPORT community set (without success)",
			routesAdd: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							Source:  net.IPv4(0),
							NextHop: net.IPv4(0),
						},
						Communities: &types.Communities{
							types.WellKnownCommunityNoExport,
						},
					},
				}),
			},
			expected: []*route.Route{},
		},
		{
			name: "Try to add route with NO_ADVERTISE community set (without success)",
			routesAdd: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							Source:  net.IPv4(0),
							NextHop: net.IPv4(0),
						},
						Communities: &types.Communities{
							types.WellKnownCommunityNoAdvertise,
						},
					},
				}),
			},
			expected: []*route.Route{},
		},
	}

	for i, test := range testSteps {
		fmt.Printf("Running iBGP best only test #%d: %s\n", i+1, test.name)
		for _, route := range test.routesAdd {
			adjRIBOut.AddPath(route.Prefix(), route.Paths()[0])
		}

		for _, route := range test.routesRemove {
			adjRIBOut.RemovePath(route.Prefix(), route.Paths()[0])
		}

		assert.Equal(t, test.expected, adjRIBOut.rt.Dump())

		actualCount := adjRIBOut.RouteCount()
		if test.expectedCount != actualCount {
			t.Errorf("Expected route count %d differs from actual route count %d!\n", test.expectedCount, actualCount)
		}
	}
}

/*
 * Test for iBGP Route Reflector client neighbor
 */

func TestBestPathOnlyRRClient(t *testing.T) {
	neighborBestOnlyRR := &routingtable.Neighbor{
		Type:                 route.BGPPathType,
		LocalAddress:         net.IPv4FromOctets(127, 0, 0, 1),
		Address:              net.IPv4FromOctets(127, 0, 0, 2),
		IBGP:                 true,
		LocalASN:             41981,
		RouteServerClient:    false,
		RouteReflectorClient: true,
		ClusterID:            net.IPv4FromOctets(2, 2, 2, 2).ToUint32(),
	}

	adjRIBOut := New(nil, neighborBestOnlyRR, filter.NewAcceptAllFilterChain(), false)

	tests := []struct {
		name          string
		routesAdd     []*route.Route
		routesRemove  []*route.Route
		expected      []*route.Route
		expectedCount int64
	}{
		{
			name: "Add an iBGP route (with success)",
			routesAdd: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							Source:  net.IPv4(0),
							NextHop: net.IPv4(0),
						},
						ASPath: &types.ASPath{},
					},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							Source:  net.IPv4(0),
							NextHop: net.IPv4(0),
						},
						ASPath: &types.ASPath{},
						ClusterList: &types.ClusterList{
							neighborBestOnlyRR.ClusterID,
						},
					},
				}),
			},
			expectedCount: 1,
		},
		{
			name: "Add an eBGP route (replacing previous iBGP route)",
			routesAdd: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							EBGP:    true,
							NextHop: net.IPv4FromOctets(1, 2, 3, 4),
							Source:  net.IPv4(0),
						},
						ASPath: &types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen: 1,
					},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop:   net.IPv4FromOctets(1, 2, 3, 4),
							Origin:    0,
							MED:       0,
							EBGP:      true,
							LocalPref: 0,
							Source:    net.IPv4(0),
						},
						ASPath: &types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen:         1,
						UnknownAttributes: nil,
						PathIdentifier:    0,
						ClusterList: &types.ClusterList{
							neighborBestOnlyRR.ClusterID,
						},
					},
				}),
			},
			expectedCount: 1,
		},
		{
			name: "Try to remove slightly different route",
			routesRemove: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop:   net.IPv4FromOctets(1, 2, 3, 4),
							Origin:    0,
							MED:       1, // Existing route has MED 0
							EBGP:      true,
							LocalPref: 0,
							Source:    net.IPv4(0),
						},
						ASPath: &types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen:         1,
						Communities:       &types.Communities{},
						LargeCommunities:  &types.LargeCommunities{},
						UnknownAttributes: nil,
						PathIdentifier:    0,
					},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop:   net.IPv4FromOctets(1, 2, 3, 4),
							Origin:    0,
							MED:       0,
							EBGP:      true,
							LocalPref: 0,
							Source:    net.IPv4(0),
						},
						ASPath: &types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen:         1,
						UnknownAttributes: nil,
						PathIdentifier:    0,
						ClusterList: &types.ClusterList{
							neighborBestOnlyRR.ClusterID,
						},
					},
				}),
			},
			expectedCount: 1,
		},
		{
			name: "Remove route added in 2nd step",
			routesRemove: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop:   net.IPv4FromOctets(1, 2, 3, 4),
							Origin:    0,
							MED:       0,
							EBGP:      true,
							LocalPref: 0,
							Source:    net.IPv4(0),
						},
						ASPath: &types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen:         1,
						UnknownAttributes: nil,
						PathIdentifier:    0,
						ClusterList: &types.ClusterList{
							neighborBestOnlyRR.ClusterID,
						},
					},
				}),
			},
			expected:      []*route.Route{},
			expectedCount: 0,
		},
		{
			name: "Try to add route with NO_ADVERTISE community set (without success)",
			routesAdd: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							Source:  net.IPv4(0),
							NextHop: net.IPv4(0),
						},
						Communities: &types.Communities{
							types.WellKnownCommunityNoAdvertise,
						},
					},
				}),
			},
			expected:      []*route.Route{},
			expectedCount: 0,
		},
		{
			name: "Try to add route with NO_EXPORT community set (with success)",
			routesAdd: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							Source:  net.IPv4(0),
							NextHop: net.IPv4(0),
						},
						Communities: &types.Communities{
							types.WellKnownCommunityNoExport,
						},
					},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							Origin:    0,
							MED:       0,
							EBGP:      false,
							LocalPref: 0,
							Source:    net.IPv4(0),
							NextHop:   net.IPv4(0),
						},
						ASPathLen: 0,
						Communities: &types.Communities{
							types.WellKnownCommunityNoExport,
						},
						PathIdentifier: 0,
						ClusterList: &types.ClusterList{
							neighborBestOnlyRR.ClusterID,
						},
					},
				}),
			},
			expectedCount: 1,
		},
		{
			name: "Remove NO_EXPORT route added before",
			routesRemove: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							Origin:    0,
							MED:       0,
							EBGP:      false,
							LocalPref: 0,
							Source:    net.IPv4(0),
							NextHop:   net.IPv4(0),
						},
						ASPathLen: 0,
						Communities: &types.Communities{
							types.WellKnownCommunityNoExport,
						},
						PathIdentifier: 0,
						ClusterList: &types.ClusterList{
							neighborBestOnlyRR.ClusterID,
						},
					},
				}),
			},
			expected:      []*route.Route{},
			expectedCount: 0,
		},
		{
			name: "Add route with one entry in ClusterList and OriginatorID set (with success)",
			routesAdd: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							OriginatorID: 42,
							Source:       net.IPv4(0),
							NextHop:      net.IPv4(0),
						},
						ClusterList: &types.ClusterList{
							23,
						},
					},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						ASPathLen: 0,
						BGPPathA: &route.BGPPathA{
							Origin:       0,
							MED:          0,
							EBGP:         false,
							LocalPref:    0,
							Source:       net.IPv4(0),
							OriginatorID: 42,
							NextHop:      net.IPv4(0),
						},
						UnknownAttributes: nil,
						PathIdentifier:    0,
						ClusterList: &types.ClusterList{
							neighborBestOnlyRR.ClusterID,
							23,
						},
					},
				}),
			},
			expectedCount: 1,
		},
	}

	for i, test := range tests {
		fmt.Printf("Running RR client best only test #%d: %s\n", i+1, test.name)
		for _, route := range test.routesAdd {
			adjRIBOut.AddPath(route.Prefix(), route.Paths()[0])
		}

		for _, route := range test.routesRemove {
			adjRIBOut.RemovePath(route.Prefix(), route.Paths()[0])
		}

		assert.Equal(t, test.expected, adjRIBOut.rt.Dump())

		actualCount := adjRIBOut.RouteCount()
		if test.expectedCount != actualCount {
			t.Errorf("Expected route count %d differs from actual route count %d!\n", test.expectedCount, actualCount)
		}
	}
}

/*
 * Test for AddPath capable peer / AdjRIBOut
 */

func TestAddPathIBGP(t *testing.T) {
	neighborBestOnlyEBGP := &routingtable.Neighbor{
		Type:              route.BGPPathType,
		LocalAddress:      net.IPv4FromOctets(127, 0, 0, 1),
		Address:           net.IPv4FromOctets(127, 0, 0, 2),
		IBGP:              true,
		LocalASN:          41981,
		RouteServerClient: false,
	}

	adjRIBOut := New(nil, neighborBestOnlyEBGP, filter.NewAcceptAllFilterChain(), true)

	tests := []struct {
		name          string
		routesAdd     []*route.Route
		routesRemove  []*route.Route
		expected      []*route.Route
		expectedCount int64
	}{
		{
			name: "Add an iBGP route (without success)",
			routesAdd: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							Source:  net.IPv4(0),
							NextHop: net.IPv4(0),
						},
					},
				}),
			},
			expected:      []*route.Route{},
			expectedCount: 0,
		},
		{
			name: "Add an eBGP route (with success)",
			routesAdd: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							EBGP:    true,
							Source:  net.IPv4(0),
							NextHop: net.IPv4FromOctets(1, 2, 3, 4),
						},
						ASPath: &types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen: 1,
					},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop:   net.IPv4FromOctets(1, 2, 3, 4),
							Origin:    0,
							MED:       0,
							EBGP:      true,
							LocalPref: 0,
							Source:    net.IPv4(0),
						},
						ASPath: &types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen:         1,
						UnknownAttributes: nil,
						PathIdentifier:    1,
					},
				}),
			},
			expectedCount: 1,
		},
		{
			name: "Try to remove slightly different route",
			routesRemove: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop:   net.IPv4FromOctets(1, 2, 3, 4),
							Origin:    0,
							MED:       1, // MED of route present in table is 0
							EBGP:      true,
							LocalPref: 0,
							Source:    net.IPv4(0),
						},
						ASPath: &types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen:         1,
						Communities:       &types.Communities{},
						LargeCommunities:  &types.LargeCommunities{},
						UnknownAttributes: nil,
						PathIdentifier:    0,
					},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop:   net.IPv4FromOctets(1, 2, 3, 4),
							Origin:    0,
							MED:       0,
							EBGP:      true,
							LocalPref: 0,
							Source:    net.IPv4(0),
						},
						ASPath: &types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen:         1,
						UnknownAttributes: nil,
						PathIdentifier:    1,
					},
				}),
			},
			expectedCount: 1,
		},
		{
			name: "Remove route added in 2nd step",
			routesRemove: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							EBGP:    true,
							Source:  net.IPv4(0),
							NextHop: net.IPv4FromOctets(1, 2, 3, 4),
						},
						ASPath: &types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen: 1,
					},
				}),
			},
			expected:      []*route.Route{},
			expectedCount: 0,
		},
		{
			name: "Try to add route with NO_EXPORT community set (without success)",
			routesAdd: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							Source:  net.IPv4(0),
							NextHop: net.IPv4FromOctets(1, 2, 3, 4),
						},
						Communities: &types.Communities{
							types.WellKnownCommunityNoExport,
						},
					},
				}),
			},
			expected:      []*route.Route{},
			expectedCount: 0,
		},
		{
			name: "Try to add route with NO_EXPORT community set (without success)",
			routesAdd: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							Source:  net.IPv4(0),
							NextHop: net.IPv4FromOctets(1, 2, 3, 4),
						},
						Communities: &types.Communities{
							types.WellKnownCommunityNoAdvertise,
						},
					},
				}),
			},
			expected:      []*route.Route{},
			expectedCount: 0,
		},

		// Ok table is empty, re add previous route
		{
			name: "Readd an eBGP route (with success)",
			routesAdd: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							EBGP:    true,
							NextHop: net.IPv4FromOctets(1, 2, 3, 4),
							Source:  net.IPv4(0),
						},
						ASPath: &types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen: 1,
					},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop:   net.IPv4FromOctets(1, 2, 3, 4),
							Origin:    0,
							MED:       0,
							EBGP:      true,
							LocalPref: 0,
							Source:    net.IPv4(0),
						},
						ASPath: &types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen:         1,
						UnknownAttributes: nil,
						PathIdentifier:    2,
					},
				}),
			},
			expectedCount: 1,
		},
		{
			name: "Add 2nd path to existing one with different NH (with success)",
			routesAdd: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							EBGP:    true,
							NextHop: net.IPv4FromOctets(2, 3, 4, 5),
							Source:  net.IPv4(0),
						},
						ASPath: &types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen: 1,
					},
				}),
			},
			expected: []*route.Route{
				route.NewRouteAddPath(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), []*route.Path{
					{
						Type: route.BGPPathType,
						BGPPath: &route.BGPPath{
							BGPPathA: &route.BGPPathA{
								NextHop:   net.IPv4FromOctets(1, 2, 3, 4),
								Origin:    0,
								MED:       0,
								EBGP:      true,
								LocalPref: 0,
								Source:    net.IPv4(0),
							},
							ASPath: &types.ASPath{
								types.ASPathSegment{
									Type: types.ASSequence,
									ASNs: []uint32{
										201701,
									},
								},
							},
							ASPathLen:         1,
							UnknownAttributes: nil,
							PathIdentifier:    2,
						},
					},
					{
						Type: route.BGPPathType,
						BGPPath: &route.BGPPath{
							BGPPathA: &route.BGPPathA{
								NextHop:   net.IPv4FromOctets(2, 3, 4, 5),
								Origin:    0,
								MED:       0,
								EBGP:      true,
								LocalPref: 0,
								Source:    net.IPv4(0),
							},
							ASPath: &types.ASPath{
								types.ASPathSegment{
									Type: types.ASSequence,
									ASNs: []uint32{
										201701,
									},
								},
							},
							ASPathLen:         1,
							UnknownAttributes: nil,
							PathIdentifier:    3,
						},
					}}),
			},
			expectedCount: 1,
		},
		{
			name: "Remove 2nd path added above",
			routesRemove: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							EBGP:    true,
							NextHop: net.IPv4FromOctets(2, 3, 4, 5),
							Source:  net.IPv4(0),
						},
						ASPath: &types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen: 1,
					},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							NextHop:   net.IPv4FromOctets(1, 2, 3, 4),
							Origin:    0,
							MED:       0,
							EBGP:      true,
							LocalPref: 0,
							Source:    net.IPv4(0),
						},
						ASPath: &types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen:         1,
						UnknownAttributes: nil,
						PathIdentifier:    2,
					},
				}),
			},
			expectedCount: 1,
		},
		{
			name: "Re-add 2nd path to existing one with different NH (with success)",
			routesAdd: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							EBGP:    true,
							NextHop: net.IPv4FromOctets(3, 4, 5, 6),
							Source:  net.IPv4(0),
						},
						ASPath: &types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen: 1,
					},
				}),
			},
			expected: []*route.Route{
				route.NewRouteAddPath(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), []*route.Path{
					{
						Type: route.BGPPathType,
						BGPPath: &route.BGPPath{
							BGPPathA: &route.BGPPathA{
								NextHop:   net.IPv4FromOctets(1, 2, 3, 4),
								Origin:    0,
								MED:       0,
								EBGP:      true,
								LocalPref: 0,
								Source:    net.IPv4(0),
							},
							ASPath: &types.ASPath{
								types.ASPathSegment{
									Type: types.ASSequence,
									ASNs: []uint32{
										201701,
									},
								},
							},
							ASPathLen:         1,
							UnknownAttributes: nil,
							PathIdentifier:    2,
						},
					},
					{
						Type: route.BGPPathType,
						BGPPath: &route.BGPPath{
							BGPPathA: &route.BGPPathA{
								NextHop:   net.IPv4FromOctets(3, 4, 5, 6),
								Origin:    0,
								MED:       0,
								EBGP:      true,
								LocalPref: 0,
								Source:    net.IPv4(0),
							},
							ASPath: &types.ASPath{
								types.ASPathSegment{
									Type: types.ASSequence,
									ASNs: []uint32{
										201701,
									},
								},
							},
							ASPathLen:         1,
							UnknownAttributes: nil,
							PathIdentifier:    4,
						},
					}}),
			},
			expectedCount: 1,
		},
		{
			name: "Add 3rd path to existing ones, containing NO_EXPORT community (successful)",
			routesAdd: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							EBGP:    true,
							NextHop: net.IPv4FromOctets(4, 5, 6, 7),
							Source:  net.IPv4(0),
						},
						ASPath: &types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen: 1,
						Communities: &types.Communities{
							types.WellKnownCommunityNoExport,
						},
					},
				}),
			},
			expected: []*route.Route{
				route.NewRouteAddPath(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), []*route.Path{
					{
						Type: route.BGPPathType,
						BGPPath: &route.BGPPath{
							BGPPathA: &route.BGPPathA{
								NextHop:   net.IPv4FromOctets(1, 2, 3, 4),
								Origin:    0,
								MED:       0,
								EBGP:      true,
								LocalPref: 0,
								Source:    net.IPv4(0),
							},
							ASPath: &types.ASPath{
								types.ASPathSegment{
									Type: types.ASSequence,
									ASNs: []uint32{
										201701,
									},
								},
							},
							ASPathLen:         1,
							UnknownAttributes: nil,
							PathIdentifier:    2,
						},
					},
					{
						Type: route.BGPPathType,
						BGPPath: &route.BGPPath{
							BGPPathA: &route.BGPPathA{
								NextHop:   net.IPv4FromOctets(3, 4, 5, 6),
								Origin:    0,
								MED:       0,
								EBGP:      true,
								LocalPref: 0,
								Source:    net.IPv4(0),
							},
							ASPath: &types.ASPath{
								types.ASPathSegment{
									Type: types.ASSequence,
									ASNs: []uint32{
										201701,
									},
								},
							},
							ASPathLen:         1,
							UnknownAttributes: nil,
							PathIdentifier:    4,
						},
					},
					{
						Type: route.BGPPathType,
						BGPPath: &route.BGPPath{
							BGPPathA: &route.BGPPathA{
								NextHop:   net.IPv4FromOctets(4, 5, 6, 7),
								Origin:    0,
								MED:       0,
								EBGP:      true,
								LocalPref: 0,
								Source:    net.IPv4(0),
							},
							ASPath: &types.ASPath{
								types.ASPathSegment{
									Type: types.ASSequence,
									ASNs: []uint32{
										201701,
									},
								},
							},
							ASPathLen: 1,
							Communities: &types.Communities{
								types.WellKnownCommunityNoExport,
							},
							UnknownAttributes: nil,
							PathIdentifier:    5,
						},
					},
				}),
			},
			expectedCount: 1,
		},
		{
			name: "Add 4th path to existing ones, containing NO_ADVERTISE community",
			routesAdd: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							EBGP:    true,
							NextHop: net.IPv4FromOctets(5, 6, 7, 8),
							Source:  net.IPv4(0),
						},
						ASPath: &types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen: 1,
						Communities: &types.Communities{
							types.WellKnownCommunityNoAdvertise,
						},
					},
				}),
			},
			expected:      []*route.Route{},
			expectedCount: 0,
		},
	}

	for i, test := range tests {
		fmt.Printf("Running iBGP AddPath test #%d: %s\n", i+1, test.name)
		for _, route := range test.routesAdd {
			adjRIBOut.AddPath(route.Prefix(), route.Paths()[0])
		}

		for _, route := range test.routesRemove {
			adjRIBOut.RemovePath(route.Prefix(), route.Paths()[0])
		}

		if !assert.Equalf(t, test.expected, adjRIBOut.rt.Dump(), "Test %q", test.name) {
			return
		}

		actualCount := adjRIBOut.RouteCount()
		if test.expectedCount != actualCount {
			t.Errorf("Expected route count %d differs from actual route count %d!\n", test.expectedCount, actualCount)
			return
		}
	}
}
