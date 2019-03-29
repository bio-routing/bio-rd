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

	adjRIBOut := New(neighborBestOnlyEBGP, filter.NewAcceptAllFilter(), false)

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
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{NextHop: neighborBestOnlyEBGP.LocalAddress,
						ASPath: types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									neighborBestOnlyEBGP.LocalASN,
								},
							},
						},
						ASPathLen:         1,
						Origin:            0,
						MED:               0,
						EBGP:              false,
						Communities:       []uint32{},
						LargeCommunities:  []types.LargeCommunity{},
						UnknownAttributes: nil,
						PathIdentifier:    0,
						LocalPref:         0,
						Source:            net.IP{}},
				}),
			},
			expectedCount: 1,
		},
		{
			name: "Try to remove unpresent route",
			routesRemove: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{NextHop: neighborBestOnlyEBGP.LocalAddress,
						ASPath: types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									neighborBestOnlyEBGP.LocalASN,
								},
							},
						},
						ASPathLen:         1,
						Origin:            0,
						MED:               1,
						EBGP:              false,
						Communities:       []uint32{},
						LargeCommunities:  []types.LargeCommunity{},
						UnknownAttributes: nil,
						PathIdentifier:    0,
						LocalPref:         0,
						Source:            net.IP{}},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{NextHop: neighborBestOnlyEBGP.LocalAddress,
						ASPath: types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									neighborBestOnlyEBGP.LocalASN,
								},
							},
						},
						ASPathLen:         1,
						Origin:            0,
						MED:               0,
						EBGP:              false,
						Communities:       []uint32{},
						LargeCommunities:  []types.LargeCommunity{},
						UnknownAttributes: nil,
						PathIdentifier:    0,
						LocalPref:         0,
						Source:            net.IP{}},
				}),
			},
			expectedCount: 1,
		},
		{
			name: "Remove route added in first step",
			routesRemove: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{NextHop: neighborBestOnlyEBGP.LocalAddress,
						ASPath: types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									neighborBestOnlyEBGP.LocalASN,
								},
							},
						},
						ASPathLen:         1,
						Origin:            0,
						MED:               0,
						EBGP:              false,
						Communities:       []uint32{},
						LargeCommunities:  []types.LargeCommunity{},
						UnknownAttributes: nil,
						PathIdentifier:    0,
						LocalPref:         0,
						Source:            net.IP{}},
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
						Communities: []uint32{
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
						Communities: []uint32{
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
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{NextHop: neighborBestOnlyEBGP.LocalAddress,
						ASPath: types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									neighborBestOnlyEBGP.LocalASN,
								},
							},
						},
						ASPathLen:         1,
						Origin:            0,
						MED:               0,
						EBGP:              false,
						Communities:       []uint32{},
						LargeCommunities:  []types.LargeCommunity{},
						UnknownAttributes: nil,
						PathIdentifier:    0,
						LocalPref:         0,
						Source:            net.IP{}},
				}),
			},
			expectedCount: 1,
		},
		{
			name: "Try to remove route with NO_EXPORT community set",
			routesRemove: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{NextHop: neighborBestOnlyEBGP.LocalAddress,
						ASPath: types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									neighborBestOnlyEBGP.LocalASN,
								},
							},
						},
						ASPathLen: 1,
						Origin:    0,
						MED:       0,
						EBGP:      false,
						Communities: []uint32{
							types.WellKnownCommunityNoExport,
						},
						LargeCommunities:  []types.LargeCommunity{},
						UnknownAttributes: nil,
						PathIdentifier:    0,
						LocalPref:         0,
						Source:            net.IP{}},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{NextHop: neighborBestOnlyEBGP.LocalAddress,
						ASPath: types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									neighborBestOnlyEBGP.LocalASN,
								},
							},
						},
						ASPathLen:         1,
						Origin:            0,
						MED:               0,
						EBGP:              false,
						Communities:       []uint32{},
						LargeCommunities:  []types.LargeCommunity{},
						UnknownAttributes: nil,
						PathIdentifier:    0,
						LocalPref:         0,
						Source:            net.IP{}},
				}),
			},
			expectedCount: 1,
		},
		{
			name: "Try to remove non-existent prefix",
			routesRemove: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 23, 42, 0), 24), &route.Path{
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{NextHop: neighborBestOnlyEBGP.LocalAddress,
						ASPath: types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									neighborBestOnlyEBGP.LocalASN,
								},
							},
						},
						ASPathLen:         1,
						Origin:            0,
						MED:               0,
						EBGP:              false,
						Communities:       []uint32{},
						LargeCommunities:  []types.LargeCommunity{},
						UnknownAttributes: nil,
						PathIdentifier:    0,
						LocalPref:         0,
						Source:            net.IP{}},
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

	adjRIBOut := New(neighborBestOnlyEBGP, filter.NewAcceptAllFilter(), false)

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
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
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
						EBGP: true,
						ASPath: types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen: 1,
						NextHop:   net.IPv4FromOctets(1, 2, 3, 4),
					},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						NextHop: net.IPv4FromOctets(1, 2, 3, 4),
						ASPath: types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen:         1,
						Origin:            0,
						MED:               0,
						EBGP:              true,
						Communities:       []uint32{},
						LargeCommunities:  []types.LargeCommunity{},
						UnknownAttributes: nil,
						PathIdentifier:    0,
						LocalPref:         0,
						Source:            net.IP{}},
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
						NextHop: net.IPv4FromOctets(1, 2, 3, 4),
						ASPath: types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen:         1,
						Origin:            0,
						MED:               1,
						EBGP:              true,
						Communities:       []uint32{},
						LargeCommunities:  []types.LargeCommunity{},
						UnknownAttributes: nil,
						PathIdentifier:    0,
						LocalPref:         0,
						Source:            net.IP{}},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						NextHop: net.IPv4FromOctets(1, 2, 3, 4),
						ASPath: types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen:         1,
						Origin:            0,
						MED:               0,
						EBGP:              true,
						Communities:       []uint32{},
						LargeCommunities:  []types.LargeCommunity{},
						UnknownAttributes: nil,
						PathIdentifier:    0,
						LocalPref:         0,
						Source:            net.IP{}},
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
						NextHop: net.IPv4FromOctets(1, 2, 3, 4),
						ASPath: types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen:         1,
						Origin:            0,
						MED:               0,
						EBGP:              true,
						Communities:       []uint32{},
						LargeCommunities:  []types.LargeCommunity{},
						UnknownAttributes: nil,
						PathIdentifier:    0,
						LocalPref:         0,
						Source:            net.IP{}},
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
						Communities: []uint32{
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
						Communities: []uint32{
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

	adjRIBOut := New(neighborBestOnlyRR, filter.NewAcceptAllFilter(), false)

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
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						Communities:      []uint32{},
						LargeCommunities: []types.LargeCommunity{},
						ASPath:           types.ASPath{},
						ClusterList: []uint32{
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
						EBGP: true,
						ASPath: types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen: 1,
						NextHop:   net.IPv4FromOctets(1, 2, 3, 4),
					},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						NextHop: net.IPv4FromOctets(1, 2, 3, 4),
						ASPath: types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen:         1,
						Origin:            0,
						MED:               0,
						EBGP:              true,
						Communities:       []uint32{},
						LargeCommunities:  []types.LargeCommunity{},
						UnknownAttributes: nil,
						PathIdentifier:    0,
						LocalPref:         0,
						Source:            net.IP{},
						ClusterList: []uint32{
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
						NextHop: net.IPv4FromOctets(1, 2, 3, 4),
						ASPath: types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen:         1,
						Origin:            0,
						MED:               1, // Existing route has MED 0
						EBGP:              true,
						Communities:       []uint32{},
						LargeCommunities:  []types.LargeCommunity{},
						UnknownAttributes: nil,
						PathIdentifier:    0,
						LocalPref:         0,
						Source:            net.IP{}},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						NextHop: net.IPv4FromOctets(1, 2, 3, 4),
						ASPath: types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen:         1,
						Origin:            0,
						MED:               0,
						EBGP:              true,
						Communities:       []uint32{},
						LargeCommunities:  []types.LargeCommunity{},
						UnknownAttributes: nil,
						PathIdentifier:    0,
						LocalPref:         0,
						Source:            net.IP{},
						ClusterList: []uint32{
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
						NextHop: net.IPv4FromOctets(1, 2, 3, 4),
						ASPath: types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen:         1,
						Origin:            0,
						MED:               0,
						EBGP:              true,
						Communities:       []uint32{},
						LargeCommunities:  []types.LargeCommunity{},
						UnknownAttributes: nil,
						PathIdentifier:    0,
						LocalPref:         0,
						Source:            net.IP{},
						ClusterList: []uint32{
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
						Communities: []uint32{
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
						Communities: []uint32{
							types.WellKnownCommunityNoExport,
						},
					},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						ASPath:    types.ASPath{},
						ASPathLen: 0,
						Origin:    0,
						MED:       0,
						EBGP:      false,
						Communities: []uint32{
							types.WellKnownCommunityNoExport,
						},
						LargeCommunities:  []types.LargeCommunity{},
						UnknownAttributes: nil,
						PathIdentifier:    0,
						LocalPref:         0,
						Source:            net.IP{},
						ClusterList: []uint32{
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
						ASPath:    types.ASPath{},
						ASPathLen: 0,
						Origin:    0,
						MED:       0,
						EBGP:      false,
						Communities: []uint32{
							types.WellKnownCommunityNoExport,
						},
						LargeCommunities:  []types.LargeCommunity{},
						UnknownAttributes: nil,
						PathIdentifier:    0,
						LocalPref:         0,
						Source:            net.IP{},
						ClusterList: []uint32{
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
						OriginatorID: 42,
						ClusterList: []uint32{
							23,
						},
					},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						ASPath:            types.ASPath{},
						ASPathLen:         0,
						Origin:            0,
						MED:               0,
						EBGP:              false,
						Communities:       []uint32{},
						LargeCommunities:  []types.LargeCommunity{},
						UnknownAttributes: nil,
						PathIdentifier:    0,
						LocalPref:         0,
						Source:            net.IP{},
						ClusterList: []uint32{
							neighborBestOnlyRR.ClusterID,
							23,
						},
						OriginatorID: 42,
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

	adjRIBOut := New(neighborBestOnlyEBGP, filter.NewAcceptAllFilter(), true)

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
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
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
						EBGP: true,
						ASPath: types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen: 1,
						NextHop:   net.IPv4FromOctets(1, 2, 3, 4),
					},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						NextHop: net.IPv4FromOctets(1, 2, 3, 4),
						ASPath: types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen:         1,
						Origin:            0,
						MED:               0,
						EBGP:              true,
						Communities:       []uint32{},
						LargeCommunities:  []types.LargeCommunity{},
						UnknownAttributes: nil,
						PathIdentifier:    1,
						LocalPref:         0,
						Source:            net.IP{}},
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
						NextHop: net.IPv4FromOctets(1, 2, 3, 4),
						ASPath: types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen:         1,
						Origin:            0,
						MED:               1, // MED of route present in table is 0
						EBGP:              true,
						Communities:       []uint32{},
						LargeCommunities:  []types.LargeCommunity{},
						UnknownAttributes: nil,
						PathIdentifier:    0,
						LocalPref:         0,
						Source:            net.IP{}},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						NextHop: net.IPv4FromOctets(1, 2, 3, 4),
						ASPath: types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen:         1,
						Origin:            0,
						MED:               0,
						EBGP:              true,
						Communities:       []uint32{},
						LargeCommunities:  []types.LargeCommunity{},
						UnknownAttributes: nil,
						PathIdentifier:    1,
						LocalPref:         0,
						Source:            net.IP{}},
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
						NextHop: net.IPv4FromOctets(1, 2, 3, 4),
						ASPath: types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen:         1,
						Origin:            0,
						MED:               0,
						EBGP:              true,
						Communities:       []uint32{},
						LargeCommunities:  []types.LargeCommunity{},
						UnknownAttributes: nil,
						PathIdentifier:    0, // We calculate PathID in RIBOut so none is present when removing from RIBOut
						LocalPref:         0,
						Source:            net.IP{}},
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
						Communities: []uint32{
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
						Communities: []uint32{
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
						EBGP: true,
						ASPath: types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen: 1,
						NextHop:   net.IPv4FromOctets(1, 2, 3, 4),
					},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						NextHop: net.IPv4FromOctets(1, 2, 3, 4),
						ASPath: types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen:         1,
						Origin:            0,
						MED:               0,
						EBGP:              true,
						Communities:       []uint32{},
						LargeCommunities:  []types.LargeCommunity{},
						UnknownAttributes: nil,
						PathIdentifier:    2,
						LocalPref:         0,
						Source:            net.IP{}},
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
						EBGP: true,
						ASPath: types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen: 1,
						NextHop:   net.IPv4FromOctets(2, 3, 4, 5),
					},
				}),
			},
			expected: []*route.Route{
				route.NewRouteAddPath(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), []*route.Path{
					{
						Type: route.BGPPathType,
						BGPPath: &route.BGPPath{
							NextHop: net.IPv4FromOctets(1, 2, 3, 4),
							ASPath: types.ASPath{
								types.ASPathSegment{
									Type: types.ASSequence,
									ASNs: []uint32{
										201701,
									},
								},
							},
							ASPathLen:         1,
							Origin:            0,
							MED:               0,
							EBGP:              true,
							Communities:       []uint32{},
							LargeCommunities:  []types.LargeCommunity{},
							UnknownAttributes: nil,
							PathIdentifier:    2,
							LocalPref:         0,
							Source:            net.IP{}},
					},
					{
						Type: route.BGPPathType,
						BGPPath: &route.BGPPath{
							NextHop: net.IPv4FromOctets(2, 3, 4, 5),
							ASPath: types.ASPath{
								types.ASPathSegment{
									Type: types.ASSequence,
									ASNs: []uint32{
										201701,
									},
								},
							},
							ASPathLen:         1,
							Origin:            0,
							MED:               0,
							EBGP:              true,
							Communities:       []uint32{},
							LargeCommunities:  []types.LargeCommunity{},
							UnknownAttributes: nil,
							PathIdentifier:    3,
							LocalPref:         0,
							Source:            net.IP{}},
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
						NextHop: net.IPv4FromOctets(2, 3, 4, 5),
						ASPath: types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen:         1,
						Origin:            0,
						MED:               0,
						EBGP:              true,
						Communities:       []uint32{},
						LargeCommunities:  []types.LargeCommunity{},
						UnknownAttributes: nil,
						PathIdentifier:    0,
						LocalPref:         0,
						Source:            net.IP{}},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						NextHop: net.IPv4FromOctets(1, 2, 3, 4),
						ASPath: types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen:         1,
						Origin:            0,
						MED:               0,
						EBGP:              true,
						Communities:       []uint32{},
						LargeCommunities:  []types.LargeCommunity{},
						UnknownAttributes: nil,
						PathIdentifier:    2,
						LocalPref:         0,
						Source:            net.IP{}},
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
						EBGP: true,
						ASPath: types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen: 1,
						NextHop:   net.IPv4FromOctets(3, 4, 5, 6),
					},
				}),
			},
			expected: []*route.Route{
				route.NewRouteAddPath(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), []*route.Path{
					{
						Type: route.BGPPathType,
						BGPPath: &route.BGPPath{
							NextHop: net.IPv4FromOctets(1, 2, 3, 4),
							ASPath: types.ASPath{
								types.ASPathSegment{
									Type: types.ASSequence,
									ASNs: []uint32{
										201701,
									},
								},
							},
							ASPathLen:         1,
							Origin:            0,
							MED:               0,
							EBGP:              true,
							Communities:       []uint32{},
							LargeCommunities:  []types.LargeCommunity{},
							UnknownAttributes: nil,
							PathIdentifier:    2,
							LocalPref:         0,
							Source:            net.IP{}},
					},
					{
						Type: route.BGPPathType,
						BGPPath: &route.BGPPath{
							NextHop: net.IPv4FromOctets(3, 4, 5, 6),
							ASPath: types.ASPath{
								types.ASPathSegment{
									Type: types.ASSequence,
									ASNs: []uint32{
										201701,
									},
								},
							},
							ASPathLen:         1,
							Origin:            0,
							MED:               0,
							EBGP:              true,
							Communities:       []uint32{},
							LargeCommunities:  []types.LargeCommunity{},
							UnknownAttributes: nil,
							PathIdentifier:    4,
							LocalPref:         0,
							Source:            net.IP{}},
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
						EBGP: true,
						ASPath: types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen: 1,
						NextHop:   net.IPv4FromOctets(4, 5, 6, 7),
						Communities: []uint32{
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
							NextHop: net.IPv4FromOctets(1, 2, 3, 4),
							ASPath: types.ASPath{
								types.ASPathSegment{
									Type: types.ASSequence,
									ASNs: []uint32{
										201701,
									},
								},
							},
							ASPathLen:         1,
							Origin:            0,
							MED:               0,
							EBGP:              true,
							Communities:       []uint32{},
							LargeCommunities:  []types.LargeCommunity{},
							UnknownAttributes: nil,
							PathIdentifier:    2,
							LocalPref:         0,
							Source:            net.IP{}},
					},
					{
						Type: route.BGPPathType,
						BGPPath: &route.BGPPath{
							NextHop: net.IPv4FromOctets(3, 4, 5, 6),
							ASPath: types.ASPath{
								types.ASPathSegment{
									Type: types.ASSequence,
									ASNs: []uint32{
										201701,
									},
								},
							},
							ASPathLen:         1,
							Origin:            0,
							MED:               0,
							EBGP:              true,
							Communities:       []uint32{},
							LargeCommunities:  []types.LargeCommunity{},
							UnknownAttributes: nil,
							PathIdentifier:    4,
							LocalPref:         0,
							Source:            net.IP{}},
					},
					{
						Type: route.BGPPathType,
						BGPPath: &route.BGPPath{
							NextHop: net.IPv4FromOctets(4, 5, 6, 7),
							ASPath: types.ASPath{
								types.ASPathSegment{
									Type: types.ASSequence,
									ASNs: []uint32{
										201701,
									},
								},
							},
							ASPathLen: 1,
							Origin:    0,
							MED:       0,
							EBGP:      true,
							Communities: []uint32{
								types.WellKnownCommunityNoExport,
							},
							LargeCommunities:  []types.LargeCommunity{},
							UnknownAttributes: nil,
							PathIdentifier:    5,
							LocalPref:         0,
							Source:            net.IP{}},
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
						EBGP: true,
						ASPath: types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{
									201701,
								},
							},
						},
						ASPathLen: 1,
						NextHop:   net.IPv4FromOctets(5, 6, 7, 8),
						Communities: []uint32{
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

func TestPrefixLimit(t *testing.T) {
	tests := []struct {
		name                string
		prefixesToAdd       uint
		announcePrefixLimit uint
		wantsHitError       bool
		expectedLimitHit    uint
	}{
		{
			name:          "no limit configured",
			prefixesToAdd: 3,
		},
		{
			name:                "prefix limit configured but not hit",
			prefixesToAdd:       2,
			announcePrefixLimit: 2,
		},
		{
			name:                "prefix limit hit",
			prefixesToAdd:       3,
			announcePrefixLimit: 2,
			wantsHitError:       true,
			expectedLimitHit:    2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rib := New(&routingtable.Neighbor{}, filter.NewAcceptAllFilter(), false)
			rib.announcePrefixLimit = test.announcePrefixLimit

			p := &route.Path{
				Type: route.BGPPathType,
				BGPPath: &route.BGPPath{
					NextHop: net.IPv4FromOctets(192, 168, 0, 0),
				},
			}

			var err error
			for i := uint(0); i < test.prefixesToAdd; i++ {
				pfx := net.NewPfx(net.IPv4(uint32(i)), 32)
				err = rib.AddPath(pfx, p)
				if err != nil {
					break
				}
			}

			if !test.wantsHitError && err == nil {
				return
			}

			if err == nil {
				t.Fatal("expected error, got none")
			}

			switch err.(type) {
			case *routingtable.PrefixLimitError:
				assert.Equal(t, test.expectedLimitHit, err.(*routingtable.PrefixLimitError).Limit())
			default:
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
