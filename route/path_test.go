package route

import (
	"fmt"
	"testing"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route/api"
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
						NextHop: bnet.IPv4(123).Ptr(),
					},
				},
			},
			expected: bnet.IPv4(123).Ptr(),
		},
		{
			name: "Static Path",
			p: &Path{
				Type: StaticPathType,
				StaticPath: &StaticPath{
					NextHop: bnet.IPv4(456).Ptr(),
				},
			},
			expected: bnet.IPv4(456).Ptr(),
		},
		{
			name: "Netlink Path",
			p: &Path{
				Type: FIBPathType,
				FIBPath: &FIBPath{
					NextHop: bnet.IPv4(1000).Ptr(),
				},
			},
			expected: bnet.IPv4(1000).Ptr(),
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
		{
			name:     "Empty path",
			p:        &Path{},
			expected: &Path{},
		},
		{
			name:     "Static path",
			p:        &Path{Type: StaticPathType},
			expected: &Path{Type: StaticPathType},
		},
		{
			name:     "BGP path",
			p:        &Path{Type: BGPPathType},
			expected: &Path{Type: BGPPathType},
		},
		{
			name:     "FIB path",
			p:        &Path{Type: FIBPathType},
			expected: &Path{Type: FIBPathType},
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
			name:     "Both nil",
			expected: false,
		},
		{
			name:     "p nil",
			q:        &Path{Type: StaticPathType},
			expected: false,
		},
		{
			name:     "q nil",
			p:        &Path{Type: StaticPathType},
			expected: false,
		},

		{
			name:     "Different types BGP/Static",
			p:        &Path{Type: BGPPathType},
			q:        &Path{Type: StaticPathType},
			expected: false,
		},
		{
			name:     "Different types OSPF/Static",
			p:        &Path{Type: OSPFPathType},
			q:        &Path{Type: StaticPathType},
			expected: false,
		},

		{
			name:     "Both Static Paths",
			p:        &Path{Type: StaticPathType, StaticPath: &StaticPath{NextHop: &bnet.IP{}}},
			q:        &Path{Type: StaticPathType, StaticPath: &StaticPath{NextHop: &bnet.IP{}}},
			expected: true,
		},
		{
			name: "Both BGP Paths",
			p: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{
						LocalPref: 100,
						MED:       1,
						Origin:    123,
						NextHop:   bnet.IPv4(0).Ptr(),
						Source:    bnet.IPv4(0).Ptr(),
					},
					ASPathLen: 10,
				},
			},
			q: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{
						LocalPref: 100,
						MED:       1,
						Origin:    123,
						NextHop:   bnet.IPv4(0).Ptr(),
						Source:    bnet.IPv4(0).Ptr(),
					},
					ASPathLen: 10,
				},
			},
			expected: true,
		},

		{
			name: "Both FIB paths",
			p: &Path{
				Type: FIBPathType,
				FIBPath: &FIBPath{
					NextHop: bnet.IPv4(0).Ptr(),
					Src:     bnet.IPv4(0).Ptr(),
				},
			},
			q: &Path{
				Type: FIBPathType,
				FIBPath: &FIBPath{
					NextHop: bnet.IPv4(0).Ptr(),
					Src:     bnet.IPv4(0).Ptr(),
				},
			},
			expected: true,
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
					NextHop: bnet.IPv4(0).Ptr(),
				},
			},
			q: &Path{
				Type: StaticPathType,
				StaticPath: &StaticPath{
					NextHop: bnet.IPv4(0).Ptr(),
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
					NextHop: bnet.IPv4(0).Ptr(),
					Src:     bnet.IPv4(0).Ptr(),
				},
			},
			q: &Path{
				Type: FIBPathType,
				FIBPath: &FIBPath{
					NextHop: bnet.IPv4(0).Ptr(),
					Src:     bnet.IPv4(0).Ptr(),
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
						NextHop: bnet.IPv4(123).Ptr(),
					},
				},
			},
			expected: &FIBPath{
				NextHop:  bnet.IPv4(123).Ptr(),
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
						NextHop:   bnet.IPv4(0).Ptr(),
						Source:    bnet.IPv4(0).Ptr(),
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
						NextHop:   bnet.IPv4(0).Ptr(),
						Source:    bnet.IPv4(0).Ptr(),
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
						NextHop:   bnet.IPv4(0).Ptr(),
						Source:    bnet.IPv4(0).Ptr(),
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
						NextHop:   bnet.IPv4(0).Ptr(),
						Source:    bnet.IPv4(0).Ptr(),
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
					Src:      bnet.IPv4(123).Ptr(),
					NextHop:  bnet.IPv4(123).Ptr(),
					Priority: 1,
					Protocol: 1,
					Type:     1,
					Table:    1,
				},
			},
			right: &Path{
				Type: FIBPathType,
				FIBPath: &FIBPath{
					Src:      bnet.IPv4(123).Ptr(),
					NextHop:  bnet.IPv4(123).Ptr(),
					Priority: 1,
					Protocol: 1,
					Type:     1,
					Table:    1,
				},
			},
			ecmp: true,
		},
		{
			name: "Netlink Path not ecmp",
			left: &Path{
				Type: FIBPathType,
				FIBPath: &FIBPath{
					Src:      bnet.IPv4(123).Ptr(),
					NextHop:  bnet.IPv4(123).Ptr(),
					Priority: 1,
					Protocol: 1,
					Type:     1,
					Table:    1,
				},
			},
			right: &Path{
				Type: FIBPathType,
				FIBPath: &FIBPath{
					Src:      bnet.IPv4(123).Ptr(),
					NextHop:  bnet.IPv4(123).Ptr(),
					Priority: 2,
					Protocol: 1,
					Type:     1,
					Table:    1,
				},
			},
			ecmp: false,
		},
		{
			name: "static Path ecmp",
			left: &Path{
				Type: StaticPathType,
				StaticPath: &StaticPath{
					NextHop: bnet.IPv4(123).Ptr(),
				},
			},
			right: &Path{
				Type: StaticPathType,
				StaticPath: &StaticPath{
					NextHop: bnet.IPv4(123).Ptr(),
				},
			},
			ecmp: true,
		}, {
			name: "static Path not ecmp",
			left: &Path{
				Type: StaticPathType,
				StaticPath: &StaticPath{
					NextHop: bnet.IPv4(123).Ptr(),
				},
			},
			right: &Path{
				Type: StaticPathType,
				StaticPath: &StaticPath{
					NextHop: bnet.IPv4(345).Ptr(),
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
				NextHop:  bnet.IPv4(123).Ptr(),
				Src:      bnet.IPv4(234).Ptr(),
				Priority: 1,
				Protocol: 1,
				Table:    1,
			},
			right: &FIBPath{
				NextHop:  bnet.IPv4(123).Ptr(),
				Src:      bnet.IPv4(234).Ptr(),
				Priority: 1,
				Protocol: 1,
				Table:    1,
			},
			expected: 0,
		},
		{
			name: "nextHop smaller",
			left: &FIBPath{
				NextHop: bnet.IPv4(1).Ptr(),
				Src:     bnet.IPv4(234).Ptr(),
			},
			right: &FIBPath{
				NextHop: bnet.IPv4(2).Ptr(),
				Src:     bnet.IPv4(234).Ptr(),
			},
			expected: -1,
		},
		{
			name: "nextHop bigger",
			left: &FIBPath{
				NextHop: bnet.IPv4(2).Ptr(),
				Src:     bnet.IPv4(234).Ptr(),
			},
			right: &FIBPath{
				NextHop: bnet.IPv4(1).Ptr(),
				Src:     bnet.IPv4(234).Ptr(),
			},
			expected: 1,
		},
		{
			name: "src smaller",
			left: &FIBPath{
				NextHop: bnet.IPv4(0).Ptr(),
				Src:     bnet.IPv4(1).Ptr(),
			},
			right: &FIBPath{
				NextHop: bnet.IPv4(0).Ptr(),
				Src:     bnet.IPv4(2).Ptr(),
			},
			expected: -1,
		},
		{
			name: "src bigger",
			left: &FIBPath{
				NextHop: bnet.IPv4(0).Ptr(),
				Src:     bnet.IPv4(2).Ptr(),
			},
			right: &FIBPath{
				NextHop: bnet.IPv4(0).Ptr(),
				Src:     bnet.IPv4(1).Ptr(),
			},
			expected: 1,
		},
		{
			name: "priority smaller",
			left: &FIBPath{
				NextHop:  bnet.IPv4(0).Ptr(),
				Src:      bnet.IPv4(234).Ptr(),
				Priority: 1,
			},
			right: &FIBPath{
				NextHop:  bnet.IPv4(0).Ptr(),
				Src:      bnet.IPv4(234).Ptr(),
				Priority: 2,
			},
			expected: -1,
		},
		{
			name: "priority bigger",
			left: &FIBPath{
				NextHop:  bnet.IPv4(0).Ptr(),
				Src:      bnet.IPv4(234).Ptr(),
				Priority: 2,
			},
			right: &FIBPath{
				NextHop:  bnet.IPv4(0).Ptr(),
				Src:      bnet.IPv4(234).Ptr(),
				Priority: 1,
			},
			expected: 1,
		},
		{
			name: "protocol smaller",
			left: &FIBPath{
				NextHop:  bnet.IPv4(0).Ptr(),
				Src:      bnet.IPv4(234).Ptr(),
				Protocol: 1,
			},
			right: &FIBPath{
				NextHop:  bnet.IPv4(0).Ptr(),
				Src:      bnet.IPv4(234).Ptr(),
				Protocol: 2,
			},
			expected: -1,
		},
		{
			name: "protocol bigger",
			left: &FIBPath{
				NextHop:  bnet.IPv4(0).Ptr(),
				Src:      bnet.IPv4(234).Ptr(),
				Protocol: 2,
			},
			right: &FIBPath{
				NextHop:  bnet.IPv4(0).Ptr(),
				Src:      bnet.IPv4(234).Ptr(),
				Protocol: 1,
			},
			expected: 1,
		},
		{
			name: "table smaller",
			left: &FIBPath{
				NextHop: bnet.IPv4(0).Ptr(),
				Src:     bnet.IPv4(234).Ptr(),
				Table:   1,
			},
			right: &FIBPath{
				NextHop: bnet.IPv4(0).Ptr(),
				Src:     bnet.IPv4(234).Ptr(),
				Table:   2,
			},
			expected: -1,
		},
		{
			name: "table bigger",
			left: &FIBPath{
				NextHop: bnet.IPv4(0).Ptr(),
				Src:     bnet.IPv4(234).Ptr(),
				Table:   2,
			},
			right: &FIBPath{
				NextHop: bnet.IPv4(0).Ptr(),
				Src:     bnet.IPv4(234).Ptr(),
				Table:   1,
			},
			expected: 1,
		},
	}

	for _, test := range tests {
		result := test.left.Select(test.right)
		assert.Equal(t, test.expected, result, test.name)
	}

}

