package adjRIBIn

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
)

type RTMockClient struct {
	removePathParams struct {
		pfx  net.Prefix
		path *route.Path
	}
}

func NewRTMockClient() *RTMockClient {
	return &RTMockClient{}
}

func (m *RTMockClient) AddPath(pfx net.Prefix, p *route.Path) error {
	return nil
}

func (m *RTMockClient) UpdateNewClient(client routingtable.RouteTableClient) error {
	return fmt.Errorf("Not implemented")
}

func (m *RTMockClient) Register(routingtable.RouteTableClient) {
	return
}

func (m *RTMockClient) RegisterWithOptions(routingtable.RouteTableClient, routingtable.ClientOptions) {
	return
}

func (m *RTMockClient) Unregister(routingtable.RouteTableClient) {
	return
}

// RemovePath removes the path for prefix `pfx`
func (m *RTMockClient) RemovePath(pfx net.Prefix, p *route.Path) bool {
	m.removePathParams.pfx = pfx
	m.removePathParams.path = p
	return true
}

func TestAddPath(t *testing.T) {
	tests := []struct {
		name       string
		routes     []*route.Route
		removePfx  net.Prefix
		removePath *route.Path
		expected   []*route.Route
	}{
		{
			name: "Add route",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 100,
					},
				}),
			},
			removePfx:  net.NewPfx(0, 0),
			removePath: nil,
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 100,
					},
				}),
			},
		},
		{
			name: "Overwrite routes",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 100,
					},
				}),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 200,
					},
				}),
			},
			removePfx: net.NewPfx(strAddr("10.0.0.0"), 8),
			removePath: &route.Path{
				Type: route.BGPPathType,
				BGPPath: &route.BGPPath{
					LocalPref: 100,
				},
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 200,
					},
				}),
			},
		},
	}

	for _, test := range tests {
		adjRIBIn := New()
		mc := NewRTMockClient()
		adjRIBIn.ClientManager.Register(mc)

		for _, route := range test.routes {
			adjRIBIn.AddPath(route.Prefix(), route.Paths()[0])
		}

		if mc.removePathParams.pfx != test.removePfx {
			t.Errorf("Test %q failed: Call to RemovePath did not propagate prefix properly: Got: %s Want: %s", test.name, mc.removePathParams.pfx.String(), test.removePfx.String())
		}

		assert.Equal(t, test.removePath, mc.removePathParams.path)
		assert.Equal(t, test.expected, adjRIBIn.rt.Dump())
	}
}

func TestRemovePath(t *testing.T) {
	tests := []struct {
		name            string
		routes          []*route.Route
		removePfx       net.Prefix
		removePath      *route.Path
		expected        []*route.Route
		wantPropagation bool
	}{
		{
			name: "Remove an existing route",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 8), &route.Path{
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
				}),
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 9), &route.Path{
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
				}),
				route.NewRoute(net.NewPfx(strAddr("10.128.0.0"), 9), &route.Path{
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
				}),
			},
			removePfx: net.NewPfx(strAddr("10.0.0.0"), 8),
			removePath: &route.Path{
				Type:    route.BGPPathType,
				BGPPath: &route.BGPPath{},
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 9), &route.Path{
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
				}),
				route.NewRoute(net.NewPfx(strAddr("10.128.0.0"), 9), &route.Path{
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
				}),
			},
			wantPropagation: true,
		},
		{
			name: "Remove non existing route",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 9), &route.Path{
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
				}),
				route.NewRoute(net.NewPfx(strAddr("10.128.0.0"), 9), &route.Path{
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
				}),
			},
			removePfx: net.NewPfx(strAddr("10.0.0.0"), 8),
			removePath: &route.Path{
				Type:    route.BGPPathType,
				BGPPath: &route.BGPPath{},
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(strAddr("10.0.0.0"), 9), &route.Path{
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
				}),
				route.NewRoute(net.NewPfx(strAddr("10.128.0.0"), 9), &route.Path{
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
				}),
			},
			wantPropagation: false,
		},
	}

	for _, test := range tests {
		adjRIBIn := New()
		for _, route := range test.routes {
			adjRIBIn.AddPath(route.Prefix(), route.Paths()[0])
		}

		mc := NewRTMockClient()
		adjRIBIn.ClientManager.Register(mc)
		adjRIBIn.RemovePath(test.removePfx, test.removePath)

		if test.wantPropagation {
			if mc.removePathParams.pfx != test.removePfx {
				t.Errorf("Test %q failed: Call to RemovePath did not propagate prefix properly: Got: %s Want: %s", test.name, mc.removePathParams.pfx.String(), test.removePfx.String())
			}
			assert.Equal(t, test.removePath, mc.removePathParams.path)
		} else {
			if mc.removePathParams.pfx != net.NewPfx(0, 0) {
				t.Errorf("Test %q failed: Call to RemovePath propagated unexpectedly", test.name)
			}
		}

		assert.Equal(t, test.expected, adjRIBIn.rt.Dump())
	}
}

func strAddr(s string) uint32 {
	ret, _ := net.StrToAddr(s)
	return ret
}
