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
		CapAddPathRX:      false,
	}

	adjRIBOut := New(neighborBestOnlyEBGP, filter.NewAcceptAllFilter())

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
		fmt.Printf("Running test #%d: %s\n", i+1, test.name)
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
		CapAddPathRX:      false,
	}

	adjRIBOut := New(neighborBestOnlyEBGP, filter.NewAcceptAllFilter())

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
			expected: []*route.Route{},
		},
	}

	for _, test := range tests {
		fmt.Printf("Running test: %s\n", test.name)
		for _, route := range test.routesAdd {
			//fmt.Printf("Adding prefix %v\n", route.Prefix().String())
			adjRIBOut.AddPath(route.Prefix(), route.Paths()[0])
		}

		for _, route := range test.routesRemove {
			//fmt.Printf("Removing prefix %v\n", route.Prefix().String())
			adjRIBOut.RemovePath(route.Prefix(), route.Paths()[0])
		}

		assert.Equal(t, test.expected, adjRIBOut.rt.Dump())

		actualCount := adjRIBOut.RouteCount()
		if test.expectedCount != actualCount {
			t.Errorf("Expected route count %d differs from actual route count %d!\n", test.expectedCount, actualCount)
		}
		// [0].Paths()[0].BGPPath.ASPath
	}
}
