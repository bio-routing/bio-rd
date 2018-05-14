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

func (*mockAction) Do(p net.Prefix, pa *route.Path) (*route.Path, bool) {
	cp := *pa
	cp.Type = route.OSPFPathType

	return &cp, false
}

func TestProcess(t *testing.T) {
	tests := []struct {
		name           string
		prefix         net.Prefix
		path           *route.Path
		from           []*From
		then           []FilterAction
		expectReject   bool
		expectModified bool
	}{
		{
			name:   "empty from",
			prefix: net.NewPfx(strAddr("100.64.0.1"), 8),
			path:   &route.Path{},
			from:   []*From{},
			then: []FilterAction{
				&actions.AcceptAction{},
			},
			expectReject:   false,
			expectModified: false,
		},
		{
			name:   "from matches",
			prefix: net.NewPfx(strAddr("100.64.0.1"), 8),
			path:   &route.Path{},
			from: []*From{
				{
					&PrefixList{
						allowed: []net.Prefix{
							net.NewPfx(0, 0),
						},
					},
				},
			},
			then: []FilterAction{
				&actions.AcceptAction{},
			},
			expectReject:   false,
			expectModified: false,
		},
		{
			name:   "from does not match",
			prefix: net.NewPfx(strAddr("100.64.0.1"), 8),
			path:   &route.Path{},
			from: []*From{
				{
					&PrefixList{
						allowed: []net.Prefix{
							net.NewPfx(0, 32),
						},
					},
				},
			},
			then: []FilterAction{
				&actions.AcceptAction{},
			},
			expectReject:   true,
			expectModified: false,
		},
		{
			name:   "modified",
			prefix: net.NewPfx(strAddr("100.64.0.1"), 8),
			path:   &route.Path{},
			from: []*From{
				{
					&PrefixList{
						allowed: []net.Prefix{
							net.NewPfx(0, 0),
						},
					},
				},
			},
			then: []FilterAction{
				&mockAction{},
			},
			expectReject:   false,
			expectModified: true,
		},
		{
			name:   "modified and accepted (2 actions)",
			prefix: net.NewPfx(strAddr("100.64.0.1"), 8),
			path:   &route.Path{},
			from: []*From{
				{
					&PrefixList{
						allowed: []net.Prefix{
							net.NewPfx(0, 0),
						},
					},
				},
			},
			then: []FilterAction{
				&mockAction{},
				&actions.AcceptAction{},
			},
			expectReject:   false,
			expectModified: true,
		},
		{
			name:   "one of the prefix filters matches",
			prefix: net.NewPfx(strAddr("100.64.0.1"), 8),
			path:   &route.Path{},
			from: []*From{
				{
					&PrefixList{
						allowed: []net.Prefix{
							net.NewPfx(0, 32),
						},
					},
				},
				{
					&PrefixList{
						allowed: []net.Prefix{
							net.NewPfx(0, 0),
						},
					},
				},
			},
			then: []FilterAction{
				&actions.AcceptAction{},
			},
			expectReject:   false,
			expectModified: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(te *testing.T) {
			term := NewTerm(test.from, test.then)

			pa, reject := term.Process(test.prefix, test.path)
			assert.Equal(te, test.expectReject, reject, "reject")

			if pa != test.path && !test.expectModified {
				te.Fatal("expected path to be not modified but was")
			}

			if pa == test.path && test.expectModified {
				te.Fatal("expected path to be modified but was same reference")
			}
		})
	}
}

func strAddr(s string) uint32 {
	ret, _ := net.StrToAddr(s)
	return ret
}
