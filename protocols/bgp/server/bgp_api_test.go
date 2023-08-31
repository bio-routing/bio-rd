package server

import (
	"context"
	"io"
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
	vrf.GetGlobalRegistry().CreateVRFIfNotExists(vrf.DefaultVRFName, 0)

	sessionAttrs := routingtable.SessionAttrs{
		RouterID:  0,
		ClusterID: 0,
		LocalASN:  42,
		PeerASN:   42,
		AddPathRX: true,
		AddPathTX: true,
	}

	v, _ := vrf.New("vrf0", 0)

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
						peers: map[PeerKey]*peer{},
					},
				},
				vrfReg: vrf.GetGlobalRegistry(),
			},
			addRoutes: []*route.Route{},
			req: &api.DumpRIBRequest{
				Peer: bnet.IPv4FromOctets(10, 0, 0, 0).ToProto(),
				Afi:  packet.AFIIPv4,
				Safi: packet.SAFIUnicast,
			},
			expected: []*routeapi.Route{},
			wantFail: true,
		},
		{
			name: "Test #1: No routes given",
			apisrv: &BGPAPIServer{
				srv: &bgpServer{
					peers: &peerManager{
						peers: map[PeerKey]*peer{
							{
								vrf:        vrf.GetGlobalRegistry().GetVRFByName(vrf.DefaultVRFName),
								neighborIP: bnet.IPv4FromOctets(10, 0, 0, 0).Dedup(),
							}: {
								fsms: []*FSM{
									0: {
										ipv4Unicast: &fsmAddressFamily{
											adjRIBIn:  adjRIBIn.New(filter.NewAcceptAllFilterChain(), v, sessionAttrs),
											adjRIBOut: adjRIBOut.New(nil, routingtable.SessionAttrs{Type: route.BGPPathType}, filter.NewAcceptAllFilterChain()),
										},
									},
								},
							},
						},
					},
				},
				vrfReg: vrf.GetGlobalRegistry(),
			},
			addRoutes: []*route.Route{},
			req: &api.DumpRIBRequest{
				Peer: bnet.IPv4FromOctets(10, 0, 0, 0).ToProto(),
				Afi:  packet.AFIIPv4,
				Safi: packet.SAFIUnicast,
			},
			expected: []*routeapi.Route{},
			wantFail: false,
		},
		{
			name: "Test #2: One simple routes given",
			apisrv: &BGPAPIServer{
				srv: &bgpServer{
					peers: &peerManager{
						peers: map[PeerKey]*peer{
							{
								vrf:        vrf.GetGlobalRegistry().GetVRFByName(vrf.DefaultVRFName),
								neighborIP: bnet.IPv4FromOctets(10, 0, 0, 0).Dedup(),
							}: {
								addr: bnet.IPv4(123).Ptr(),
								fsms: []*FSM{
									0: {
										ipv4Unicast: &fsmAddressFamily{
											adjRIBIn:  adjRIBIn.New(filter.NewAcceptAllFilterChain(), v, sessionAttrs),
											adjRIBOut: adjRIBOut.New(nil, routingtable.SessionAttrs{Type: route.BGPPathType, RouteServerClient: true, PeerIP: bnet.IPv4(0).Ptr()}, filter.NewAcceptAllFilterChain()),
										},
									},
								},
							},
						},
					},
				},
				vrfReg: vrf.GetGlobalRegistry(),
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
				Peer:    bnet.IPv4FromOctets(10, 0, 0, 0).ToProto(),
				Afi:     packet.AFIIPv4,
				Safi:    packet.SAFIUnicast,
				VrfName: vrf.DefaultVRFName,
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
			name: "Test #3: One complex route given",
			apisrv: &BGPAPIServer{
				srv: &bgpServer{
					peers: &peerManager{
						peers: map[PeerKey]*peer{
							{
								vrf:        vrf.GetGlobalRegistry().GetVRFByName(vrf.DefaultVRFName),
								neighborIP: bnet.IPv4FromOctets(10, 0, 0, 0).Dedup(),
							}: {
								addr: bnet.IPv4(123).Ptr(),
								fsms: []*FSM{
									0: {
										ipv4Unicast: &fsmAddressFamily{
											adjRIBIn:  adjRIBIn.New(filter.NewAcceptAllFilterChain(), v, sessionAttrs),
											adjRIBOut: adjRIBOut.New(nil, routingtable.SessionAttrs{Type: route.BGPPathType, RouteServerClient: true, PeerIP: bnet.IPv4(123).Ptr()}, filter.NewAcceptAllFilterChain()),
										},
									},
								},
							},
						},
					},
				},
				vrfReg: vrf.GetGlobalRegistry(),
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
				Afi:  packet.AFIIPv4,
				Safi: packet.SAFIUnicast,
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
				pk := PeerKey{
					vrf:        vrf.GetGlobalRegistry().GetVRFByName(vrf.DefaultVRFName),
					neighborIP: bnet.IPv4FromOctets(10, 0, 0, 0).Dedup(),
				}
				test.apisrv.srv.(*bgpServer).peers.peers[pk].fsms[0].ipv4Unicast.adjRIBIn.AddPath(r.Prefix(), p)
			}
		}

		bufSize := 1024 * 1024
		lis := bufconn.Listen(bufSize)
		s := grpc.NewServer()
		api.RegisterBgpServiceServer(s, test.apisrv)
		go func() {
			if err := s.Serve(lis); err != nil {
				t.Logf("Server exited with error: %v", err)
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
				if err != io.EOF && !test.wantFail {
					t.Fatalf("stream RPC ended with non EOF error for test %q: %v", test.name, err)
				}

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
				pk := PeerKey{
					vrf:        vrf.GetGlobalRegistry().GetVRFByName(vrf.DefaultVRFName),
					neighborIP: bnet.IPv4FromOctets(10, 0, 0, 0).Dedup(),
				}
				test.apisrv.srv.(*bgpServer).peers.peers[pk].fsms[0].ipv4Unicast.adjRIBOut.AddPath(r.Prefix(), p)
			}
		}

		bufSize := 1024 * 1024
		lis := bufconn.Listen(bufSize)
		s := grpc.NewServer()
		api.RegisterBgpServiceServer(s, test.apisrv)
		go func() {
			if err := s.Serve(lis); err != nil {
				t.Logf("Server exited with error: %v", err)
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
