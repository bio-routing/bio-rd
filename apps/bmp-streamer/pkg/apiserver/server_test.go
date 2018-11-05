package apiserver

import (
	"context"
	"log"
	"net"
	"sync"
	"testing"
	"time"

	pb "github.com/bio-routing/bio-rd/apps/bmp-streamer/pkg/bmpstreamer"
	bionet "github.com/bio-routing/bio-rd/net"
	apinet "github.com/bio-routing/bio-rd/net/api"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/route"
	apiroute "github.com/bio-routing/bio-rd/route/api"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestUpdateNewClient(t *testing.T) {
	r := &ribClient{}
	l := locRIB.New()
	r.UpdateNewClient(l)
}

func TestRegister(t *testing.T) {
	r := &ribClient{}
	l := locRIB.New()

	r.Register(l)
	r.RegisterWithOptions(l, routingtable.ClientOptions{})

	r.Unregister(l)
}

func TestRouteCount(t *testing.T) {
	r := &ribClient{}
	res := r.RouteCount()
	assert.Equal(t, int64(-1), res)
}

func TestNew(t *testing.T) {
	b := &server.BMPServer{}
	s := New(b)

	expected := &APIServer{
		bmpServer: b,
	}

	assert.Equal(t, expected, s)
}

func TestUpdateToRIBUpdate(t *testing.T) {
	tests := []struct {
		name     string
		u        update
		expected *pb.RIBUpdate
	}{
		{
			name: "Basics advert.",
			u: update{
				advertisement: true,
				route: route.NewRoute(bionet.NewPfx(bionet.IPv4(200), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						PathIdentifier: 10,
						NextHop:        bionet.IPv4(210),
						LocalPref:      20,
						ASPath: types.ASPath{
							{
								Type: types.ASSequence,
								ASNs: []uint32{100, 200, 300},
							},
						},
						Origin:        1,
						MED:           1000,
						EBGP:          true,
						BGPIdentifier: 1337,
						Source:        bionet.IPv4(220),
						Communities:   []uint32{10000, 20000},
						LargeCommunities: []types.LargeCommunity{
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
								Value:      []byte{1, 1, 1, 1},
							},
						},
						OriginatorID: 5,
						ClusterList:  []uint32{3, 4, 5},
					},
				}),
			},
			expected: &pb.RIBUpdate{
				Peer: &apinet.IP{
					Lower:   220,
					Version: apinet.IP_IPv4,
				},
				Advertisement: true,
				Route: &apiroute.Route{
					Pfx: &apinet.Prefix{
						Address: &apinet.IP{
							Lower:   200,
							Version: apinet.IP_IPv4,
						},
						Pfxlen: 8,
					},
					Paths: []*apiroute.Path{
						{
							Type: apiroute.Path_BGP,
							BGPPath: &apiroute.BGPPath{
								PathIdentifier: 10,
								NextHop: &apinet.IP{
									Lower:   210,
									Version: apinet.IP_IPv4,
								},
								LocalPref: 20,
								ASPath: []*apiroute.ASPathSegment{
									{
										ASSequence: true,
										ASNs:       []uint32{100, 200, 300},
									},
								},
								Origin:        1,
								MED:           1000,
								EBGP:          true,
								BGPIdentifier: 1337,
								Source: &apinet.IP{
									Lower:   220,
									Version: apinet.IP_IPv4,
								},
								Communities: []uint32{10000, 20000},
								LargeCommunities: []*apiroute.LargeCommunity{
									{
										GlobalAdministrator: 1,
										DataPart1:           2,
										DataPart2:           3,
									},
								},
								ClusterList: []uint32{3, 4, 5},
								UnknownAttributes: []*apiroute.UnknownAttribute{
									{
										Optional:   true,
										Transitive: true,
										Partial:    true,
										TypeCode:   222,
										Value:      []byte{1, 1, 1, 1},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		res := test.u.toRIBUpdate()
		assert.Equal(t, test.expected, res, test.name)
	}
}

type mockBMPServer struct {
	RIB4 *locRIB.LocRIB
	RIB6 *locRIB.LocRIB
}

func newmockBMPServer() *mockBMPServer {
	return &mockBMPServer{
		RIB4: locRIB.New(),
		RIB6: locRIB.New(),
	}
}

func (m *mockBMPServer) SubscribeRIBs(client routingtable.RouteTableClient, rtr net.IP, afi uint8) {
	switch afi {
	case packet.IPv4AFI:
		m.RIB4.Register(client)
	case packet.IPv6AFI:
		m.RIB6.Register(client)
	}
}

func (m *mockBMPServer) UnsubscribeRIBs(client routingtable.RouteTableClient, rtr net.IP, afi uint8) {
	switch afi {
	case packet.IPv4AFI:
		m.RIB4.Unregister(client)
	case packet.IPv6AFI:
		m.RIB6.Unregister(client)
	}
}

func TestIntegration(t *testing.T) {
	bmpSrv := newmockBMPServer()
	apiSrv := New(bmpSrv)

	bufSize := 1024 * 1024
	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterRIBServiceServer(s, apiSrv)
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

	nextPhase := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		bmpSrv.RIB4.AddPath(bionet.NewPfx(bionet.IPv4FromOctets(169, 254, 0, 0), 24), &route.Path{
			Type: route.BGPPathType,
			BGPPath: &route.BGPPath{
				LocalPref: 1337,
				NextHop:   bionet.IPv4FromOctets(10, 0, 0, 1),
				Source:    bionet.IPv4FromOctets(10, 0, 0, 2),
			},
		})

		bmpSrv.RIB6.AddPath(bionet.NewPfx(bionet.IPv6FromBlocks(0x2001, 0, 0, 0, 0, 0, 0, 0), 32), &route.Path{
			Type: route.BGPPathType,
			BGPPath: &route.BGPPath{
				LocalPref: 4242,
				NextHop:   bionet.IPv6FromBlocks(0x2001, 0, 0, 0, 0, 0, 0, 1),
				Source:    bionet.IPv6FromBlocks(0x2001, 0, 0, 0, 0, 0, 0, 2),
			},
		})

		<-nextPhase

		bmpSrv.RIB4.RemovePath(bionet.NewPfx(bionet.IPv4FromOctets(169, 254, 0, 0), 24), &route.Path{
			Type: route.BGPPathType,
			BGPPath: &route.BGPPath{
				LocalPref: 1337,
				NextHop:   bionet.IPv4FromOctets(10, 0, 0, 1),
				Source:    bionet.IPv4FromOctets(10, 0, 0, 2),
			},
		})

		/*bmpSrv.RIB4.AddPath(bionet.NewPfx(bionet.IPv4FromOctets(169, 254, 0, 1), 24), &route.Path{
			Type: route.BGPPathType,
			BGPPath: &route.BGPPath{
				LocalPref: 1337,
				NextHop:   bionet.IPv4FromOctets(10, 0, 0, 1),
				Source:    bionet.IPv4FromOctets(10, 0, 0, 2),
			},
		})*/
		wg.Done()
	}()

	client := pb.NewRIBServiceClient(conn)
	streamClient, err := client.AdjRIBInStream(ctx, &pb.AdjRIBInStreamRequest{
		Router: bionet.IPv4FromOctets(10, 0, 0, 1).ToProto(),
	})
	if err != nil {
		t.Fatalf("AdjRIBInStream client call failed: %v", err)
	}

	wg.Add(1)
	go func() {
		tests := []struct {
			name      string
			expected  *pb.RIBUpdate
			wantFail  bool
			nextPhase bool
		}{
			{
				name: "Read first IPv4 announcement",
				expected: &pb.RIBUpdate{
					Advertisement: true,
					Peer: &apinet.IP{
						Lower:   167772162,
						Version: apinet.IP_IPv4,
					},
					Route: &apiroute.Route{
						Pfx: &apinet.Prefix{
							Address: &apinet.IP{
								Lower:   2851995648,
								Version: apinet.IP_IPv4,
							},
							Pfxlen: 24,
						},
						Paths: []*apiroute.Path{
							{
								Type: apiroute.Path_BGP,
								BGPPath: &apiroute.BGPPath{
									LocalPref: 1337,
									NextHop: &apinet.IP{
										Version: apinet.IP_IPv4,
										Lower:   167772161,
									},
									Source: &apinet.IP{
										Version: apinet.IP_IPv4,
										Lower:   167772162,
									},
								},
							},
						},
					},
				},
			},
			{
				name: "Read first IPv6 announcement",
				expected: &pb.RIBUpdate{
					Advertisement: true,
					Peer: &apinet.IP{
						Higher:  2306124484190404608,
						Lower:   2,
						Version: apinet.IP_IPv6,
					},
					Route: &apiroute.Route{
						Pfx: &apinet.Prefix{
							Address: &apinet.IP{
								Higher:  2306124484190404608,
								Lower:   0,
								Version: apinet.IP_IPv6,
							},
							Pfxlen: 32,
						},
						Paths: []*apiroute.Path{
							{
								Type: apiroute.Path_BGP,
								BGPPath: &apiroute.BGPPath{
									LocalPref: 4242,
									NextHop: &apinet.IP{
										Version: apinet.IP_IPv6,
										Higher:  2306124484190404608,
										Lower:   1,
									},
									Source: &apinet.IP{
										Version: apinet.IP_IPv6,
										Higher:  2306124484190404608,
										Lower:   2,
									},
								},
							},
						},
					},
				},
				nextPhase: true,
			},
			{
				name: "Read first IPv4 withdrawal",
				expected: &pb.RIBUpdate{
					Advertisement: false,
					Peer: &apinet.IP{
						Lower:   167772162,
						Version: apinet.IP_IPv4,
					},
					Route: &apiroute.Route{
						Pfx: &apinet.Prefix{
							Address: &apinet.IP{
								Lower:   2851995648,
								Version: apinet.IP_IPv4,
							},
							Pfxlen: 24,
						},
						Paths: []*apiroute.Path{
							{
								Type: apiroute.Path_BGP,
								BGPPath: &apiroute.BGPPath{
									LocalPref: 1337,
									NextHop: &apinet.IP{
										Version: apinet.IP_IPv4,
										Lower:   167772161,
									},
									Source: &apinet.IP{
										Version: apinet.IP_IPv4,
										Lower:   167772162,
									},
								},
							},
						},
					},
				},
			},
		}

		for _, test := range tests {
			update, err := streamClient.Recv()
			if err != nil {
				t.Fatalf("Recv failed: %v", err)
			}

			assert.Equal(t, test.expected, update)

			if test.nextPhase {
				nextPhase <- struct{}{}
			}
		}

		wg.Done()
	}()

	wg.Wait()
}
