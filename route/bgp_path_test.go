package route

import (
	"testing"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/route/api"
	"github.com/stretchr/testify/assert"
)

func TestBGPPathFromProtoBGPPath(t *testing.T) {
	input := &api.BGPPath{
		PathIdentifier: 100,
		NextHop:        bnet.IPv4FromOctets(10, 0, 0, 1).ToProto(),
		LocalPref:      1000,
		AsPath: []*api.ASPathSegment{
			{
				AsSequence: true,
				Asns: []uint32{
					3320,
					201701,
				},
			},
		},
		Origin:        1,
		Ebgp:          true,
		BgpIdentifier: 123,
		Source:        bnet.IPv4FromOctets(10, 0, 0, 2).ToProto(),
		Communities:   []uint32{100, 200, 300},
		LargeCommunities: []*api.LargeCommunity{
			{
				GlobalAdministrator: 222,
				DataPart1:           500,
				DataPart2:           600,
			},
			{
				GlobalAdministrator: 333,
				DataPart1:           555,
				DataPart2:           666,
			},
		},
		UnknownAttributes: []*api.UnknownPathAttribute{
			{
				Optional:   true,
				Transitive: true,
				Partial:    true,
				TypeCode:   233,
				Value:      []byte{200, 222},
			},
		},
		OriginatorId:   8888,
		ClusterList:    []uint32{999, 199},
		OnlyToCustomer: 201701,
	}

	expected := &BGPPath{
		PathIdentifier: 100,
		BGPPathA: &BGPPathA{
			BGPIdentifier:  123,
			Source:         bnet.IPv4FromOctets(10, 0, 0, 2).Ptr(),
			NextHop:        bnet.IPv4FromOctets(10, 0, 0, 1).Ptr(),
			LocalPref:      1000,
			Origin:         1,
			EBGP:           true,
			OriginatorID:   8888,
			OnlyToCustomer: 201701,
		},
		ASPath: &types.ASPath{
			{
				Type: types.ASSequence,
				ASNs: []uint32{
					3320,
					201701,
				},
			},
		},

		Communities: &types.Communities{100, 200, 300},
		LargeCommunities: &types.LargeCommunities{
			{
				GlobalAdministrator: 222,
				DataPart1:           500,
				DataPart2:           600,
			},
			{
				GlobalAdministrator: 333,
				DataPart1:           555,
				DataPart2:           666,
			},
		},
		UnknownAttributes: []types.UnknownPathAttribute{
			{
				Optional:   true,
				Transitive: true,
				Partial:    true,
				TypeCode:   233,
				Value:      []byte{200, 222},
			},
		},
		ClusterList: &types.ClusterList{999, 199},
	}

	result := BGPPathFromProtoBGPPath(input, false)
	assert.Equal(t, expected, result)
}

func TestBGPPathToProto(t *testing.T) {
	tests := []struct {
		name     string
		value    *BGPPath
		expected *api.BGPPath
	}{
		{
			name:     "nil",
			value:    nil,
			expected: nil,
		},
		{
			name: "basic attrs only, like from withdraw()",
			value: &BGPPath{
				BMPPostPolicy:  false,
				PathIdentifier: 1,
			},
			expected: &api.BGPPath{
				PathIdentifier:    1,
				BmpPostPolicy:     false,
				UnknownAttributes: make([]*api.UnknownPathAttribute, 0),
			},
		},
		{
			name: "Path with empty BGPPathA",
			value: &BGPPath{
				BMPPostPolicy:  false,
				PathIdentifier: 1,
				BGPPathA:       &BGPPathA{},
			},
			expected: &api.BGPPath{
				PathIdentifier:    1,
				BmpPostPolicy:     false,
				UnknownAttributes: make([]*api.UnknownPathAttribute, 0),
			},
		},
		{
			name: "Path with BGPPathA + NextHop + Source",
			value: &BGPPath{
				BMPPostPolicy:  false,
				PathIdentifier: 1,
				BGPPathA: &BGPPathA{
					NextHop: bnet.IPv4FromOctets(10, 0, 0, 2).Ptr(),
					Source:  bnet.IPv4FromOctets(10, 0, 0, 2).Ptr(),
				},
			},
			expected: &api.BGPPath{
				PathIdentifier:    1,
				BmpPostPolicy:     false,
				UnknownAttributes: make([]*api.UnknownPathAttribute, 0),
				NextHop:           bnet.IPv4FromOctets(10, 0, 0, 2).ToProto(),
				Source:            bnet.IPv4FromOctets(10, 0, 0, 2).ToProto(),
			},
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.expected, test.value.ToProto(), test.name)
	}
}

