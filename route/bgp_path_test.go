package route

import (
	"bytes"
	"testing"

	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
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
		comms    []packet.LargeCommunity
		expected string
	}{
		{
			name: "two attributes",
			comms: []packet.LargeCommunity{
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
		options  packet.Options
		wantFail bool
	}{
		{
			name: "Test 1",
			path: &BGPPath{},
		},
	}

	for _, test := range tests {
		calcLen := test.path.Length()
		pa, err := test.path.PathAttributes()
		if err != nil {
			if test.wantFail {
				continue
			}

			t.Errorf("Unexpected failure for test %q: %v", test.name, err)
			continue
		}

		if test.wantFail {
			t.Errorf("Unexpected success for test %q", test.name)
			continue
		}

		buf := bytes.Buffer(nil)
		pa.Serialize(buf, test.options)
		realLen := len(buf.Bytes())
		if realLen != calcLen {
			t.Errorf("Unexpected result for test %q: Expected: %d Got: %d", test.name, realLen, calcLen)
		}
	}
}
