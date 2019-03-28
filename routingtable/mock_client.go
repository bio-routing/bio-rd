package routingtable

import (
	"fmt"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
)

type RemovePathParams struct {
	Pfx  net.Prefix
	Path *route.Path
}

type RTMockClient struct {
	removed        []*RemovePathParams
	FakeRouteCount int64
}

func NewRTMockClient() *RTMockClient {
	return &RTMockClient{
		removed: make([]*RemovePathParams, 0),
	}
}

func (m *RTMockClient) ClientCount() uint64 {
	return 0
}

func (m *RTMockClient) Removed() []*RemovePathParams {
	return m.removed
}

// Dump is here to fulfill an interface
func (m *RTMockClient) Dump() []*route.Route {
	return nil
}

func (m *RTMockClient) AddPath(pfx net.Prefix, p *route.Path) error {
	return nil
}

func (m *RTMockClient) UpdateNewClient(client RouteTableClient) error {
	return fmt.Errorf("Not implemented")
}

func (m *RTMockClient) Register(RouteTableClient) {
	return
}

func (m *RTMockClient) RegisterWithOptions(RouteTableClient, ClientOptions) {
	return
}

func (m *RTMockClient) Unregister(RouteTableClient) {
	return
}

// RemovePath removes the path for prefix `pfx`
func (m *RTMockClient) RemovePath(pfx net.Prefix, p *route.Path) bool {
	params := &RemovePathParams{
		Path: p,
		Pfx:  pfx,
	}
	m.removed = append(m.removed, params)

	return true
}

func (m *RTMockClient) RouteCount() int64 {
	return m.FakeRouteCount
}