func TestBGPSelect(t *testing.T) {
	tests := []struct {
		name     string
		p        *BGPPath
		q        *BGPPath
		expected int8
	}{
		{
			name: "Lpref",
			p: &BGPPath{
				BGPPathA: &BGPPathA{
					LocalPref: 200,
					Source:    bnet.IPv4(0).Ptr(),
					NextHop:   bnet.IPv4(0).Ptr(),
				},
			},
			q: &BGPPath{
				BGPPathA: &BGPPathA{
					LocalPref: 100,
					Source:    bnet.IPv4(0).Ptr(),
					NextHop:   bnet.IPv4(0).Ptr(),
				},
			},
			expected: 1,
		},
		{
			name: "Lpref #2",
			p: &BGPPath{
				BGPPathA: &BGPPathA{
					LocalPref: 100,
					Source:    bnet.IPv4(0).Ptr(),
					NextHop:   bnet.IPv4(0).Ptr(),
				},
			},
			q: &BGPPath{
				BGPPathA: &BGPPathA{
					LocalPref: 200,
					Source:    bnet.IPv4(0).Ptr(),
					NextHop:   bnet.IPv4(0).Ptr(),
				},
			},
			expected: -1,
		},
		{
			name: "AS Path Len",
			p: &BGPPath{
				ASPathLen: 100,
				BGPPathA: &BGPPathA{
					Source:  bnet.IPv4(0).Ptr(),
					NextHop: bnet.IPv4(0).Ptr(),
				},
			},
			q: &BGPPath{
				ASPathLen: 200,
				BGPPathA: &BGPPathA{
					Source:  bnet.IPv4(0).Ptr(),
					NextHop: bnet.IPv4(0).Ptr(),
				},
			},
			expected: 1,
		},
		{
			name: "AS Path Len #2",
			p: &BGPPath{
				ASPathLen: 200,
				BGPPathA: &BGPPathA{
					Source:  bnet.IPv4(0).Ptr(),
					NextHop: bnet.IPv4(0).Ptr(),
				},
			},
			q: &BGPPath{
				ASPathLen: 100,
				BGPPathA: &BGPPathA{
					Source:  bnet.IPv4(0).Ptr(),
					NextHop: bnet.IPv4(0).Ptr(),
				},
			},
			expected: -1,
		},
		{
			name: "Origin",
			p: &BGPPath{
				BGPPathA: &BGPPathA{
					Origin:  1,
					Source:  bnet.IPv4(0).Ptr(),
					NextHop: bnet.IPv4(0).Ptr(),
				},
			},
			q: &BGPPath{
				BGPPathA: &BGPPathA{
					Origin:  2,
					Source:  bnet.IPv4(0).Ptr(),
					NextHop: bnet.IPv4(0).Ptr(),
				},
			},
			expected: 1,
		},
		{
			name: "Origin #2",
			p: &BGPPath{
				BGPPathA: &BGPPathA{
					Origin:  2,
					Source:  bnet.IPv4(0).Ptr(),
					NextHop: bnet.IPv4(0).Ptr(),
				},
			},
			q: &BGPPath{
				BGPPathA: &BGPPathA{
					Origin:  1,
					Source:  bnet.IPv4(0).Ptr(),
					NextHop: bnet.IPv4(0).Ptr(),
				},
			},
			expected: -1,
		},
		{
			name: "MED",
			p: &BGPPath{
				BGPPathA: &BGPPathA{
					MED:     1,
					Source:  bnet.IPv4(0).Ptr(),
					NextHop: bnet.IPv4(0).Ptr(),
				},
			},
			q: &BGPPath{
				BGPPathA: &BGPPathA{
					MED:     2,
					Source:  bnet.IPv4(0).Ptr(),
					NextHop: bnet.IPv4(0).Ptr(),
				},
			},
			expected: 1,
		},
		{
			name: "MED #2",
			p: &BGPPath{
				BGPPathA: &BGPPathA{
					MED:     2,
					Source:  bnet.IPv4(0).Ptr(),
					NextHop: bnet.IPv4(0).Ptr(),
				},
			},
			q: &BGPPath{
				BGPPathA: &BGPPathA{
					MED:     1,
					Source:  bnet.IPv4(0).Ptr(),
					NextHop: bnet.IPv4(0).Ptr(),
				},
			},
			expected: -1,
		},
		{
			name: "EBGP",
			p: &BGPPath{
				BGPPathA: &BGPPathA{
					EBGP:    true,
					Source:  bnet.IPv4(0).Ptr(),
					NextHop: bnet.IPv4(0).Ptr(),
				},
			},
			q: &BGPPath{
				BGPPathA: &BGPPathA{
					EBGP:    false,
					Source:  bnet.IPv4(0).Ptr(),
					NextHop: bnet.IPv4(0).Ptr(),
				},
			},
			expected: 1,
		},
		{
			name: "EBGP #2",
			p: &BGPPath{
				BGPPathA: &BGPPathA{
					EBGP:    false,
					Source:  bnet.IPv4(0).Ptr(),
					NextHop: bnet.IPv4(0).Ptr(),
				},
			},
			q: &BGPPath{
				BGPPathA: &BGPPathA{
					EBGP:    true,
					Source:  bnet.IPv4(0).Ptr(),
					NextHop: bnet.IPv4(0).Ptr(),
				},
			},
			expected: -1,
		},
	}

	for _, test := range tests {
		res := test.p.Select(test.q)
		assert.Equal(t, test.expected, res, test.name)
	}
}

