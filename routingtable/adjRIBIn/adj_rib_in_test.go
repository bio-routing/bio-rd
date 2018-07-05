package adjRIBIn

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"
)

func TestAddPath(t *testing.T) {
	routerID := net.IPv4FromOctets(1, 1, 1, 1).ToUint32()
	clusterID := net.IPv4FromOctets(2, 2, 2, 2).ToUint32()

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
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 100,
					},
				}),
			},
			removePfx:  net.Prefix{},
			removePath: nil,
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
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
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 100,
					},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 200,
					},
				}),
			},
			removePfx: net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8),
			removePath: &route.Path{
				Type: route.BGPPathType,
				BGPPath: &route.BGPPath{
					LocalPref: 100,
				},
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref: 200,
					},
				}),
			},
		},
		{
			name: "Add route with our RouterID as OriginatorID",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 32), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref:    111,
						OriginatorID: routerID,
					},
				}),
			},
			expected: []*route.Route{},
		},
		{
			name: "Add route with our ClusterID within ClusterList",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 32), &route.Path{
					Type: route.BGPPathType,
					BGPPath: &route.BGPPath{
						LocalPref:    222,
						OriginatorID: 23,
						ClusterList: []uint32{
							clusterID,
						},
					},
				}),
			},
			expected: []*route.Route{},
		},
	}

	for _, test := range tests {
		adjRIBIn := New(filter.NewAcceptAllFilter(), routingtable.NewContributingASNs(), routerID, clusterID)
		mc := routingtable.NewRTMockClient()
		adjRIBIn.ClientManager.Register(mc)

		for _, route := range test.routes {
			adjRIBIn.AddPath(route.Prefix(), route.Paths()[0])
		}

		if test.removePath != nil {
			removePathParams := mc.GetRemovePathParams()
			if removePathParams.Pfx != test.removePfx {
				t.Errorf("Test %q failed: Call to RemovePath did not propagate prefix properly: Got: %s Want: %s", test.name, removePathParams.Pfx.String(), test.removePfx.String())
			}

			assert.Equal(t, test.removePath, removePathParams.Path)
		}
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
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8), &route.Path{
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 9), &route.Path{
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 128, 0, 0), 9), &route.Path{
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
				}),
			},
			removePfx: net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8),
			removePath: &route.Path{
				Type:    route.BGPPathType,
				BGPPath: &route.BGPPath{},
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 9), &route.Path{
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 128, 0, 0), 9), &route.Path{
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
				}),
			},
			wantPropagation: true,
		},
		{
			name: "Remove non existing route",
			routes: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 9), &route.Path{
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 128, 0, 0), 9), &route.Path{
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
				}),
			},
			removePfx: net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8),
			removePath: &route.Path{
				Type:    route.BGPPathType,
				BGPPath: &route.BGPPath{},
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 9), &route.Path{
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
				}),
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 128, 0, 0), 9), &route.Path{
					Type:    route.BGPPathType,
					BGPPath: &route.BGPPath{},
				}),
			},
			wantPropagation: false,
		},
	}

	for _, test := range tests {
		adjRIBIn := New(filter.NewAcceptAllFilter(), routingtable.NewContributingASNs(), 1, 2)
		for _, route := range test.routes {
			adjRIBIn.AddPath(route.Prefix(), route.Paths()[0])
		}

		mc := routingtable.NewRTMockClient()
		adjRIBIn.ClientManager.Register(mc)
		adjRIBIn.RemovePath(test.removePfx, test.removePath)

		if test.wantPropagation {
			removePathParams := mc.GetRemovePathParams()
			if removePathParams.Pfx != test.removePfx {
				t.Errorf("Test %q failed: Call to RemovePath did not propagate prefix properly: Got: %s Want: %s", test.name, removePathParams.Pfx.String(), test.removePfx.String())
			}
			assert.Equal(t, test.removePath, removePathParams.Path)
		} else {
			removePathParams := mc.GetRemovePathParams()
			uninitialized := net.Prefix{}
			if removePathParams.Pfx != uninitialized {
				t.Errorf("Test %q failed: Call to RemovePath propagated unexpectedly", test.name)
			}
		}

		assert.Equal(t, test.expected, adjRIBIn.rt.Dump())
	}
}
