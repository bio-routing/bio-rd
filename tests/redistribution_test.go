package tests

import (
	"fmt"
	"testing"

	bnet "github.com/bio-routing/bio-rd/net"

	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	GRPAdjRibIn "github.com/bio-routing/bio-rd/protocols/grp/adjRIBIn"
	GRPAdjRibOut "github.com/bio-routing/bio-rd/protocols/grp/adjRIBOut"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	BGPAdjRIBIn "github.com/bio-routing/bio-rd/routingtable/adjRIBIn"
	BGPAdjRIBOut "github.com/bio-routing/bio-rd/routingtable/adjRIBOut"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/routingtable/filter/actions"
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

	// GRP Meta data
	GRPMetaData := map[string]string{
		"foo": "bar",
		"key": "value",
	}
	GRPNhIP := bnet.IPv4FromOctets(1, 1, 1, 1).Ptr()

	tests := []struct {
		name                        string
		BGPadjRIBInImportFilter     filter.Chain
		BGPadjRIBOutExportFilter    filter.Chain
		BGPadjRibOutSessionAttrs    routingtable.SessionAttrs
		GRPadjRibInImportFilter     filter.Chain
		addRoutesToBGPAdjRIBIn      []*route.Route
		addRoutesToGRPAdjRIBIn      []*route.Route
		addRouteToLocRIB            []*route.Route
		expecxtedRoutesLocRIB       []*route.Route
		expecxtedRoutesBGPAdjRIBOut []*route.Route
		expecxtedRoutesGRPAdjRIBOut []*route.Route
		toBGP                       bool
		toGRP                       bool
	}{

		{
			name:                     "eBGP to eBGP",
			BGPadjRIBInImportFilter:  filter.NewAcceptAllFilterChain(),
			BGPadjRIBOutExportFilter: filter.NewAcceptAllFilterChain(),
			BGPadjRibOutSessionAttrs: sessionAttrsBeBGP,
			toBGP:                    true,
			addRoutesToBGPAdjRIBIn: []*route.Route{
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
			expecxtedRoutesBGPAdjRIBOut: []*route.Route{
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
			name:                     "Static to eBGP",
			BGPadjRIBInImportFilter:  filter.NewAcceptAllFilterChain(),
			BGPadjRIBOutExportFilter: filter.NewAcceptAllFilterChain(),
			BGPadjRibOutSessionAttrs: sessionAttrsBeBGP,
			toBGP:                    true,
			addRoutesToBGPAdjRIBIn:   []*route.Route{},
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
			expecxtedRoutesBGPAdjRIBOut: []*route.Route{
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
			name:                     "Static to iBGP",
			BGPadjRIBInImportFilter:  filter.NewAcceptAllFilterChain(),
			BGPadjRIBOutExportFilter: filter.NewAcceptAllFilterChain(),
			BGPadjRibOutSessionAttrs: sessionAttrsBiBGP,
			addRoutesToBGPAdjRIBIn:   []*route.Route{},
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
			expecxtedRoutesBGPAdjRIBOut: []*route.Route{
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
			name:                     "Static + eBGP to eBGP",
			BGPadjRIBInImportFilter:  filter.NewAcceptAllFilterChain(),
			BGPadjRIBOutExportFilter: filter.NewAcceptAllFilterChain(),
			BGPadjRibOutSessionAttrs: sessionAttrsBeBGP,
			toBGP:                    true,
			addRoutesToBGPAdjRIBIn: []*route.Route{
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
			expecxtedRoutesBGPAdjRIBOut: []*route.Route{
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
		{
			name:                    "Static + eBGP to GRP",
			BGPadjRIBInImportFilter: filter.NewAcceptAllFilterChain(),
			GRPadjRibInImportFilter: filter.NewAcceptAllFilterChain(),
			toBGP:                   false,
			toGRP:                   true,
			addRoutesToBGPAdjRIBIn: []*route.Route{
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
			expecxtedRoutesGRPAdjRIBOut: []*route.Route{
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type:              route.GRPPathType,
					RedistributedFrom: route.BGPPathType,
					GRPPath: &route.GRPPath{
						NextHop:  &peerAIP,
						MetaData: GRPMetaData,
					},
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
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(11, 0, 0, 0), 8).Ptr(), &route.Path{
					Type:              route.GRPPathType,
					RedistributedFrom: route.StaticPathType,
					GRPPath: &route.GRPPath{
						NextHop:  &peerAIP,
						MetaData: GRPMetaData,
					},
					StaticPath: &route.StaticPath{
						NextHop: &peerAIP,
					},
				}),
			},
		},
		{
			name:                    "eBGP to GRP with filter",
			BGPadjRIBInImportFilter: filter.NewAcceptAllFilterChain(),
			GRPadjRibInImportFilter: filter.Chain{
				filter.NewFilter(
					"accept & set NH",
					[]*filter.Term{
						filter.NewTerm(
							"accept & set NH",
							[]*filter.TermCondition{},
							[]actions.Action{
								actions.NewSetNextHopAction(GRPNhIP),
								actions.NewAcceptAction(),
							}),
					}),
			},
			toGRP: true,
			addRoutesToBGPAdjRIBIn: []*route.Route{
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
			expecxtedRoutesGRPAdjRIBOut: []*route.Route{
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type:              route.GRPPathType,
					RedistributedFrom: route.BGPPathType,
					GRPPath: &route.GRPPath{
						NextHop:  GRPNhIP,
						MetaData: GRPMetaData,
					},
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
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(11, 0, 0, 0), 8).Ptr(), &route.Path{
					Type:              route.GRPPathType,
					RedistributedFrom: route.StaticPathType,
					GRPPath: &route.GRPPath{
						NextHop:  GRPNhIP,
						MetaData: GRPMetaData,
					},
					StaticPath: &route.StaticPath{
						NextHop: &peerAIP,
					},
				}),
			},
		},
		{
			name:                     "GRP to eBGP + GRP with filter",
			BGPadjRIBInImportFilter:  filter.NewAcceptAllFilterChain(),
			BGPadjRibOutSessionAttrs: sessionAttrsBeBGP,
			GRPadjRibInImportFilter: filter.Chain{
				filter.NewFilter(
					"accept & set NH",
					[]*filter.Term{
						filter.NewTerm(
							"Ignore GRP routes",
							[]*filter.TermCondition{
								filter.NewTermConditionWithProtocols(route.GRPPathType),
							},
							[]actions.Action{
								actions.NewRejectAction(),
							},
						),
						filter.NewTerm(
							"accept & set NH",
							[]*filter.TermCondition{},
							[]actions.Action{
								actions.NewSetNextHopAction(GRPNhIP),
								actions.NewAcceptAction(),
							}),
					}),
			},
			toBGP: true,
			toGRP: true,
			addRoutesToGRPAdjRIBIn: []*route.Route{
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(11, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.GRPPathType,
					GRPPath: &route.GRPPath{
						NextHop:  bnet.IPv4FromOctets(1, 2, 3, 4).Ptr(),
						MetaData: GRPMetaData,
					},
				}),
			},
			expecxtedRoutesLocRIB: []*route.Route{
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(11, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.GRPPathType,
					GRPPath: &route.GRPPath{
						NextHop:  bnet.IPv4FromOctets(1, 2, 3, 4).Ptr(),
						MetaData: GRPMetaData,
					},
				}),
			},
			expecxtedRoutesGRPAdjRIBOut: []*route.Route{},
			expecxtedRoutesBGPAdjRIBOut: []*route.Route{
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(11, 0, 0, 0), 8).Ptr(), &route.Path{
					Type:              route.BGPPathType,
					RedistributedFrom: route.GRPPathType,
					GRPPath: &route.GRPPath{
						NextHop:  bnet.IPv4FromOctets(1, 2, 3, 4).Ptr(),
						MetaData: GRPMetaData,
					},
					BGPPath: &route.BGPPath{
						ASPath:    types.NewASPath([]uint32{57165}),
						ASPathLen: 1,
						BGPPathA: &route.BGPPathA{
							Source:  bnet.IPv4(0).Ptr(),
							NextHop: &localBIP,
						},
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

		// Set up an AdjRIBIn, associated with peer A, and register the LocRIB as a client
		bgpAdjRIBIn := BGPAdjRIBIn.New(test.BGPadjRIBInImportFilter, vrf, sessionAttrsA)
		bgpAdjRIBIn.Register(locRIB)

		// If we're sending to BGP set up AdjRIBOut and register it as a client to LocRIB
		var bgpAdjRIBOut *BGPAdjRIBOut.AdjRIBOut
		if test.toBGP {
			bgpAdjRIBOut = BGPAdjRIBOut.New(locRIB, test.BGPadjRibOutSessionAttrs, test.BGPadjRIBOutExportFilter)
			locRIB.Register(bgpAdjRIBOut)
		}

		// Set up an AdjRIBIn and register the LocRIB as a client
		grpAdjRIBIn := GRPAdjRibIn.New(filter.NewAcceptAllFilterChain(), vrf)
		grpAdjRIBIn.Register(locRIB)

		// If we're sending to GRP set up AdjRIBOut and register it as a client to LocRIB
		var grpAdjRibOut *GRPAdjRibOut.AdjRIBOut
		if test.toGRP {
			grpAdjRibOut = GRPAdjRibOut.New(locRIB, test.GRPadjRibInImportFilter, GRPMetaData)
			locRIB.Register(grpAdjRibOut)
		}

		for _, r := range test.addRoutesToBGPAdjRIBIn {
			for _, p := range r.Paths() {
				bgpAdjRIBIn.AddPath(r.Prefix(), p)
			}
		}

		for _, r := range test.addRoutesToGRPAdjRIBIn {
			for _, p := range r.Paths() {
				grpAdjRIBIn.AddPath(r.Prefix(), p)
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

		if test.toBGP {
			assert.Equal(t, test.expecxtedRoutesBGPAdjRIBOut, bgpAdjRIBOut.Dump(), fmt.Sprintf("BGP AdjRIBOut does not contain expected routes for test %q", test.name))
		}

		if test.toGRP {
			assert.Equal(t, test.expecxtedRoutesGRPAdjRIBOut, grpAdjRibOut.Dump(), fmt.Sprintf("GRP AdjRIBOut does not contain expected routes for test %q", test.name))
		}
	}
}
