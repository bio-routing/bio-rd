package filter

import (
	"testing"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable/filter/actions"
	"github.com/stretchr/testify/assert"
)

type mockAction struct {
}

func (*mockAction) Do(p net.Prefix, pa *route.Path) actions.Result {
	cp := *pa
	cp.Type = route.OSPFPathType

	return actions.Result{Path: &cp}
}

func TestProcess(t *testing.T) {
	tests := []struct {
		name           string
		prefix         net.Prefix
		path           *route.Path
		from           []*TermCondition
		then           []Action
		expectReject   bool
		expectModified bool
	}{
		{
			name:   "empty from",
			prefix: net.NewPfx(net.IPv4FromOctets(100, 64, 0, 1), 8),
			path:   &route.Path{},
			from:   []*TermCondition{},
			then: []Action{
				&actions.AcceptAction{},
			},
			expectReject:   false,
			expectModified: false,
		},
		{
			name:   "from matches",
			prefix: net.NewPfx(net.IPv4FromOctets(100, 64, 0, 1), 8),
			path:   &route.Path{},
			from: []*TermCondition{
				NewTermConditionWithPrefixLists(
					NewPrefixList(net.NewPfx(net.IPv4FromOctets(100, 64, 0, 1), 8))),
			},
			then: []Action{
				&actions.AcceptAction{},
			},
			expectReject:   false,
			expectModified: false,
		},
		{
			name:   "from does not match",
			prefix: net.NewPfx(net.IPv4FromOctets(100, 64, 0, 1), 8),
			path:   &route.Path{},
			from: []*TermCondition{
				NewTermConditionWithPrefixLists(
					NewPrefixList(net.NewPfx(net.IPv4(0), 32))),
			},
			then: []Action{
				&actions.AcceptAction{},
			},
			expectReject:   false,
			expectModified: false,
		},
		{
			name:   "modified",
			prefix: net.NewPfx(net.IPv4FromOctets(100, 64, 0, 1), 8),
			path:   &route.Path{},
			from: []*TermCondition{
				NewTermConditionWithPrefixLists(
					NewPrefixList(net.NewPfx(net.IPv4FromOctets(100, 64, 0, 1), 8))),
			},
			then: []Action{
				&mockAction{},
			},
			expectReject:   false,
			expectModified: true,
		},
		{
			name:   "modified and accepted (2 actions)",
			prefix: net.NewPfx(net.IPv4FromOctets(100, 64, 0, 1), 8),
			path:   &route.Path{},
			from: []*TermCondition{
				NewTermConditionWithRouteFilters(
					NewRouteFilter(net.NewPfx(net.IPv4FromOctets(100, 64, 0, 1), 8), Exact())),
			},
			then: []Action{
				&mockAction{},
				&actions.AcceptAction{},
			},
			expectReject:   false,
			expectModified: true,
		},
		{
			name:   "one of the prefix filters matches",
			prefix: net.NewPfx(net.IPv4FromOctets(100, 64, 0, 1), 8),
			path:   &route.Path{},
			from: []*TermCondition{
				{
					prefixLists: []*PrefixList{
						NewPrefixListWithMatcher(Exact(), net.NewPfx(net.IPv4(0), 32)),
						NewPrefixList(net.NewPfx(net.IPv4FromOctets(100, 64, 0, 1), 8)),
					},
				},
			},
			then: []Action{
				&actions.AcceptAction{},
			},
			expectReject:   false,
			expectModified: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			term := NewTerm(test.from, test.then)

			res := term.Process(test.prefix, test.path)
			assert.Equal(t, test.expectReject, res.Reject, "reject")

			if res.Path != test.path && !test.expectModified {
				t.Fatal("expected path to be not modified but was")
			}

			if res.Path == test.path && test.expectModified {
				t.Fatal("expected path to be modified but was same reference")
			}
		})
	}
}
