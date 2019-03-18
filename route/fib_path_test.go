package route

import (
	//"net"
	"testing"

	"github.com/bio-routing/bio-rd/net"
	"github.com/stretchr/testify/assert"
)

func TestNetlinkRouteDiff(t *testing.T) {
	tests := []struct {
		name     string
		left     []*FIBPath
		right    []*FIBPath
		expected []*FIBPath
	}{
		{
			name: "Equal",
			left: []*FIBPath{
				{
					Src:   net.IPv4(123),
					Table: 1,
				},
				{
					Src:   net.IPv4(456),
					Table: 2,
				},
			},
			right: []*FIBPath{
				{
					Src:   net.IPv4(123),
					Table: 1,
				},
				{
					Src:   net.IPv4(456),
					Table: 2,
				},
			},
			expected: []*FIBPath{},
		}, {
			name: "Left empty",
			left: make([]*FIBPath, 0),
			right: []*FIBPath{
				{
					Src:   net.IPv4(123),
					Table: 1,
				},
				{
					Src:   net.IPv4(456),
					Table: 2,
				},
			},
			expected: []*FIBPath{},
		}, {
			name: "Right empty",
			left: []*FIBPath{
				{
					Src:   net.IPv4(123),
					Table: 1,
				},
				{
					Src:   net.IPv4(456),
					Table: 2,
				},
			},
			right: make([]*FIBPath, 0),
			expected: []*FIBPath{
				{
					Src:   net.IPv4(123),
					Table: 1,
				},
				{
					Src:   net.IPv4(456),
					Table: 2,
				},
			},
		}, {
			name: "Diff",
			left: []*FIBPath{
				{
					Src:   net.IPv4(123),
					Table: 1,
				},
				{
					Src:   net.IPv4(456),
					Table: 2,
				},
			},
			right: []*FIBPath{
				{
					Src:   net.IPv4(123),
					Table: 1,
				},
			},
			expected: []*FIBPath{
				{
					Src:   net.IPv4(456),
					Table: 2,
				},
			},
		},
	}

	for _, test := range tests {
		res := FIBPathsDiff(test.left, test.right)
		assert.Equal(t, test.expected, res)
	}
}
