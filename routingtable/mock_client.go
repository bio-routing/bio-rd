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
	removePathParams RemovePathParams
}

func NewRTMockClient() *RTMockClient {
	return &RTMockClient{}
}

func (m *RTMockClient) GetRemovePathParams() RemovePathParams {
	return m.removePathParams
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
	m.removePathParams.Pfx = pfx
	m.removePathParams.Path = p
	return true
}

func (m *RTMockClient) RouteCount() int64 {
	return 0
}
