package server

import (
	"context"
	"testing"

	"github.com/bio-routing/bio-rd/protocols/bgp/api"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/route"
	routeapi "github.com/bio-routing/bio-rd/route/api"
	"github.com/bio-routing/bio-rd/routingtable/adjRIBIn"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/stretchr/testify/assert"

	bnet "github.com/bio-routing/bio-rd/net"
)

func TestDumpRIBIn(t *testing.T) {
	tests := []struct {
		name      string
		apisrv    *BGPAPIServer
		addRoutes []*route.Route
		req       *api.DumpRIBRequest
		expected  *api.DumpRIBResponse
		wantFail  bool
	}{
		{
			name: "Test #1: No routes given",
			apisrv: &BGPAPIServer{
				srv: &bgpServer{
					peers: &peerManager{
						peers: map[bnet.IP]*peer{
							bnet.IPv4FromOctets(10, 0, 0, 0): {
								fsms: []*FSM{
									0: &FSM{
										ipv4Unicast: &fsmAddressFamily{
											adjRIBIn: adjRIBIn.New(filter.NewAcceptAllFilter(), nil, 0, 0, true),
										},
									},
								},
							},
						},
					},
				},
			},
			addRoutes: []*route.Route{},
			req: &api.DumpRIBRequest{
				Peer: bnet.IPv4FromOctets(10, 0, 0, 0).ToProto(),
				Afi:  packet.IPv4AFI,
				Safi: packet.UnicastSAFI,
			},
			expected: &api.DumpRIBResponse{
				Routes: []*routeapi.Route{},
			},
			wantFail: false,
		},
		{
			name: "Test #2: One routes given",
			apisrv: &BGPAPIServer{
				srv: &bgpServer{
					peers: &peerManager{
						peers: map[bnet.IP]*peer{
							bnet.IPv4FromOctets(10, 0, 0, 0): {
								fsms: []*FSM{
									0: &FSM{
										ipv4Unicast: &fsmAddressFamily{
											adjRIBIn: adjRIBIn.New(filter.NewAcceptAllFilter(), nil, 0, 0, true),
										},
									},
								},
							},
						},
					},
				},
			},
			addRoutes: []*route.Route{
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(20, 0, 0, 0), 16), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						OriginatorID: 1,
						NextHop:      bnet.IPv4FromOctets(100, 100, 100, 100),
						Source:       bnet.IPv4FromOctets(100, 100, 100, 100),
					},
				}),
			},
			req: &api.DumpRIBRequest{
				Peer: bnet.IPv4FromOctets(10, 0, 0, 0).ToProto(),
				Afi:  packet.IPv4AFI,
				Safi: packet.UnicastSAFI,
			},
			expected: &api.DumpRIBResponse{
				Routes: []*routeapi.Route{
					{
						Pfx: bnet.NewPfx(bnet.IPv4FromOctets(20, 0, 0, 0), 16).ToProto(),
						Paths: []*routeapi.Path{
							{
								Type: routeapi.Path_BGP,
								BGPPath: &routeapi.BGPPath{
									OriginatorId:      1,
									NextHop:           bnet.IPv4FromOctets(100, 100, 100, 100).ToProto(),
									Source:            bnet.IPv4FromOctets(100, 100, 100, 100).ToProto(),
									ASPath:            []*routeapi.ASPathSegment{},
									Communities:       []uint32{},
									LargeCommunities:  []*routeapi.LargeCommunity{},
									UnknownAttributes: []*routeapi.UnknownPathAttribute{},
									ClusterList:       []uint32{},
								},
							},
						},
					},
				},
			},
			wantFail: false,
		},
	}

	for _, test := range tests {
		for _, r := range test.addRoutes {
			for _, p := range r.Paths() {
				test.apisrv.srv.(*bgpServer).peers.peers[bnet.IPv4FromOctets(10, 0, 0, 0)].fsms[0].ipv4Unicast.adjRIBIn.AddPath(r.Prefix(), p)
			}
		}

		res, err := test.apisrv.DumpRIBIn(context.Background(), test.req)
		if err != nil {
			if test.wantFail {
				continue
			}

			t.Errorf("Unexpected failure for %q: %v", test.name, err)
			continue
		}

		if test.wantFail {
			t.Errorf("Unexpected success for test %q", test.name)
			continue
		}

		assert.Equal(t, test.expected, res, test.name)
	}
}
