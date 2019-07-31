package route

import (
	"fmt"
	"testing"

	"github.com/bio-routing/bio-rd/net"
	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/stretchr/testify/assert"
)

func TestPathNextHop(t *testing.T) {
	tests := []struct {
		name     string
		p        *Path
		expected *bnet.IP
	}{
		{
			name: "BGP Path",
			p: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{
						NextHop: bnet.IPv4(123),
					},
				},
			},
			expected: bnet.IPv4(123),
		},
		{
			name: "Static Path",
			p: &Path{
				Type: StaticPathType,
				StaticPath: &StaticPath{
					NextHop: bnet.IPv4(456),
				},
			},
			expected: bnet.IPv4(456),
		},
		{
			name: "Netlink Path",
			p: &Path{
				Type: FIBPathType,
				FIBPath: &FIBPath{
					NextHop: bnet.IPv4(1000),
				},
			},
			expected: bnet.IPv4(1000),
		},
	}

	for _, test := range tests {
		res := test.p.NextHop()
		assert.Equal(t, test.expected, res, test.name)
	}
}

func TestPathCopy(t *testing.T) {
	tests := []struct {
		name     string
		p        *Path
		expected *Path
	}{
		{
			name: "nil test",
		},
	}

	for _, test := range tests {
		res := test.p.Copy()
		assert.Equal(t, test.expected, res, test.name)
	}
}

func TestEqual(t *testing.T) {
	tests := []struct {
		name     string
		p        *Path
		q        *Path
		expected bool
	}{
		{
			name:     "Different types",
			p:        &Path{Type: 100},
			q:        &Path{Type: 200},
			expected: false,
		},
	}

	for _, test := range tests {
		res := test.p.Equal(test.q)
		assert.Equalf(t, test.expected, res, test.name)
	}
}

func TestSelect(t *testing.T) {
	tests := []struct {
		name     string
		p        *Path
		q        *Path
		expected int8
	}{
		{
			name:     "All nil",
			expected: 0,
		},
		{
			name:     "p nil",
			q:        &Path{},
			expected: -1,
		},
		{
			name:     "q nil",
			p:        &Path{},
			expected: 1,
		},
		{
			name:     "p > q",
			p:        &Path{Type: 20},
			q:        &Path{Type: 10},
			expected: 1,
		},
		{
			name:     "p < q",
			p:        &Path{Type: 10},
			q:        &Path{Type: 20},
			expected: -1,
		},
		{
			name: "Static",
			p: &Path{
				Type: StaticPathType,
				StaticPath: &StaticPath{
					NextHop: net.IPv4(0),
				},
			},
			q: &Path{
				Type: StaticPathType,
				StaticPath: &StaticPath{
					NextHop: net.IPv4(0),
				},
			},
			expected: 0,
		},
		{
			name: "BGP",
			p: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: NewBGPPathA(),
				},
			},
			q: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: NewBGPPathA(),
				},
			},
			expected: 0,
		},
		{
			name: "Netlink",
			p: &Path{
				Type: FIBPathType,
				FIBPath: &FIBPath{
					NextHop: net.IPv4(0),
					Src:     net.IPv4(0),
				},
			},
			q: &Path{
				Type: FIBPathType,
				FIBPath: &FIBPath{
					NextHop: net.IPv4(0),
					Src:     net.IPv4(0),
				},
			},
			expected: 0,
		},
	}

	for _, test := range tests {
		res := test.p.Select(test.q)
		assert.Equalf(t, test.expected, res, "Test %q", test.name)
	}
}

