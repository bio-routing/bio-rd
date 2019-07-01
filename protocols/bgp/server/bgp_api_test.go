package server

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/bio-routing/bio-rd/protocols/bgp/api"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/route"
	routeapi "github.com/bio-routing/bio-rd/route/api"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/adjRIBIn"
	"github.com/bio-routing/bio-rd/routingtable/adjRIBOut"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	bnet "github.com/bio-routing/bio-rd/net"
)

func TestDumpRIBInOut(t *testing.T) {
	tests := []struct {
		name      string
		apisrv    *BGPAPIServer
		addRoutes []*route.Route
		req       *api.DumpRIBRequest
		expected  []*routeapi.Route
		wantFail  bool
	}{
		{
			name: "Test #0: Non existent peer",
			apisrv: &BGPAPIServer{
				srv: &bgpServer{
					peers: &peerManager{
						peers: map[bnet.IP]*peer{},
					},
				},
			},
			addRoutes: []*route.Route{},
			req: &api.DumpRIBRequest{
				Peer: bnet.IPv4FromOctets(10, 0, 0, 0).ToProto(),
				Afi:  packet.IPv4AFI,
				Safi: packet.UnicastSAFI,
			},
			expected: []*routeapi.Route{},
			wantFail: false,
		},
		{
			name: "Test #1: No routes given",
			apisrv: &BGPAPIServer{
				srv: &bgpServer{
					peers: &peerManager{
						peers: map[bnet.IP]*peer{
							bnet.IPv4FromOctets(10, 0, 0, 0): {
								fsms: []*FSM{
									0: {
										ipv4Unicast: &fsmAddressFamily{
											adjRIBIn:  adjRIBIn.New(filter.NewAcceptAllFilter(), nil, 0, 0, true),
											adjRIBOut: adjRIBOut.New(&routingtable.Neighbor{Type: route.BGPPathType}, filter.NewAcceptAllFilter(), true),
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
			expected: []*routeapi.Route{},
			wantFail: false,
		},
		{
			name: "Test #2: One simple routes given",
			apisrv: &BGPAPIServer{
				srv: &bgpServer{
					peers: &peerManager{
						peers: map[bnet.IP]*peer{
							bnet.IPv4FromOctets(10, 0, 0, 0): {
								fsms: []*FSM{
									0: {
										ipv4Unicast: &fsmAddressFamily{
											adjRIBIn:  adjRIBIn.New(filter.NewAcceptAllFilter(), nil, 0, 0, true),
											adjRIBOut: adjRIBOut.New(&routingtable.Neighbor{Type: route.BGPPathType, RouteServerClient: true}, filter.NewAcceptAllFilter(), false),
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
			expected: []*routeapi.Route{
				{
					Pfx: bnet.NewPfx(bnet.IPv4FromOctets(20, 0, 0, 0), 16).ToProto(),
					Paths: []*routeapi.Path{
						{
							Type: routeapi.Path_BGP,
							BgpPath: &routeapi.BGPPath{
								OriginatorId:      1,
								NextHop:           bnet.IPv4FromOctets(100, 100, 100, 100).ToProto(),
								Source:            bnet.IPv4FromOctets(100, 100, 100, 100).ToProto(),
								AsPath:            nil,
								Communities:       nil,
								LargeCommunities:  nil,
								UnknownAttributes: nil,
								ClusterList:       nil,
							},
						},
					},
				},
			},
			wantFail: false,
		},
		{
			name: "Test #3: One complex routes given",
			apisrv: &BGPAPIServer{
				srv: &bgpServer{
					peers: &peerManager{
						peers: map[bnet.IP]*peer{
							bnet.IPv4FromOctets(10, 0, 0, 0): {
								fsms: []*FSM{
									0: {
										ipv4Unicast: &fsmAddressFamily{
											adjRIBIn:  adjRIBIn.New(filter.NewAcceptAllFilter(), routingtable.NewContributingASNs(), 0, 0, true),
											adjRIBOut: adjRIBOut.New(&routingtable.Neighbor{Type: route.BGPPathType, RouteServerClient: true}, filter.NewAcceptAllFilter(), false),
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
						ASPath: types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{15169, 3320},
							},
						},
						Communities: []uint32{100, 200, 300},
						LargeCommunities: []types.LargeCommunity{
							{
								GlobalAdministrator: 1,
								DataPart1:           2,
								DataPart2:           3,
							},
						},
						LocalPref: 1000,
						MED:       2000,
						UnknownAttributes: []types.UnknownPathAttribute{
							{
								Optional:   true,
								Transitive: true,
								Partial:    true,
								TypeCode:   222,
								Value:      []byte{0xff, 0xff},
							},
						},
						ClusterList: []uint32{},
					},
				}),
			},
			req: &api.DumpRIBRequest{
				Peer: bnet.IPv4FromOctets(10, 0, 0, 0).ToProto(),
				Afi:  packet.IPv4AFI,
				Safi: packet.UnicastSAFI,
			},
			expected: []*routeapi.Route{
				{
					Pfx: bnet.NewPfx(bnet.IPv4FromOctets(20, 0, 0, 0), 16).ToProto(),
					Paths: []*routeapi.Path{
						{
							Type: routeapi.Path_BGP,
							BgpPath: &routeapi.BGPPath{
								OriginatorId: 1,
								LocalPref:    1000,
								Med:          2000,
								NextHop:      bnet.IPv4FromOctets(100, 100, 100, 100).ToProto(),
								Source:       bnet.IPv4FromOctets(100, 100, 100, 100).ToProto(),
								AsPath: []*routeapi.ASPathSegment{
									{
										AsSequence: true,
										Asns:       []uint32{15169, 3320},
									},
								},
								Communities: []uint32{100, 200, 300},
								LargeCommunities: []*routeapi.LargeCommunity{
									{
										GlobalAdministrator: 1,
										DataPart1:           2,
										DataPart2:           3,
									},
								},
								ClusterList: nil,
								UnknownAttributes: []*routeapi.UnknownPathAttribute{
									{
										Optional:   true,
										Transitive: true,
										Partial:    true,
										TypeCode:   222,
										Value:      []byte{0xff, 0xff},
									},
								},
							},
						},
					},
				},
			},
			wantFail: false,
		},
	}

	// Test RIBin
	for _, test := range tests {
		for _, r := range test.addRoutes {
			for _, p := range r.Paths() {
				test.apisrv.srv.(*bgpServer).peers.peers[bnet.IPv4FromOctets(10, 0, 0, 0)].fsms[0].ipv4Unicast.adjRIBIn.AddPath(r.Prefix(), p)
			}
		}

		bufSize := 1024 * 1024
		lis := bufconn.Listen(bufSize)
		s := grpc.NewServer()
		api.RegisterBgpServiceServer(s, test.apisrv)
		go func() {
			if err := s.Serve(lis); err != nil {
				log.Fatalf("Server exited with error: %v", err)
			}
		}()

		ctx := context.Background()
		conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithDialer(func(string, time.Duration) (net.Conn, error) {
			return lis.Dial()
		}), grpc.WithInsecure())
		if err != nil {
			t.Fatalf("Failed to dial bufnet: %v", err)
		}
		defer conn.Close()

		client := api.NewBgpServiceClient(conn)
		streamClient, err := client.DumpRIBIn(ctx, test.req)
		if err != nil {
			t.Fatalf("AdjRIBInStream client call failed: %v", err)
		}

		res := make([]*routeapi.Route, 0)
		for {
			r, err := streamClient.Recv()
			if err != nil {
				break
			}

			res = append(res, r)
		}

		assert.Equal(t, test.expected, res, test.name)
	}

	// Test RIBout
	for _, test := range tests {
		for _, r := range test.addRoutes {
			for _, p := range r.Paths() {
				test.apisrv.srv.(*bgpServer).peers.peers[bnet.IPv4FromOctets(10, 0, 0, 0)].fsms[0].ipv4Unicast.adjRIBOut.AddPath(r.Prefix(), p)
			}
		}

		bufSize := 1024 * 1024
		lis := bufconn.Listen(bufSize)
		s := grpc.NewServer()
		api.RegisterBgpServiceServer(s, test.apisrv)
		go func() {
			if err := s.Serve(lis); err != nil {
				log.Fatalf("Server exited with error: %v", err)
			}
		}()

		ctx := context.Background()
		conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithDialer(func(string, time.Duration) (net.Conn, error) {
			return lis.Dial()
		}), grpc.WithInsecure())
		if err != nil {
			t.Fatalf("Failed to dial bufnet: %v", err)
		}
		defer conn.Close()

		client := api.NewBgpServiceClient(conn)
		streamClient, err := client.DumpRIBOut(ctx, test.req)
		if err != nil {
			t.Fatalf("AdjRIBInStream client call failed: %v", err)
		}

		res := make([]*routeapi.Route, 0)
		for {
			r, err := streamClient.Recv()
			if err != nil {
				break
			}

			res = append(res, r)
		}

		assert.Equal(t, test.expected, res, test.name)
	}
}
