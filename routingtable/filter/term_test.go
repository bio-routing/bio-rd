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

func (m *mockAction) Equal(x actions.Action) bool {
	return false
}

func (*mockAction) Do(p *net.Prefix, pa *route.Path) actions.Result {
	pa.Type = route.StaticPathType

	return actions.Result{Path: pa}
}

func TestProcess(t *testing.T) {
	tests := []struct {
		name           string
		prefix         *net.Prefix
		path           *route.Path
		from           []*TermCondition
		then           []actions.Action
		expectReject   bool
		expectModified bool
	}{
		{
			name:   "empty from",
			prefix: net.NewPfx(net.IPv4FromOctets(100, 64, 0, 1), 8).Ptr(),
			path:   &route.Path{},
			from:   []*TermCondition{},
			then: []actions.Action{
				&actions.AcceptAction{},
			},
			expectReject:   false,
			expectModified: false,
		},
		{
			name:   "from matches",
			prefix: net.NewPfx(net.IPv4FromOctets(100, 64, 0, 1), 8).Ptr(),
			path:   &route.Path{},
			from: []*TermCondition{
				NewTermConditionWithPrefixLists(
					NewPrefixList(net.NewPfx(net.IPv4FromOctets(100, 64, 0, 1), 8).Ptr())),
			},
			then: []actions.Action{
				&actions.AcceptAction{},
			},
			expectReject:   false,
			expectModified: false,
		},
		{
			name:   "from does not match",
			prefix: net.NewPfx(net.IPv4FromOctets(100, 64, 0, 1), 8).Ptr(),
			path:   &route.Path{},
			from: []*TermCondition{
				NewTermConditionWithPrefixLists(
					NewPrefixList(net.NewPfx(net.IPv4(0), 32).Ptr())),
			},
			then: []actions.Action{
				&actions.AcceptAction{},
			},
			expectReject:   false,
			expectModified: false,
		},
		{
			name:   "modified",
			prefix: net.NewPfx(net.IPv4FromOctets(100, 64, 0, 1), 8).Ptr(),
			path:   &route.Path{},
			from: []*TermCondition{
				NewTermConditionWithPrefixLists(
					NewPrefixList(net.NewPfx(net.IPv4FromOctets(100, 64, 0, 1), 8).Ptr())),
			},
			then: []actions.Action{
				&mockAction{},
			},
			expectReject:   false,
			expectModified: true,
		},
		{
			name:   "modified and accepted (2 actions)",
			prefix: net.NewPfx(net.IPv4FromOctets(100, 64, 0, 1), 8).Ptr(),
			path:   &route.Path{},
			from: []*TermCondition{
				NewTermConditionWithRouteFilters(
					NewRouteFilter(net.NewPfx(net.IPv4FromOctets(100, 64, 0, 1), 8).Ptr(), NewExactMatcher())),
			},
			then: []actions.Action{
				&mockAction{},
				&actions.AcceptAction{},
			},
			expectReject:   false,
			expectModified: true,
		},
		{
			name:   "one of the prefix filters matches",
			prefix: net.NewPfx(net.IPv4FromOctets(100, 64, 0, 1), 8).Ptr(),
			path:   &route.Path{},
			from: []*TermCondition{
				{
					prefixLists: []*PrefixList{
						NewPrefixListWithMatcher(NewExactMatcher(), net.NewPfx(net.IPv4(0), 32).Ptr()),
						NewPrefixList(net.NewPfx(net.IPv4FromOctets(100, 64, 0, 1), 8).Ptr()),
					},
				},
			},
			then: []actions.Action{
				&actions.AcceptAction{},
			},
			expectReject:   false,
			expectModified: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			term := NewTerm("some name", test.from, test.then)

			res := term.Process(test.prefix, test.path.Copy())
			assert.Equal(t, test.expectReject, res.Reject, "reject")

			if !res.Path.Equal(test.path) && !test.expectModified {
				t.Fatal("expected path to be not modified but was")
			}

			if res.Path == test.path && test.expectModified {
				t.Fatal("expected path to be modified but was same reference")
			}
		})
	}
}
