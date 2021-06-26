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
	"github.com/bio-routing/bio-rd/routingtable/vrf"
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
											adjRIBIn:  adjRIBIn.New(filter.NewAcceptAllFilterChain(), nil, 0, 0, true),
											adjRIBOut: adjRIBOut.New(nil, &routingtable.Neighbor{Type: route.BGPPathType}, filter.NewAcceptAllFilterChain(), true),
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
								addr: bnet.IPv4(123).Ptr(),
								fsms: []*FSM{
									0: {
										ipv4Unicast: &fsmAddressFamily{
											adjRIBIn:  adjRIBIn.New(filter.NewAcceptAllFilterChain(), nil, 0, 0, true),
											adjRIBOut: adjRIBOut.New(nil, &routingtable.Neighbor{Type: route.BGPPathType, RouteServerClient: true, Address: bnet.IPv4(0).Ptr()}, filter.NewAcceptAllFilterChain(), false),
										},
									},
								},
							},
						},
					},
				},
			},
			addRoutes: []*route.Route{
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(20, 0, 0, 0), 16).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							OriginatorID: 1,
							NextHop:      bnet.IPv4FromOctets(100, 100, 100, 100).Ptr(),
							Source:       bnet.IPv4FromOctets(100, 100, 100, 100).Ptr(),
						},
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
								addr: bnet.IPv4(123).Ptr(),
								fsms: []*FSM{
									0: {
										ipv4Unicast: &fsmAddressFamily{
											adjRIBIn:  adjRIBIn.New(filter.NewAcceptAllFilterChain(), routingtable.NewContributingASNs(), 0, 0, true),
											adjRIBOut: adjRIBOut.New(nil, &routingtable.Neighbor{Type: route.BGPPathType, RouteServerClient: true, Address: bnet.IPv4(123).Ptr()}, filter.NewAcceptAllFilterChain(), false),
										},
									},
								},
							},
						},
					},
				},
			},
			addRoutes: []*route.Route{
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(20, 0, 0, 0), 16).Ptr(), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							OriginatorID: 1,
							NextHop:      bnet.IPv4FromOctets(100, 100, 100, 100).Ptr(),
							Source:       bnet.IPv4FromOctets(100, 100, 100, 100).Ptr(),
							LocalPref:    1000,
							MED:          2000,
						},
						ASPath: &types.ASPath{
							types.ASPathSegment{
								Type: types.ASSequence,
								ASNs: []uint32{15169, 3320},
							},
						},
						Communities: &types.Communities{100, 200, 300},
						LargeCommunities: &types.LargeCommunities{
							{
								GlobalAdministrator: 1,
								DataPart1:           2,
								DataPart2:           3,
							},
						},
						UnknownAttributes: []types.UnknownPathAttribute{
							{
								Optional:   true,
								Transitive: true,
								Partial:    true,
								TypeCode:   222,
								Value:      []byte{0xff, 0xff},
							},
						},
						ClusterList: &types.ClusterList{},
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

		expected := make([]string, 0)
		for _, exp := range test.expected {
			expected = append(expected, exp.String())
		}

		results := make([]string, 0)
		for _, r := range res {
			results = append(results, r.String())
		}
		assert.Equal(t, expected, results, test.name)
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

		expected := make([]string, 0)
		for _, exp := range test.expected {
			expected = append(expected, exp.String())
		}

		results := make([]string, 0)
		for _, r := range res {
			results = append(results, r.String())
		}
		assert.Equal(t, expected, results, test.name)
	}
}

