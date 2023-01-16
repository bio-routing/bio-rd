package filter

import (
	"testing"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable/filter/actions"
	"github.com/stretchr/testify/assert"
)

func TestProcessTerms(t *testing.T) {
	tests := []struct {
		name           string
		prefix         *net.Prefix
		path           *route.Path
		expectedPath   *route.Path
		terms          []*Term
		expectAccept   bool
		expectModified bool
	}{
		{
			name:   "accept",
			prefix: net.NewPfx(net.IPv4(0), 0).Ptr(),
			path:   &route.Path{},
			terms: []*Term{
				{
					then: []actions.Action{
						&actions.AcceptAction{},
					},
				},
			},
			expectAccept:   true,
			expectModified: false,
		},
		{
			name:   "reject",
			prefix: net.NewPfx(net.IPv4(0), 0).Ptr(),
			path:   &route.Path{},
			terms: []*Term{
				{
					then: []actions.Action{
						&actions.RejectAction{},
					},
				},
			},
			expectAccept:   false,
			expectModified: false,
		},
		{
			name:   "accept before reject",
			prefix: net.NewPfx(net.IPv4(0), 0).Ptr(),
			path:   &route.Path{},
			terms: []*Term{
				{
					then: []actions.Action{
						&actions.AcceptAction{},
						&actions.RejectAction{},
					},
				},
			},
			expectAccept:   true,
			expectModified: false,
		},
		{
			name:   "modified",
			prefix: net.NewPfx(net.IPv4(0), 0).Ptr(),
			path:   &route.Path{},
			terms: []*Term{
				{
					then: []actions.Action{
						&mockAction{},
						&actions.AcceptAction{},
					},
				},
			},
			expectAccept:   true,
			expectModified: true,
		},
		{
			name:   "Overwrite Next-Hop",
			prefix: net.NewPfx(net.IPv4(0), 0).Ptr(),
			path: &route.Path{
				Type: route.StaticPathType,
				StaticPath: &route.StaticPath{
					NextHop: net.IPv4(0).Ptr(),
				},
			},
			terms: []*Term{
				{
					then: []actions.Action{
						actions.NewSetNextHopAction(net.IPv4(1).Ptr()),
						&actions.AcceptAction{},
					},
				},
			},
			expectedPath: &route.Path{
				Type: route.StaticPathType,
				StaticPath: &route.StaticPath{
					NextHop: net.IPv4(1).Ptr(),
				},
			},
			expectAccept:   true,
			expectModified: true,
		},
		{
			name:   "Overwrite Next-Hop & Localpref",
			prefix: net.NewPfx(net.IPv4(0), 0).Ptr(),
			path: &route.Path{
				Type: route.BGPPathType,
				BGPPath: &route.BGPPath{
					BGPPathA: &route.BGPPathA{
						LocalPref: 23,
						NextHop:   net.IPv4(0).Ptr(),
					},
				},
			},
			terms: []*Term{
				{
					then: []actions.Action{
						actions.NewSetNextHopAction(net.IPv4(1).Ptr()),
					},
				},
				{
					then: []actions.Action{
						actions.NewSetLocalPrefAction(42),
						&actions.AcceptAction{},
					},
				},
			},
			expectAccept:   true,
			expectModified: true,
			expectedPath: &route.Path{
				Type: route.BGPPathType,
				BGPPath: &route.BGPPath{
					BGPPathA: &route.BGPPathA{
						LocalPref: 42,
						NextHop:   net.IPv4(1).Ptr(),
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := NewFilter("some Name", test.terms)
			res := f.Process(test.prefix, test.path.Copy())
			p := res.Path
			reject := res.Reject

			assert.Equal(t, test.expectAccept, !reject)

			if test.expectModified {
				assert.NotEqual(t, test.path, p, test.name)
			}

			if test.expectedPath != nil {
				assert.Equal(t, test.expectedPath, p)
			}
		})
	}
}
