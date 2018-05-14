package filter

import (
	"testing"

	"github.com/bio-routing/bio-rd/routingtable/filter/actions"
	"github.com/stretchr/testify/assert"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
)

type clientMock struct {
	routingtable.ClientManager
	addPathCalled    bool
	removePathCalled bool
	path             *route.Path
}

func (m *clientMock) AddPath(p net.Prefix, pa *route.Path) error {
	m.path = pa
	m.addPathCalled = true
	return nil
}

func (m *clientMock) RemovePath(p net.Prefix, pa *route.Path) bool {
	m.path = pa
	m.removePathCalled = true
	return false
}

func (m *clientMock) UpdateNewClient(c routingtable.RouteTableClient) error {
	return nil
}

func newClientMock() *clientMock {
	m := &clientMock{}
	m.ClientManager = routingtable.NewClientManager(m)
	return m
}

func TestAddPath(t *testing.T) {
	tests := []struct {
		name           string
		prefix         net.Prefix
		path           *route.Path
		term           *Term
		exptectCalled  bool
		expectModified bool
	}{
		{
			name:   "accept",
			prefix: net.NewPfx(0, 0),
			path:   &route.Path{},
			term: &Term{
				then: []FilterAction{
					&actions.AcceptAction{},
				},
			},
			exptectCalled:  true,
			expectModified: false,
		},
		{
			name:   "reject",
			prefix: net.NewPfx(0, 0),
			path:   &route.Path{},
			term: &Term{
				then: []FilterAction{
					&actions.RejectAction{},
				},
			},
			exptectCalled:  false,
			expectModified: false,
		},
		{
			name:   "modified",
			prefix: net.NewPfx(0, 0),
			path:   &route.Path{},
			term: &Term{
				then: []FilterAction{
					&mockAction{},
					&actions.AcceptAction{},
				},
			},
			exptectCalled:  true,
			expectModified: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(te *testing.T) {
			m := newClientMock()

			f := NewFilter([]*Term{test.term})
			f.Register(m)

			f.AddPath(test.prefix, test.path)
			assert.Equal(te, test.exptectCalled, m.addPathCalled, "called")

			if !test.exptectCalled {
				return
			}

			if m.path != test.path && !test.expectModified {
				te.Fatal("expected path to be not modified but was")
			}

			if m.path == test.path && test.expectModified {
				te.Fatal("expected path to be modified but was same reference")
			}
		})
	}
}

func TestRemovePath(t *testing.T) {
	tests := []struct {
		name           string
		prefix         net.Prefix
		path           *route.Path
		term           *Term
		exptectCalled  bool
		expectModified bool
	}{
		{
			name:   "accept",
			prefix: net.NewPfx(0, 0),
			path:   &route.Path{},
			term: &Term{
				then: []FilterAction{
					&actions.AcceptAction{},
				},
			},
			exptectCalled:  true,
			expectModified: false,
		},
		{
			name:   "reject",
			prefix: net.NewPfx(0, 0),
			path:   &route.Path{},
			term: &Term{
				then: []FilterAction{
					&actions.RejectAction{},
				},
			},
			exptectCalled:  false,
			expectModified: false,
		},
		{
			name:   "modified",
			prefix: net.NewPfx(0, 0),
			path:   &route.Path{},
			term: &Term{
				then: []FilterAction{
					&mockAction{},
					&actions.AcceptAction{},
				},
			},
			exptectCalled:  true,
			expectModified: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(te *testing.T) {
			m := newClientMock()

			f := NewFilter([]*Term{test.term})
			f.Register(m)

			f.RemovePath(test.prefix, test.path)
			assert.Equal(te, test.exptectCalled, m.removePathCalled, "called")

			if !test.exptectCalled {
				return
			}

			if m.path != test.path && !test.expectModified {
				te.Fatal("expected path to be not modified but was")
			}

			if m.path == test.path && test.expectModified {
				te.Fatal("expected path to be modified but was same reference")
			}
		})
	}
}