func TestCommunitiesString(t *testing.T) {
	tests := []struct {
		name     string
		comms    types.Communities
		expected string
	}{
		{
			name:     "two attributes",
			comms:    types.Communities{131080, 16778241},
			expected: "(2,8) (256,1025)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(te *testing.T) {
			p := &BGPPath{
				Communities: &test.comms,
			}

			assert.Equal(te, test.expected, p.CommunitiesString())
		})
	}
}

func TestLargeCommunitiesString(t *testing.T) {
	tests := []struct {
		name     string
		comms    types.LargeCommunities
		expected string
	}{
		{
			name: "two attributes",
			comms: types.LargeCommunities{
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
				LargeCommunities: &test.comms,
			}
			assert.Equal(te, test.expected, p.LargeCommunitiesString())
		})
	}
}

func TestBGPECMP(t *testing.T) {
	tests := []struct {
		name     string
		p        *BGPPath
		q        *BGPPath
		expected bool
	}{
		{
			name: "Equal",
			p: &BGPPath{
				BGPPathA: NewBGPPathA(),
			},
			q: &BGPPath{
				BGPPathA: NewBGPPathA(),
			},
			expected: true,
		},
		{
			name: "Lpref",
			p: &BGPPath{
				BGPPathA: &BGPPathA{
					LocalPref: 200,
				},
			},
			q: &BGPPath{
				BGPPathA: NewBGPPathA(),
			},
			expected: false,
		},
		{
			name: "MED",
			p: &BGPPath{
				BGPPathA: &BGPPathA{
					MED: 200,
				},
			},
			q: &BGPPath{
				BGPPathA: NewBGPPathA(),
			},
			expected: false,
		},
		{
			name: "ASPath Len",
			p: &BGPPath{
				BGPPathA:  NewBGPPathA(),
				ASPathLen: 2,
			},
			q: &BGPPath{
				BGPPathA:  NewBGPPathA(),
				ASPathLen: 1,
			},
			expected: false,
		},
		{
			name: "Origin",
			p: &BGPPath{
				BGPPathA: &BGPPathA{
					Origin: 1,
				},
			},
			q: &BGPPath{
				BGPPathA: NewBGPPathA(),
			},
			expected: false,
		},
	}

	for _, test := range tests {
		res := test.p.ECMP(test.q)
		assert.Equal(t, test.expected, res, test.name)
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
				BGPPathA: NewBGPPathA(),
				ASPath: &types.ASPath{
					{
						Type: types.ASSequence,
						ASNs: []uint32{15169, 199714},
					},
				},
				LargeCommunities: &types.LargeCommunities{},
				Communities:      &types.Communities{},
			},
			expected: 44,
		},
		{
			name: "communities",
			path: &BGPPath{
				BGPPathA: NewBGPPathA(),
				ASPath: &types.ASPath{
					{
						Type: types.ASSequence,
						ASNs: []uint32{15169, 199714},
					},
				},
				LargeCommunities: &types.LargeCommunities{},
				Communities:      &types.Communities{10, 20, 30},
			},
			expected: 59,
		},
		{
			name: "large communities",
			path: &BGPPath{
				BGPPathA: NewBGPPathA(),
				ASPath: &types.ASPath{
					{
						Type: types.ASSequence,
						ASNs: []uint32{15169, 199714},
					},
				},
				LargeCommunities: &types.LargeCommunities{
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
		{
			name: "Cluster list and originator",
			path: &BGPPath{
				ASPath: &types.ASPath{
					{
						Type: types.ASSequence,
						ASNs: []uint32{13335, 41981},
					},
				},
				ClusterList: &types.ClusterList{10, 20, 30},
				BGPPathA: &BGPPathA{
					OriginatorID: 10,
				},
			},
			expected: 44 + 3 + 3*4 + 4,
		},
		{
			name: "Cluster list + originator, OTC and unknown attr",
			path: &BGPPath{
				ASPath: &types.ASPath{
					{
						Type: types.ASSequence,
						ASNs: []uint32{15169, 199714},
					},
				},
				ClusterList: &types.ClusterList{10, 20, 30},
				UnknownAttributes: []types.UnknownPathAttribute{
					{
						TypeCode: 100,
						Value:    []byte{1, 2, 3},
					},
				},
				BGPPathA: &BGPPathA{
					OriginatorID:   10,
					Source:         bnet.IPv4(0).Ptr(),
					NextHop:        bnet.IPv4(0).Ptr(),
					OnlyToCustomer: 199714,
				},
			},
			expected: 44 + 19 + 6 + 4,
		},
	}

	for _, test := range tests {
		calcLen := test.path.Length()
		assert.Equal(t, test.expected, calcLen, test.name)
	}
}
func TestBGPPathString(t *testing.T) {
	tests := []struct {
		input          BGPPath
		expectedPrint  string
		expectedString string
	}{
		{
			input: BGPPath{
				BGPPathA: &BGPPathA{
					EBGP:           true,
					OriginatorID:   23,
					NextHop:        bnet.IPv6(0, 0).Ptr(),
					Source:         bnet.IPv6(0, 0).Ptr(),
					OnlyToCustomer: 2342,
				},
				ASPath:           &types.ASPath{},
				ClusterList:      &types.ClusterList{10, 20},
				Communities:      &types.Communities{},
				LargeCommunities: &types.LargeCommunities{},
			},
			expectedString: "Local Pref: 0, Origin: IGP, AS Path: , BGP type: external, NEXT HOP: ::, MED: 0, Path ID: 0, Source: ::, OnlyToCustomer: 2342, Communities: [], LargeCommunities: [], OriginatorID: 0.0.0.23, ClusterList 0.0.0.10 0.0.0.20",
			expectedPrint: `		Local Pref: 0
		Origin: IGP
		AS Path: 
		BGP type: external
		NEXT HOP: ::
		MED: 0
		Path ID: 0
		Source: ::
		OnlyToCustomer: 2342
		Communities: []
		LargeCommunities: []
		OriginatorID: 0.0.0.23
		ClusterList 0.0.0.10 0.0.0.20
`,
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expectedString, test.input.String())
		assert.Equal(t, test.expectedPrint, test.input.Print())
	}
}
