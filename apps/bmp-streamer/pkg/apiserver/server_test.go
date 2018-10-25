package apiserver

import (
	"testing"

	pb "github.com/bio-routing/bio-rd/apps/bmp-streamer/pkg/bmpsrvapi"
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/route"
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
				prefix:        net.NewPfx(net.IPv4(200), 8),
				path: &route.Path{
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
				},
			},
			expected: &pb.RIBUpdate{
				Peer: &pb.IP{
					Lower:     220,
					IPVersion: 4,
				},
				Advertisement: true,
				Route: &pb.Route{
					Pfx: &pb.Prefix{
						Address: &pb.IP{
							Lower:     200,
							IPVersion: 4,
						},
						Pfxlen: 8,
					},
					Path: &pb.Path{
						Type: route.BGPPathType,
						BGPPath: &pb.BGPPath{
							PathIdentifier: 10,
							NextHop: &pb.IP{
								Lower:     210,
								IPVersion: 4,
							},
							LocalPref: 20,
							ASPath: []*pb.ASPathSegment{
								{
									ASSequence: true,
									ASNs:       []uint32{100, 200, 300},
								},
							},
							Origin:        1,
							MED:           1000,
							EBGP:          true,
							BGPIdentifier: 1337,
							Source: &pb.IP{
								Lower:     220,
								IPVersion: 4,
							},
							Communities: []uint32{10000, 20000},
							LargeCommunities: []*pb.LargeCommunity{
								{
									GlobalAdministrator: 1,
									DataPart1:           2,
									DataPart2:           3,
								},
							},
							ClusterList: []uint32{3, 4, 5},
							UnknownAttributes: []*pb.UnknownAttribute{
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
	}

	for _, test := range tests {
		res := test.u.toRIBUpdate()
		assert.Equal(t, test.expected, res, test.name)
	}
}
