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
		OriginatorId: 8888,
		ClusterList:  []uint32{999, 199},
	}

	expected := &BGPPath{
		PathIdentifier: 100,
		NextHop:        bnet.IPv4FromOctets(10, 0, 0, 1),
		LocalPref:      1000,
		ASPath: types.ASPath{
			{
				Type: types.ASSequence,
				ASNs: []uint32{
					3320,
					201701,
				},
			},
		},
		Origin:        1,
		EBGP:          true,
		BGPIdentifier: 123,
		Source:        bnet.IPv4FromOctets(10, 0, 0, 2),
		Communities:   []uint32{100, 200, 300},
		LargeCommunities: []types.LargeCommunity{
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
		OriginatorID: 8888,
		ClusterList:  []uint32{999, 199},
	}

	result := BGPPathFromProtoBGPPath(input)
	assert.Equal(t, expected, result)
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
				LocalPref: 200,
			},
			q: &BGPPath{
				LocalPref: 100,
			},
			expected: 1,
		},
		{
			name: "Lpref #2",
			p: &BGPPath{
				LocalPref: 100,
			},
			q: &BGPPath{
				LocalPref: 200,
			},
			expected: -1,
		},
		{
			name: "AS Path Len",
			p: &BGPPath{
				ASPathLen: 100,
			},
			q: &BGPPath{
				ASPathLen: 200,
			},
			expected: 1,
		},
		{
			name: "AS Path Len #2",
			p: &BGPPath{
				ASPathLen: 200,
			},
			q: &BGPPath{
				ASPathLen: 100,
			},
			expected: -1,
		},
		{
			name: "Origin",
			p: &BGPPath{
				Origin: 1,
			},
			q: &BGPPath{
				Origin: 2,
			},
			expected: 1,
		},
		{
			name: "Origin #2",
			p: &BGPPath{
				Origin: 2,
			},
			q: &BGPPath{
				Origin: 1,
			},
			expected: -1,
		},
		{
			name: "MED",
			p: &BGPPath{
				MED: 1,
			},
			q: &BGPPath{
				MED: 2,
			},
			expected: 1,
		},
		{
			name: "MED #2",
			p: &BGPPath{
				MED: 2,
			},
			q: &BGPPath{
				MED: 1,
			},
			expected: -1,
		},
		{
			name: "EBGP",
			p: &BGPPath{
				EBGP: true,
			},
			q: &BGPPath{
				EBGP: false,
			},
			expected: 1,
		},
		{
			name: "EBGP #2",
			p: &BGPPath{
				EBGP: false,
			},
			q: &BGPPath{
				EBGP: true,
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

func TestBGPECMP(t *testing.T) {
	tests := []struct {
		name     string
		p        *BGPPath
		q        *BGPPath
		expected bool
	}{
		{
			name:     "Equal",
			p:        &BGPPath{},
			q:        &BGPPath{},
			expected: true,
		},
		{
			name:     "Lpref",
			p:        &BGPPath{LocalPref: 200},
			q:        &BGPPath{},
			expected: false,
		},
		{
			name:     "MED",
			p:        &BGPPath{MED: 200},
			q:        &BGPPath{},
			expected: false,
		},
		{
			name: "ASPath Len",
			p: &BGPPath{
				ASPathLen: 2,
			},
			q: &BGPPath{
				ASPathLen: 1,
			},
			expected: false,
		},
		{
			name:     "Origin",
			p:        &BGPPath{Origin: 1},
			q:        &BGPPath{},
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
		{
			name: "Cluster list, unknown attr and originator",
			path: &BGPPath{
				ASPath: []types.ASPathSegment{
					{
						Type: types.ASSequence,
						ASNs: []uint32{15169, 199714},
					},
				},
				ClusterList: []uint32{10, 20, 30},
				UnknownAttributes: []types.UnknownPathAttribute{
					{
						TypeCode: 100,
						Value:    []byte{1, 2, 3},
					},
				},
				OriginatorID: 10,
			},
			expected: 54,
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
				EBGP:         true,
				OriginatorID: 23,
				ClusterList:  []uint32{10, 20},
			},
			expectedString: "Local Pref: 0, Origin: Incomplete, AS Path: , BGP type: external, NEXT HOP: 0:0:0:0:0:0:0:0, MED: 0, Path ID: 0, Source: 0:0:0:0:0:0:0:0, Communities: [], LargeCommunities: [], OriginatorID: 0.0.0.23, ClusterList 0.0.0.10 0.0.0.20",
			expectedPrint: `		Local Pref: 0
		Origin: Incomplete
		AS Path: 
		BGP type: external
		NEXT HOP: 0:0:0:0:0:0:0:0
		MED: 0
		Path ID: 0
		Source: 0:0:0:0:0:0:0:0
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