func TestPathIsHidden(t *testing.T) {
	tests := []struct {
		name   string
		source *Path
		hidden bool
	}{
		{
			name: "BGPPath without hidden reason",
			source: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{
						NextHop: bnet.IPv4(123).Ptr(),
					},
				},
			},
			hidden: false,
		},
		{
			name: "BGPPath with hidden reason",
			source: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{
						NextHop: bnet.IPv4(123).Ptr(),
					},
				},
				HiddenReason: HiddenReasonOTCMismatch,
			},
			hidden: true,
		},
	}

	for _, test := range tests {
		assert.Equalf(t, test.hidden, test.source.IsHidden(), test.name)
	}
}

func TestPathHiddenReasonString(t *testing.T) {
	tests := []struct {
		name   string
		source *Path
		reason string
	}{
		{
			name: "Unknown",
			source: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{
						NextHop: bnet.IPv4(123).Ptr(),
					},
				},
				HiddenReason: 255,
			},
			reason: "unknown",
		},
		{
			name: "BGPPath without hidden reason",
			source: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{
						NextHop: bnet.IPv4(123).Ptr(),
					},
				},
			},
			reason: "",
		},
		{
			name: "Next-Hop unreachable",
			source: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{
						NextHop: bnet.IPv4(123).Ptr(),
					},
				},
				HiddenReason: HiddenReasonNextHopUnreachable,
			},
			reason: "Next-Hop unreachable",
		},
		{
			name: "Filtered by Policy",
			source: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{
						NextHop: bnet.IPv4(123).Ptr(),
					},
				},
				HiddenReason: HiddenReasonFilteredByPolicy,
			},
			reason: "Filtered by Policy",
		},
		{
			name: "HiddenReasonASLoop",
			source: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{
						NextHop: bnet.IPv4(123).Ptr(),
					},
				},
				HiddenReason: HiddenReasonASLoop,
			},
			reason: "AS Path loop",
		},
		{
			name: "Contains our Originator ID",
			source: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{
						NextHop: bnet.IPv4(123).Ptr(),
					},
				},
				HiddenReason: HiddenReasonOurOriginatorID,
			},
			reason: "Found our Router ID as Originator ID",
		},
		{
			name: "Cluster list loop",
			source: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{
						NextHop: bnet.IPv4(123).Ptr(),
					},
				},
				HiddenReason: HiddenReasonClusterLoop,
			},
			reason: "Found our cluster ID in cluster list",
		},
		{
			name: "OTC mismatch",
			source: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{
						NextHop: bnet.IPv4(123).Ptr(),
					},
				},
				HiddenReason: HiddenReasonOTCMismatch,
			},
			reason: "OTC mismatch",
		},
	}

	for _, test := range tests {
		assert.Equalf(t, test.reason, test.source.HiddenReasonString(), test.name)
	}
}