func TestPathsDiff(t *testing.T) {
	tests := []struct {
		name     string
		any      []*Path
		a        []int
		b        []int
		expected []*Path
	}{
		{
			name: "Equal",
			any: []*Path{
				{
					Type: 10,
				},
				{
					Type: 20,
				},
			},
			a: []int{
				0, 1,
			},
			b: []int{
				0, 1,
			},
			expected: []*Path{},
		},
		{
			name: "Left empty",
			any: []*Path{
				{
					Type: 10,
				},
				{
					Type: 20,
				},
			},
			a: []int{},
			b: []int{
				0, 1,
			},
			expected: []*Path{},
		},
		{
			name: "Right empty",
			any: []*Path{
				{
					Type: 10,
				},
				{
					Type: 20,
				},
			},
			a: []int{0, 1},
			b: []int{},
			expected: []*Path{
				{
					Type: 10,
				},
				{
					Type: 20,
				},
			},
		},
		{
			name: "Disjunct",
			any: []*Path{
				{
					Type: 10,
				},
				{
					Type: 20,
				},
				{
					Type: 30,
				},
				{
					Type: 40,
				},
			},
			a: []int{0, 1},
			b: []int{2, 3},
			expected: []*Path{{
				Type: 10,
			},
				{
					Type: 20,
				}},
		},
	}

	for _, test := range tests {
		listA := make([]*Path, 0)
		for _, i := range test.a {
			listA = append(listA, test.any[i])
		}

		listB := make([]*Path, 0)
		for _, i := range test.b {
			listB = append(listB, test.any[i])
		}

		res := PathsDiff(listA, listB)
		assert.Equal(t, test.expected, res)
	}
}

func TestPathsContains(t *testing.T) {
	tests := []struct {
		name     string
		needle   int
		haystack []*Path
		expected bool
	}{
		{
			name:   "Existent",
			needle: 0,
			haystack: []*Path{
				{
					Type: 100,
				},
				{
					Type: 200,
				},
			},
			expected: true,
		},
		{
			name:   "Non existent",
			needle: -1,
			haystack: []*Path{
				{
					Type: 100,
				},
				{
					Type: 200,
				},
			},
			expected: false,
		},
	}

	for _, test := range tests {
		var needle *Path
		if test.needle >= 0 {
			needle = test.haystack[test.needle]
		} else {
			needle = &Path{}
		}

		res := pathsContains(needle, test.haystack)
		if res != test.expected {
			t.Errorf("Unexpected result for test %q: %v", test.name, res)
		}
	}
}

func TestNewNlPath(t *testing.T) {
	tests := []struct {
		name     string
		source   *Path
		expected *FIBPath
	}{
		{
			name: "BGPPath",
			source: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{
						NextHop: bnet.IPv4(123),
					},
				},
			},
			expected: &FIBPath{
				NextHop:  bnet.IPv4(123),
				Protocol: ProtoBio,
			},
		},
	}

	for _, test := range tests {
		var converted *FIBPath

		switch test.source.Type {
		case BGPPathType:
			converted = NewNlPathFromBgpPath(test.source.BGPPath)

		default:
			assert.Fail(t, fmt.Sprintf("Source-type %d is not supported", test.source.Type))
		}

		assert.Equalf(t, test.expected, converted, test.name)
	}
}

func TestECMP(t *testing.T) {
	tests := []struct {
		name  string
		left  *Path
		right *Path
		ecmp  bool
	}{
		{
			name: "BGP Path ecmp",
			left: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{
						LocalPref: 100,
						MED:       1,
						Origin:    123,
					},
					ASPathLen: 10,
				},
			},
			right: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{
						LocalPref: 100,
						MED:       1,
						Origin:    123,
					},
					ASPathLen: 10,
				},
			},
			ecmp: true,
		}, {
			name: "BGP Path not ecmp",
			left: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{
						LocalPref: 100,
						MED:       1,
						Origin:    123,
					},
					ASPathLen: 10,
				},
			},
			right: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{
						LocalPref: 100,
						MED:       1,
						Origin:    123,
					},
					ASPathLen: 5,
				},
			},
			ecmp: false,
		},
		{
			name: "Netlink Path ecmp",
			left: &Path{
				Type: FIBPathType,
				FIBPath: &FIBPath{
					Src:      bnet.IPv4(123),
					Priority: 1,
					Protocol: 1,
					Type:     1,
					Table:    1,
				},
			},
			right: &Path{
				Type: FIBPathType,
				FIBPath: &FIBPath{
					Src:      bnet.IPv4(123),
					Priority: 1,
					Protocol: 1,
					Type:     1,
					Table:    1,
				},
			},
			ecmp: true,
		}, {
			name: "Netlink Path not ecmp",
			left: &Path{
				Type: FIBPathType,
				FIBPath: &FIBPath{
					Src:      bnet.IPv4(123),
					Priority: 1,
					Protocol: 1,
					Type:     1,
					Table:    1,
				},
			},
			right: &Path{
				Type: FIBPathType,
				FIBPath: &FIBPath{
					Src:      bnet.IPv4(123),
					Priority: 2,
					Protocol: 1,
					Type:     1,
					Table:    1,
				},
			},
			ecmp: false,
		}, {
			name: "static Path ecmp",
			left: &Path{
				Type: StaticPathType,
				StaticPath: &StaticPath{
					NextHop: bnet.IPv4(123),
				},
			},
			right: &Path{
				Type: StaticPathType,
				StaticPath: &StaticPath{
					NextHop: bnet.IPv4(123),
				},
			},
			ecmp: true,
		}, {
			name: "static Path not ecmp",
			left: &Path{
				Type: StaticPathType,
				StaticPath: &StaticPath{
					NextHop: bnet.IPv4(123),
				},
			},
			right: &Path{
				Type: StaticPathType,
				StaticPath: &StaticPath{
					NextHop: bnet.IPv4(345),
				},
			},
			// ECMP is always true for staticPath
			ecmp: true,
		},
	}

	for _, test := range tests {
		ecmp := test.left.ECMP(test.right)
		assert.Equal(t, test.ecmp, ecmp, test.name)
	}
}

