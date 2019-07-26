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
		prefix         net.Prefix
		path           *route.Path
		term           *Term
		expectAccept   bool
		expectModified bool
	}{
		{
			name:   "accept",
			prefix: net.NewPfx(net.IPv4(0), 0),
			path:   &route.Path{},
			term: &Term{
				then: []actions.Action{
					&actions.AcceptAction{},
				},
			},
			expectAccept:   true,
			expectModified: false,
		},
		{
			name:   "reject",
			prefix: net.NewPfx(net.IPv4(0), 0),
			path:   &route.Path{},
			term: &Term{
				then: []actions.Action{
					&actions.RejectAction{},
				},
			},
			expectAccept:   false,
			expectModified: false,
		},
		{
			name:   "accept before reject",
			prefix: net.NewPfx(net.IPv4(0), 0),
			path:   &route.Path{},
			term: &Term{
				then: []actions.Action{
					&actions.AcceptAction{},
					&actions.RejectAction{},
				},
			},
			expectAccept:   true,
			expectModified: false,
		},
		{
			name:   "modified",
			prefix: net.NewPfx(net.IPv4(0), 0),
			path:   &route.Path{},
			term: &Term{
				then: []actions.Action{
					&mockAction{},
					&actions.AcceptAction{},
				},
			},
			expectAccept:   true,
			expectModified: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := NewFilter("some Name", []*Term{test.term})
			res := f.Process(test.prefix, test.path.Copy())
			p := res.Path
			reject := res.Reject

			assert.Equal(t, test.expectAccept, !reject)

			if test.expectModified {
				assert.NotEqual(t, test.path, p, test.name)
			}
		})
	}
}