func TestString(t *testing.T) {
	staticPath := &Path{
		Type: StaticPathType,
		StaticPath: &StaticPath{
			NextHop: &bnet.IP{},
		},
	}

	bgpPath := &Path{
		Type: BGPPathType,
		BGPPath: &BGPPath{
			BGPPathA: &BGPPathA{},
		},
	}

	fibPath := &Path{
		Type: FIBPathType,
		FIBPath: &FIBPath{
			Src:     &bnet.IP{},
			NextHop: &bnet.IP{},
		},
	}

	tests := []struct {
		name   string
		path   *Path
		result string
	}{
		{
			name:   "unknown",
			path:   &Path{},
			result: "Unknown path type. Probably not implemented yet (0)",
		},
		{
			name:   "Static Path",
			path:   staticPath,
			result: staticPath.String(),
		},
		{
			name:   "BGP path",
			path:   bgpPath,
			result: bgpPath.BGPPath.String(),
		},
		{
			name:   "FIB path",
			path:   fibPath,
			result: fibPath.String(),
		},
	}

	for _, test := range tests {
		assert.Equalf(t, test.result, test.path.String(), test.name)
	}
}

func TestPrint(t *testing.T) {
	tests := []struct {
		name   string
		path   *Path
		result string
	}{
		{
			name:   "unknown",
			path:   &Path{},
			result: "\tProtocol: unknown\n\tHidden: no\n",
		},
		{
			name:   "Static Path",
			path:   &Path{Type: StaticPathType, StaticPath: &StaticPath{NextHop: &bnet.IP{}}},
			result: "\tProtocol: static\n\tHidden: no\n\t\tNext hop: ::\n",
		},
		{
			name: "Static Path (hidden)",
			path: &Path{
				Type:         StaticPathType,
				StaticPath:   &StaticPath{NextHop: &bnet.IP{}},
				HiddenReason: HiddenReasonFilteredByPolicy,
			},
			result: "\tProtocol: static\n\tHidden: yes (Filtered by Policy)\n\t\tNext hop: ::\n",
		},
	}

	for _, test := range tests {
		assert.Equalf(t, test.result, test.path.Print(), test.name)
	}
}