func TestListSessions(t *testing.T) {
	vrf1, _ := vrf.New("majestic-cat", 0)
	vrf2, _ := vrf.New("glorious-seagull", 1)

	tests := []struct {
		name     string
		apisrv   *BGPAPIServer
		req      *api.ListSessionsRequest
		expected *api.ListSessionsResponse
		wantFail bool
	}{
		{
			name: "Simple ListSessions, without filter",
			apisrv: &BGPAPIServer{
				srv: &bgpServer{
					peers: &peerManager{
						peers: map[bnet.IP]*peer{
							bnet.IPv4FromOctets(10, 0, 0, 0): {
								config: &PeerConfig{
									PeerAS:       65100,
									LocalAS:      65000,
									LocalAddress: bnet.IPv4FromOctets(172, 0, 0, 0).Ptr(),
									PeerAddress:  bnet.IPv4FromOctets(10, 0, 0, 0).Ptr(),
								},
								peerASN:  65100,
								localASN: 65000,
								addr:     bnet.IPv4FromOctets(10, 0, 0, 0).Ptr(),
								vrf:      vrf1,
							},
						},
					},
				},
			},
			req: &api.ListSessionsRequest{},
			expected: &api.ListSessionsResponse{
				Sessions: []*api.Session{
					{
						LocalAddress:    bnet.IPv4FromOctets(172, 0, 0, 0).ToProto(),
						NeighborAddress: bnet.IPv4FromOctets(10, 0, 0, 0).ToProto(),
						PeerAsn:         65100,
						LocalAsn:        65000,
						Status:          api.Session_Active,
						Stats:           &api.SessionStats{},
						VrfName:         "majestic-cat",
					},
				},
			},
			wantFail: false,
		},
		{
			name: "ListSessions with two peers without filter",
			apisrv: &BGPAPIServer{
				srv: &bgpServer{
					peers: &peerManager{
						peers: map[bnet.IP]*peer{
							bnet.IPv4FromOctets(10, 0, 0, 0): {
								config: &PeerConfig{
									PeerAS:       65100,
									LocalAS:      65000,
									LocalAddress: bnet.IPv4FromOctets(172, 0, 0, 0).Ptr(),
									PeerAddress:  bnet.IPv4FromOctets(10, 0, 0, 0).Ptr(),
								},
								peerASN:  65100,
								localASN: 65000,
								addr:     bnet.IPv4FromOctets(10, 0, 0, 0).Ptr(),
								vrf:      vrf1,
							},
							bnet.IPv4FromOctets(192, 168, 0, 0): {
								config: &PeerConfig{
									PeerAS:       64999,
									LocalAS:      65000,
									LocalAddress: bnet.IPv4FromOctets(172, 0, 0, 0).Ptr(),
									PeerAddress:  bnet.IPv4FromOctets(192, 168, 0, 0).Ptr(),
								},
								peerASN:  64999,
								localASN: 65000,
								addr:     bnet.IPv4FromOctets(192, 168, 0, 0).Ptr(),
								vrf:      vrf1,
							},
						},
					},
				},
			},
			req: &api.ListSessionsRequest{},
			expected: &api.ListSessionsResponse{
				Sessions: []*api.Session{
					{
						LocalAddress:    bnet.IPv4FromOctets(172, 0, 0, 0).ToProto(),
						NeighborAddress: bnet.IPv4FromOctets(10, 0, 0, 0).ToProto(),
						PeerAsn:         65100,
						LocalAsn:        65000,
						Status:          api.Session_Active,
						Stats:           &api.SessionStats{},
						VrfName:         "majestic-cat",
					},
					{
						LocalAddress:    bnet.IPv4FromOctets(172, 0, 0, 0).ToProto(),
						NeighborAddress: bnet.IPv4FromOctets(192, 168, 0, 0).ToProto(),
						PeerAsn:         64999,
						LocalAsn:        65000,
						Status:          api.Session_Active,
						Stats:           &api.SessionStats{},
						VrfName:         "majestic-cat",
					},
				},
			},
			wantFail: false,
		},
		{
			name: "ListSession with two peers and filter for vrf",
			apisrv: &BGPAPIServer{
				srv: &bgpServer{
					peers: &peerManager{
						peers: map[bnet.IP]*peer{
							bnet.IPv4FromOctets(10, 0, 0, 0): {
								config: &PeerConfig{
									PeerAS:       65100,
									LocalAS:      65000,
									LocalAddress: bnet.IPv4FromOctets(172, 0, 0, 0).Ptr(),
									PeerAddress:  bnet.IPv4FromOctets(10, 0, 0, 0).Ptr(),
								},
								peerASN:  65100,
								localASN: 65000,
								addr:     bnet.IPv4FromOctets(10, 0, 0, 0).Ptr(),
								vrf:      vrf2,
							},
							bnet.IPv4FromOctets(192, 168, 0, 0): {
								config: &PeerConfig{
									PeerAS:       64999,
									LocalAS:      65000,
									LocalAddress: bnet.IPv4FromOctets(172, 0, 0, 0).Ptr(),
									PeerAddress:  bnet.IPv4FromOctets(192, 168, 0, 0).Ptr(),
								},
								peerASN:  64999,
								localASN: 65000,
								addr:     bnet.IPv4FromOctets(192, 168, 0, 0).Ptr(),
								vrf:      vrf1,
							},
						},
					},
				},
			},
			req: &api.ListSessionsRequest{
				Filter: &api.SessionFilter{
					VrfName: "glorious-seagull",
				},
			},
			expected: &api.ListSessionsResponse{
				Sessions: []*api.Session{
					{
						LocalAddress:    bnet.IPv4FromOctets(172, 0, 0, 0).ToProto(),
						NeighborAddress: bnet.IPv4FromOctets(10, 0, 0, 0).ToProto(),
						PeerAsn:         65100,
						LocalAsn:        65000,
						Status:          api.Session_Active,
						Stats:           &api.SessionStats{},
						VrfName:         "glorious-seagull",
					},
				},
			},
			wantFail: false,
		},
		{
			name: "ListSession with two peers and filter for neighbor",
			apisrv: &BGPAPIServer{
				srv: &bgpServer{
					peers: &peerManager{
						peers: map[bnet.IP]*peer{
							bnet.IPv4FromOctets(10, 0, 0, 0): {
								config: &PeerConfig{
									PeerAS:       65100,
									LocalAS:      65000,
									LocalAddress: bnet.IPv4FromOctets(172, 0, 0, 0).Ptr(),
									PeerAddress:  bnet.IPv4FromOctets(10, 0, 0, 0).Ptr(),
								},
								peerASN:  65100,
								localASN: 65000,
								addr:     bnet.IPv4FromOctets(10, 0, 0, 0).Ptr(),
								vrf:      vrf1,
							},
							bnet.IPv4FromOctets(192, 168, 0, 0): {
								config: &PeerConfig{
									PeerAS:       64999,
									LocalAS:      65000,
									LocalAddress: bnet.IPv4FromOctets(172, 0, 0, 0).Ptr(),
									PeerAddress:  bnet.IPv4FromOctets(192, 168, 0, 0).Ptr(),
								},
								peerASN:  64999,
								localASN: 65000,
								addr:     bnet.IPv4FromOctets(192, 168, 0, 0).Ptr(),
								vrf:      vrf1,
							},
						},
					},
				},
			},
			req: &api.ListSessionsRequest{
				Filter: &api.SessionFilter{
					NeighborIp: bnet.IPv4FromOctets(10, 0, 0, 0).ToProto(),
				},
			},
			expected: &api.ListSessionsResponse{
				Sessions: []*api.Session{
					{
						LocalAddress:    bnet.IPv4FromOctets(172, 0, 0, 0).ToProto(),
						NeighborAddress: bnet.IPv4FromOctets(10, 0, 0, 0).ToProto(),
						PeerAsn:         65100,
						LocalAsn:        65000,
						Status:          api.Session_Active,
						Stats:           &api.SessionStats{},
						VrfName:         "majestic-cat",
					},
				},
			},
			wantFail: false,
		},
		{
			name: "ListSession with routes for stats",
			apisrv: &BGPAPIServer{
				srv: &bgpServer{
					peers: &peerManager{
						peers: map[bnet.IP]*peer{
							bnet.IPv4FromOctets(10, 0, 0, 0): {
								ipv4: &peerAddressFamily{},
								ipv6: &peerAddressFamily{},
								fsms: []*FSM{
									0: {
										ribsInitialized: true,
										ipv4Unicast: &fsmAddressFamily{
											adjRIBIn:  &routingtable.RTMockClient{FakeRouteCount: 3},
											adjRIBOut: &routingtable.RTMockClient{FakeRouteCount: 2},
										},
										ipv6Unicast: &fsmAddressFamily{
											adjRIBIn:  &routingtable.RTMockClient{FakeRouteCount: 10},
											adjRIBOut: &routingtable.RTMockClient{FakeRouteCount: 12},
										},
										counters: fsmCounters{
											updatesReceived: 23,
											updatesSent:     42,
										},
									},
								},
								config: &PeerConfig{
									PeerAS:       65100,
									LocalAS:      65000,
									LocalAddress: bnet.IPv4FromOctets(172, 0, 0, 0).Ptr(),
									PeerAddress:  bnet.IPv4FromOctets(10, 0, 0, 0).Ptr(),
								},
								peerASN:  65100,
								localASN: 65000,
								addr:     bnet.IPv4FromOctets(10, 0, 0, 0).Ptr(),
								vrf:      vrf1,
							},
						},
					},
				},
			},
			req: &api.ListSessionsRequest{},
			expected: &api.ListSessionsResponse{
				Sessions: []*api.Session{
					{
						LocalAddress:    bnet.IPv4FromOctets(172, 0, 0, 0).ToProto(),
						NeighborAddress: bnet.IPv4FromOctets(10, 0, 0, 0).ToProto(),
						PeerAsn:         65100,
						LocalAsn:        65000,
						Status:          api.Session_Active,
						Stats: &api.SessionStats{
							RoutesReceived: 13,
							RoutesExported: 14,
							MessagesIn:     23,
							MessagesOut:    42,
						},
						VrfName: "majestic-cat",
					},
				},
			},
			wantFail: false,
		},
	}

	for _, test := range tests {
		testSrv := test.apisrv.srv.(*bgpServer)
		testSrv.metrics = &metricsService{testSrv}
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
		neighborResp, err := client.ListSessions(ctx, test.req)
		if err != nil {
			t.Fatalf("ListSessions call failed: %v", err)
		}
		assert.Equal(t, test.expected, neighborResp)
	}

	// As tests seem to share state we need to clean up the vrf here
	vrf1.Unregister()
	vrf1.Dispose()
	vrf2.Unregister()
	vrf2.Dispose()
}
