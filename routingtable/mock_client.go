package routingtable

import (
	"fmt"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable/filter"
)

type RemovePathParams struct {
	Pfx  *net.Prefix
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

func (m *RTMockClient) EndOfRIB() {}

func (m *RTMockClient) AddPath(pfx *net.Prefix, p *route.Path) error {
	return nil
}

func (m *RTMockClient) AddPathInitialDump(pfx *net.Prefix, p *route.Path) error {
	return nil
}

func (m *RTMockClient) UpdateNewClient(client RouteTableClient) error {
	return fmt.Errorf("not implemented")
}

func (m *RTMockClient) Register(RouteTableClient) {}

func (m *RTMockClient) RegisterWithOptions(RouteTableClient, ClientOptions) {}

func (m *RTMockClient) Unregister(RouteTableClient) {}

// RemovePath removes the path for prefix `pfx`
func (m *RTMockClient) RemovePath(pfx *net.Prefix, p *route.Path) bool {
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

func (m *RTMockClient) RefreshRoute(*net.Prefix, []*route.Path) {}

func (m *RTMockClient) ReplaceFilterChain(filter.Chain) {}

func (m *RTMockClient) ReplacePath(*net.Prefix, *route.Path, *route.Path) {}

func (m *RTMockClient) Dispose() {}

func (m *RTMockClient) Flush() {}