func TestToProto(t *testing.T) {
	ip := bnet.IPv4FromOctets(10, 0, 0, 0).Ptr()
	bgpPath := &Path{
		Type: BGPPathType,
		BGPPath: &BGPPath{
			BGPPathA: &BGPPathA{},
		},
	}

	tests := []struct {
		name   string
		path   *Path
		result *api.Path
	}{
		{
			name:   "Empty path",
			path:   &Path{},
			result: &api.Path{},
		},

		{
			name: "Static Path (empty)",
			path: &Path{
				Type:       StaticPathType,
				StaticPath: nil,
			},
			result: &api.Path{
				Type:       api.Path_Static,
				StaticPath: nil,
			},
		},
		{
			name: "Static Path with NH, LTime",
			path: &Path{
				Type: StaticPathType,
				StaticPath: &StaticPath{
					NextHop: ip,
				},
				LTime: 2342,
			},
			result: &api.Path{
				Type: api.Path_Static,
				StaticPath: &api.StaticPath{
					NextHop: ip.ToProto(),
				},
				TimeLearned: 2342,
			},
		},

		{
			name: "BGP path",
			path: bgpPath,
			result: &api.Path{
				Type:    api.Path_BGP,
				BgpPath: bgpPath.BGPPath.ToProto(),
			},
		},

		/*
		 * Hidden paths
		 */
		{
			name: "Not hidden",
			path: &Path{
				HiddenReason: HiddenReasonNone,
			},
			result: &api.Path{
				HiddenReason: api.Path_HiddenReasonNone,
			},
		},
		{
			name: "Hidden: NH unreach",
			path: &Path{
				HiddenReason: HiddenReasonNextHopUnreachable,
			},
			result: &api.Path{
				HiddenReason: api.Path_HiddenReasonNextHopUnreachable,
			},
		},
		{
			name: "Hidden: Filtered",
			path: &Path{
				HiddenReason: HiddenReasonFilteredByPolicy,
			},
			result: &api.Path{
				HiddenReason: api.Path_HiddenReasonFilteredByPolicy,
			},
		},
		{
			name: "Hidden: AS loop",
			path: &Path{
				HiddenReason: HiddenReasonASLoop,
			},
			result: &api.Path{
				HiddenReason: api.Path_HiddenReasonASLoop,
			},
		},
		{
			name: "Hidden: Our originator ID",
			path: &Path{
				HiddenReason: HiddenReasonOurOriginatorID,
			},
			result: &api.Path{
				HiddenReason: api.Path_HiddenReasonOurOriginatorID,
			},
		},
		{
			name: "Hidden: Cluster loop",
			path: &Path{
				HiddenReason: HiddenReasonClusterLoop,
			},
			result: &api.Path{
				HiddenReason: api.Path_HiddenReasonClusterLoop,
			},
		},
		{
			name: "Hidden: OTC mismatch",
			path: &Path{
				HiddenReason: HiddenReasonOTCMismatch,
			},
			result: &api.Path{
				HiddenReason: api.Path_HiddenReasonOTCMismatch,
			},
		},

		/*
			{
				name:   "Static Path",
				path:   &Path{Type: StaticPathType, StaticPath: &StaticPath{NextHop: &bnet.IP{}}},
				result: "\tProtocol: static\n\tHidden: no\n\t\tNext hop: ::\n",
			},
			{
				name: "Static Path (hidden)",
				path: &Path{
					Type:         StaticPathType,
					StaticPath:   &StaticPath{NextHop: &bnet.IP{}},
					HiddenReason: HiddenReasonFilteredByPolicy,
				},
				result: "\tProtocol: static\n\tHidden: yes (Filtered by Policy)\n\t\tNext hop: ::\n",
			},*/
	}

	for _, test := range tests {
		assert.Equalf(t, test.result, test.path.ToProto(), test.name)
	}
}

