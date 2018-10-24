package route

import (
	"testing"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/stretchr/testify/assert"
)

func TestPathNextHop(t *testing.T) {
	tests := []struct {
		name     string
		p        *Path
		expected bnet.IP
	}{
		{
			name: "BGP Path",
			p: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					NextHop: bnet.IPv4(123),
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
				Type: NetlinkPathType,
				NetlinkPath: &NetlinkPath{
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
				Type:       StaticPathType,
				StaticPath: &StaticPath{},
			},
			q: &Path{
				Type:       StaticPathType,
				StaticPath: &StaticPath{},
			},
			expected: 0,
		},
		{
			name: "BGP",
			p: &Path{
				Type:    BGPPathType,
				BGPPath: &BGPPath{},
			},
			q: &Path{
				Type:    BGPPathType,
				BGPPath: &BGPPath{},
			},
			expected: 0,
		},
		{
			name: "Netlink",
			p: &Path{
				Type:        NetlinkPathType,
				NetlinkPath: &NetlinkPath{},
			},
			q: &Path{
				Type:        NetlinkPathType,
				NetlinkPath: &NetlinkPath{},
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
