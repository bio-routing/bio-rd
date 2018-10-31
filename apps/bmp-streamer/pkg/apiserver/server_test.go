package apiserver

import (
	"testing"

	pb "github.com/bio-routing/bio-rd/apps/bmp-streamer/pkg/bmpstreamer"
	"github.com/bio-routing/bio-rd/net"
	apinet "github.com/bio-routing/bio-rd/net/api"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/route"
	apiroute "github.com/bio-routing/bio-rd/route/api"
	"github.com/stretchr/testify/assert"
)

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
				route: route.NewRoute(net.NewPfx(net.IPv4(200), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						PathIdentifier: 10,
						NextHop:        net.IPv4(210),
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
						Source:        net.IPv4(220),
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