func TestGetNextHop(t *testing.T) {
	ip := bnet.IPv4FromOctets(10, 0, 0, 0).Ptr()

	tests := []struct {
		name   string
		path   *Path
		result *bnet.IP
	}{
		{
			name:   "Empty path",
			path:   &Path{},
			result: nil,
		},

		{
			name: "Static Path (empty)",
			path: &Path{
				Type:       StaticPathType,
				StaticPath: nil,
			},
			result: nil,
		},
		{
			name: "Static Path with NH",
			path: &Path{
				Type: StaticPathType,
				StaticPath: &StaticPath{
					NextHop: ip,
				},
			},
			result: ip,
		},

		{
			name: "BGP path (empty)",
			path: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{},
				},
			},
			result: nil,
		},
		{
			name: "BGP path with NH",
			path: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{
						NextHop: ip,
					},
				},
			},
			result: ip,
		},

		{
			name: "GRP Path (empty)",
			path: &Path{
				Type:       GRPPathType,
				StaticPath: nil,
			},
			result: nil,
		},
		{
			name: "GRP Path with NH",
			path: &Path{
				Type: GRPPathType,
				GRPPath: &GRPPath{
					NextHop: ip,
				},
			},
			result: ip,
		},
	}

	for _, test := range tests {
		assert.Equalf(t, test.result, test.path.GetNextHop(), test.name)
	}
}

