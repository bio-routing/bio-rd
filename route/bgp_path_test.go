package route

import (
	"testing"

	"github.com/bio-routing/bio-rd/net"
	netapi "github.com/bio-routing/bio-rd/net/api"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/stretchr/testify/assert"

	pb "github.com/bio-routing/bio-rd/route/api"
)

func TestCommunitiesString(t *testing.T) {
	tests := []struct {
		name     string
		comms    []uint32
		expected string
	}{
		{
			name:     "two attributes",
			comms:    []uint32{131080, 16778241},
			expected: "(2,8) (256,1025)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(te *testing.T) {
			p := &BGPPath{
				Communities: test.comms,
			}

			assert.Equal(te, test.expected, p.CommunitiesString())
		})
	}
}

func TestLargeCommunitiesString(t *testing.T) {
	tests := []struct {
		name     string
		comms    []types.LargeCommunity
		expected string
	}{
		{
			name: "two attributes",
			comms: []types.LargeCommunity{
				{
					GlobalAdministrator: 1,
					DataPart1:           2,
					DataPart2:           3,
				},
				{
					GlobalAdministrator: 4,
					DataPart1:           5,
					DataPart2:           6,
				},
			},
			expected: "(1,2,3) (4,5,6)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(te *testing.T) {
			p := &BGPPath{
				LargeCommunities: test.comms,
			}
			assert.Equal(te, test.expected, p.LargeCommunitiesString())
		})
	}
}

func TestLength(t *testing.T) {
	tests := []struct {
		name     string
		path     *BGPPath
		expected uint16
	}{
		{
			name: "No communities",
			path: &BGPPath{
				ASPath: []types.ASPathSegment{
					{
						Type: types.ASSequence,
						ASNs: []uint32{15169, 199714},
					},
				},
				LargeCommunities: []types.LargeCommunity{},
				Communities:      []uint32{},
			},
			expected: 44,
		},
		{
			name: "communities",
			path: &BGPPath{
				ASPath: []types.ASPathSegment{
					{
						Type: types.ASSequence,
						ASNs: []uint32{15169, 199714},
					},
				},
				LargeCommunities: []types.LargeCommunity{},
				Communities:      []uint32{10, 20, 30},
			},
			expected: 59,
		},
		{
			name: "large communities",
			path: &BGPPath{
				ASPath: []types.ASPathSegment{
					{
						Type: types.ASSequence,
						ASNs: []uint32{15169, 199714},
					},
				},
				LargeCommunities: []types.LargeCommunity{
					{
						GlobalAdministrator: 199714,
						DataPart1:           100,
						DataPart2:           200,
					},
					{
						GlobalAdministrator: 199714,
						DataPart1:           100,
						DataPart2:           201,
					},
				},
			},
			expected: 71,
		},
	}

	for _, test := range tests {
		calcLen := test.path.Length()

		if calcLen != test.expected {
			t.Errorf("Unexpected result for test %q: Expected: %d Got: %d", test.name, test.expected, calcLen)
		}
	}
}

func TestBGPPathToProto(t *testing.T) {
	tests := []struct {
		name     string
		p        *BGPPath
		expected *pb.BGPPath
	}{
		{
			name: "Basics advert.",
			p: &BGPPath{
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
			expected: &pb.BGPPath{
				PathIdentifier: 10,
				NextHop: &netapi.IP{
					Lower:    210,
					Version: api.IP_IPv4,
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
				Source: &netapi.IP{
					Lower:    220,
					IsLegacy: true,
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
	}

	for _, test := range tests {
		res := test.p.ToProto()
		assert.Equal(t, test.expected, res, test.name)
	}
}
