package routingtable

import (
	"fmt"
	"testing"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/adjRIBIn"
	"github.com/bio-routing/bio-rd/routingtable/adjRIBOut"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
	"github.com/stretchr/testify/assert"
)

// This test sets up
func TestRedistribute(t *testing.T) {
	nullIP := bnet.IPv4FromOctets(0, 0, 0, 0)

	// Set up session attrs for a BGP peer A where we learn prefixes from (-> AdjRIBIn)
	peerAIP := bnet.IPv4FromOctets(10, 0, 0, 0)
	localAIP := bnet.IPv4FromOctets(10, 0, 0, 1)
	sessionAttrsA := routingtable.SessionAttrs{
		PeerIP:   &peerAIP,
		LocalIP:  &localAIP,
		LocalASN: 57165,
		PeerASN:  201701,
		RouterID: 42,
	}

	// Set up session attrs for a BGP peer B where send prefixes to,
	peerBIP := bnet.IPv4FromOctets(20, 0, 0, 0)
	localBIP := bnet.IPv4FromOctets(20, 0, 0, 1)
	sessionAttrsBeBGP := routingtable.SessionAttrs{
		PeerIP:   &peerBIP,
		LocalIP:  &localBIP,
		LocalASN: 57165,
		PeerASN:  3320,
		RouterID: 42,
		IBGP:     false,
	}
	sessionAttrsBiBGP := routingtable.SessionAttrs{
		PeerIP:   &peerBIP,
		LocalIP:  &localBIP,
		LocalASN: 57165,
		PeerASN:  57165,
		RouterID: 42,
		IBGP:     true,
	}

	tests := []struct {
		name                     string
		adjRIBInImportFilter     filter.Chain
		adjRIBOutExportFilter    filter.Chain
		adjRibOutSessionAttrs    routingtable.SessionAttrs
		addRoutesToAdjRIBIN      []*route.Route
		addRouteToLocRIB         []*route.Route
		expecxtedRoutesLocRIB    []*route.Route
		expecxtedRoutesAdjRIBOut []*route.Route
	}{

		{
			name:                  "eBGP to eBGP",
			adjRIBInImportFilter:  filter.NewAcceptAllFilterChain(),
			adjRIBOutExportFilter: filter.NewAcceptAllFilterChain(),
			adjRibOutSessionAttrs: sessionAttrsBeBGP,
			addRoutesToAdjRIBIN: []*route.Route{
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						ASPath:    types.NewASPath([]uint32{201701}),
						ASPathLen: 1,
						BGPPathA: &route.BGPPathA{
							Source:  &peerAIP,
							NextHop: &peerAIP,
							EBGP:    true,
						},
					},
				}),
			},
			addRouteToLocRIB: nil,
			expecxtedRoutesLocRIB: []*route.Route{
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						ASPath:    types.NewASPath([]uint32{201701}),
						ASPathLen: 1,
						BGPPathA: &route.BGPPathA{
							Source:  &peerAIP,
							NextHop: &peerAIP,
							EBGP:    true,
							// FIXME? LocalPref: 100,
						},
					},
				}),
			},
			expecxtedRoutesAdjRIBOut: []*route.Route{
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						ASPath:    types.NewASPath([]uint32{57165, 201701}),
						ASPathLen: 2,
						BGPPathA: &route.BGPPathA{
							Source:  &peerAIP,
							NextHop: &localBIP,
							EBGP:    true,
							// FIXME? LocalPref: 100,
						},
					},
				}),
			},
		},
		{
			name:                  "Static to eBGP",
			adjRIBInImportFilter:  filter.NewAcceptAllFilterChain(),
			adjRIBOutExportFilter: filter.NewAcceptAllFilterChain(),
			adjRibOutSessionAttrs: sessionAttrsBeBGP,
			addRoutesToAdjRIBIN:   []*route.Route{},
			addRouteToLocRIB: []*route.Route{
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(11, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.StaticPathType,
					StaticPath: &route.StaticPath{
						NextHop: &peerAIP,
					},
				}),
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(11, 0, 0, 0), 9).Ptr(), &route.Path{
					Type:       route.StaticPathType,
					StaticPath: nil,
				}),
			},
			expecxtedRoutesLocRIB: []*route.Route{
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(11, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.StaticPathType,
					StaticPath: &route.StaticPath{
						NextHop: &peerAIP,
					},
				}),
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(11, 0, 0, 0), 9).Ptr(), &route.Path{
					Type:       route.StaticPathType,
					StaticPath: nil,
				}),
			},
			expecxtedRoutesAdjRIBOut: []*route.Route{
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(11, 0, 0, 0), 8).Ptr(), &route.Path{
					Type:              route.BGPPathType,
					RedistributedFrom: route.StaticPathType,
					BGPPath: &route.BGPPath{
						ASPath:    types.NewASPath([]uint32{57165}),
						ASPathLen: 1,
						BGPPathA: &route.BGPPathA{
							Source:  &nullIP,
							NextHop: &localBIP,
						},
					},
					StaticPath: &route.StaticPath{
						NextHop: &peerAIP,
					},
				}),
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(11, 0, 0, 0), 9).Ptr(), &route.Path{
					Type:              route.BGPPathType,
					RedistributedFrom: route.StaticPathType,
					BGPPath: &route.BGPPath{
						ASPath:    types.NewASPath([]uint32{57165}),
						ASPathLen: 1,
						BGPPathA: &route.BGPPathA{
							Source:  &nullIP,
							NextHop: &localBIP,
						},
					},
					StaticPath: nil,
				}),
			},
		},
		{
			name:                  "Static to iBGP",
			adjRIBInImportFilter:  filter.NewAcceptAllFilterChain(),
			adjRIBOutExportFilter: filter.NewAcceptAllFilterChain(),
			adjRibOutSessionAttrs: sessionAttrsBiBGP,
			addRoutesToAdjRIBIN:   []*route.Route{},
			addRouteToLocRIB: []*route.Route{
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(11, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.StaticPathType,
					StaticPath: &route.StaticPath{
						NextHop: &peerAIP,
					},
				}),
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(11, 0, 0, 0), 9).Ptr(), &route.Path{
					Type:       route.StaticPathType,
					StaticPath: nil,
				}),
			},
			expecxtedRoutesLocRIB: []*route.Route{
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(11, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.StaticPathType,
					StaticPath: &route.StaticPath{
						NextHop: &peerAIP,
					},
				}),
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(11, 0, 0, 0), 9).Ptr(), &route.Path{
					Type:       route.StaticPathType,
					StaticPath: nil,
				}),
			},
			expecxtedRoutesAdjRIBOut: []*route.Route{
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(11, 0, 0, 0), 8).Ptr(), &route.Path{
					Type:              route.BGPPathType,
					RedistributedFrom: route.StaticPathType,
					BGPPath: &route.BGPPath{
						ASPath:    types.NewASPath([]uint32{}),
						ASPathLen: 0,
						BGPPathA: &route.BGPPathA{
							Source:  &nullIP,
							NextHop: &peerAIP,
						},
					},
					StaticPath: &route.StaticPath{
						NextHop: &peerAIP,
					},
				}),
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(11, 0, 0, 0), 9).Ptr(), &route.Path{
					Type:              route.BGPPathType,
					RedistributedFrom: route.StaticPathType,
					BGPPath: &route.BGPPath{
						ASPath:    types.NewASPath([]uint32{}),
						ASPathLen: 0,
						BGPPathA: &route.BGPPathA{
							Source:  &nullIP,
							NextHop: &localBIP,
						},
					},
					StaticPath: nil,
				}),
			},
		},

		{
			name:                  "Static + eBGP to eBGP",
			adjRIBInImportFilter:  filter.NewAcceptAllFilterChain(),
			adjRIBOutExportFilter: filter.NewAcceptAllFilterChain(),
			adjRibOutSessionAttrs: sessionAttrsBeBGP,
			addRoutesToAdjRIBIN: []*route.Route{
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						ASPath:    types.NewASPath([]uint32{201701}),
						ASPathLen: 1,
						BGPPathA: &route.BGPPathA{
							Source:  &peerAIP,
							NextHop: &peerAIP,
							EBGP:    true,
						},
					},
				}),
			},
			addRouteToLocRIB: []*route.Route{
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(11, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.StaticPathType,
					StaticPath: &route.StaticPath{
						NextHop: &peerAIP,
					},
				}),
			},
			expecxtedRoutesLocRIB: []*route.Route{
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						ASPath:    types.NewASPath([]uint32{201701}),
						ASPathLen: 1,
						BGPPathA: &route.BGPPathA{
							Source:  &peerAIP,
							NextHop: &peerAIP,
							EBGP:    true,
							// FIXME? LocalPref: 100,
						},
					},
				}),
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(11, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.StaticPathType,
					StaticPath: &route.StaticPath{
						NextHop: &peerAIP,
					},
				}),
			},
			expecxtedRoutesAdjRIBOut: []*route.Route{
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						ASPath:    types.NewASPath([]uint32{57165, 201701}),
						ASPathLen: 2,
						BGPPathA: &route.BGPPathA{
							Source:  &peerAIP,
							NextHop: &localBIP,
							EBGP:    true,
							// FIXME? LocalPref: 100,
						},
					},
				}),
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(11, 0, 0, 0), 8).Ptr(), &route.Path{
					Type:              route.BGPPathType,
					RedistributedFrom: route.StaticPathType,
					BGPPath: &route.BGPPath{
						ASPath:    types.NewASPath([]uint32{57165}),
						ASPathLen: 1,
						BGPPathA: &route.BGPPathA{
							Source:  &nullIP,
							NextHop: &localBIP,
						},
					},
					StaticPath: &route.StaticPath{
						NextHop: &peerAIP,
					},
				}),
			},
		},
	}

	for _, test := range tests {
		// Set up VRF and create an IPv4 Unicast LocRIB within
		vrf := vrf.NewUntrackedVRF("my shiny VRF", 0)

		locRIB, err := vrf.CreateIPv4UnicastLocRIB("inet.0")
		if err != nil {
			t.Fatalf("Failed to create IPv4 Unicast LocRIB in VRF")
		}

		// Set up an AdjRIBIn associated with peer A and register the LocRIB as a client
		adjRIBIn := adjRIBIn.New(test.adjRIBInImportFilter, vrf, sessionAttrsA)
		adjRIBIn.Register(locRIB)

		// Set up AdjRIBOut and register it as a client to LocRIB
		adjRIBOut := adjRIBOut.New(locRIB, test.adjRibOutSessionAttrs, test.adjRIBOutExportFilter)
		locRIB.Register(adjRIBOut)

		for _, r := range test.addRoutesToAdjRIBIN {
			for _, p := range r.Paths() {
				adjRIBIn.AddPath(r.Prefix(), p)
			}
		}

		for _, r := range test.addRouteToLocRIB {
			for _, p := range r.Paths() {
				locRIB.AddPath(r.Prefix(), p)
			}
		}

		// Make sure ecmpPaths is calcualated on expected routes in LocRIB
		for _, r := range test.expecxtedRoutesLocRIB {
			r.PathSelection()
		}

		assert.Equal(t, test.expecxtedRoutesLocRIB, locRIB.Dump(), fmt.Sprintf("LocRIB does not contain expected routes for test %q", test.name))
		assert.Equal(t, test.expecxtedRoutesAdjRIBOut, adjRIBOut.Dump(), fmt.Sprintf("AdjRIBOut does not contain expected routes for test %q", test.name))
	}
}