func TestSetNextHop(t *testing.T) {
	ip := bnet.IPv4FromOctets(10, 0, 0, 0).Ptr()
	newNh := bnet.IPv4FromOctets(0, 0, 0, 1).Ptr()

	tests := []struct {
		name   string
		path   *Path
		result *Path
	}{
		{
			name:   "Empty path",
			path:   &Path{},
			result: &Path{},
		},

		{
			name: "Static Path (empty)",
			path: &Path{
				Type:       StaticPathType,
				StaticPath: nil,
			},
			result: &Path{
				Type:       StaticPathType,
				StaticPath: nil,
			},
		},
		{
			name: "Static Path with NH",

			path: &Path{
				Type: StaticPathType,
				StaticPath: &StaticPath{
					NextHop: ip,
				},
			},
			result: &Path{
				Type: StaticPathType,
				StaticPath: &StaticPath{
					NextHop: newNh,
				},
			},
		},

		{
			name: "BGP path (empty)",
			path: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{},
				},
			},
			result: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{
						NextHop: newNh,
					},
				},
			},
		},
		{
			name: "BGP path with NH",
			path: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{
						NextHop: ip,
					},
				},
			},
			result: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{
						NextHop: newNh,
					},
				},
			},
		},

		{
			name: "GRP path (empty)",
			path: &Path{
				Type:    GRPPathType,
				GRPPath: &GRPPath{},
			},
			result: &Path{
				Type: GRPPathType,
				GRPPath: &GRPPath{
					NextHop: newNh,
				},
			},
		},
		{
			name: "BGP path with NH",
			path: &Path{
				Type: GRPPathType,
				GRPPath: &GRPPath{
					NextHop: ip,
				},
			},
			result: &Path{
				Type: GRPPathType,
				GRPPath: &GRPPath{
					NextHop: newNh,
				},
			},
		},
	}

	for _, test := range tests {
		test.path.SetNextHop(newNh)

		assert.Equalf(t, test.result, test.path, test.name)
	}
}

func TestPurgePathInformation(t *testing.T) {
	ip := bnet.IPv4FromOctets(10, 0, 0, 0).Ptr()

	tests := []struct {
		name            string
		path            *Path
		pathTypeToPurge uint8
		result          *Path
	}{
		{
			name: "Static Path",
			path: &Path{
				Type: StaticPathType,
				StaticPath: &StaticPath{
					NextHop: ip,
				},
			},
			pathTypeToPurge: StaticPathType,
			result: &Path{
				Type: StaticPathType,
			},
		},
		{
			name: "BGP path with NH",
			path: &Path{
				Type: BGPPathType,
				BGPPath: &BGPPath{
					BGPPathA: &BGPPathA{
						NextHop: ip,
					},
				},
			},
			pathTypeToPurge: BGPPathType,
			result: &Path{
				Type: BGPPathType,
			},
		},
		{
			name: "GRP Path with NH",
			path: &Path{
				Type: GRPPathType,
				GRPPath: &GRPPath{
					NextHop: ip,
				},
			},
			pathTypeToPurge: GRPPathType,
			result: &Path{
				Type: GRPPathType,
			},
		},
	}

	for _, test := range tests {
		test.path.PurgePathInformation(test.pathTypeToPurge)
		assert.Equalf(t, test.result, test.path, test.name)
	}
}