func TestFIBPathSelect(t *testing.T) {
	tests := []struct {
		name     string
		left     *FIBPath
		right    *FIBPath
		expected int8
	}{
		{
			name: "equal",
			left: &FIBPath{
				NextHop:  bnet.IPv4(123),
				Src:      bnet.IPv4(234),
				Priority: 1,
				Protocol: 1,
				Table:    1,
			},
			right: &FIBPath{
				NextHop:  bnet.IPv4(123),
				Src:      bnet.IPv4(234),
				Priority: 1,
				Protocol: 1,
				Table:    1,
			},
			expected: 0,
		},
		{
			name: "nextHop smaller",
			left: &FIBPath{
				NextHop: bnet.IPv4(1),
			},
			right: &FIBPath{
				NextHop: bnet.IPv4(2),
			},
			expected: -1,
		},
		{
			name: "nextHop bigger",
			left: &FIBPath{
				NextHop: bnet.IPv4(2),
			},
			right: &FIBPath{
				NextHop: bnet.IPv4(1),
			},
			expected: 1,
		},
		{
			name: "src smaller",
			left: &FIBPath{
				Src: bnet.IPv4(1),
			},
			right: &FIBPath{
				Src: bnet.IPv4(2),
			},
			expected: -1,
		},
		{
			name: "src bigger",
			left: &FIBPath{
				Src: bnet.IPv4(2),
			},
			right: &FIBPath{
				Src: bnet.IPv4(1),
			},
			expected: 1,
		},
		{
			name: "priority smaller",
			left: &FIBPath{
				Priority: 1,
			},
			right: &FIBPath{
				Priority: 2,
			},
			expected: -1,
		},
		{
			name: "priority bigger",
			left: &FIBPath{
				Priority: 2,
			},
			right: &FIBPath{
				Priority: 1,
			},
			expected: 1,
		},
		{
			name: "protocol smaller",
			left: &FIBPath{
				Protocol: 1,
			},
			right: &FIBPath{
				Protocol: 2,
			},
			expected: -1,
		},
		{
			name: "protocol bigger",
			left: &FIBPath{
				Protocol: 2,
			},
			right: &FIBPath{
				Protocol: 1,
			},
			expected: 1,
		},
		{
			name: "table smaller",
			left: &FIBPath{
				Table: 1,
			},
			right: &FIBPath{
				Table: 2,
			},
			expected: -1,
		},
		{
			name: "table bigger",
			left: &FIBPath{
				Table: 2,
			},
			right: &FIBPath{
				Table: 1,
			},
			expected: 1,
		},
	}

	for _, test := range tests {
		result := test.left.Select(test.right)
		assert.Equal(t, test.expected, result, test.name)
	}

}
