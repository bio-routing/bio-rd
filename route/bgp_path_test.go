package route

import (
	"testing"

	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/stretchr/testify/assert"
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
