package rtmirror

import (
	"testing"

	"github.com/bio-routing/bio-rd/route"
	routeapi "github.com/bio-routing/bio-rd/route/api"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/stretchr/testify/assert"

	bnet "github.com/bio-routing/bio-rd/net"
)

type srcRouteTuple struct {
	src   interface{}
	route *routeapi.Route
}

func TestRTMirror(t *testing.T) {
	tests := []struct {
		name                string
		add                 []*srcRouteTuple
		expectedAfterAdd    []*route.Route
		remove              []*srcRouteTuple
		expectedAfterRemove []*route.Route
	}{
		{
			name: "Test #1: Single source",
			add: []*srcRouteTuple{
				{
					src: "a",
					route: &routeapi.Route{
						Pfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).ToProto(),
						Paths: []*routeapi.Path{
							{
								Type: routeapi.Path_Static,
								StaticPath: &routeapi.StaticPath{
									NextHop: bnet.IPv4FromOctets(1, 1, 1, 1).ToProto(),
								},
							},
						},
					},
				},
			},
			expectedAfterAdd: []*route.Route{
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.StaticPathType,
					StaticPath: &route.StaticPath{
						NextHop: bnet.IPv4FromOctets(1, 1, 1, 1).Ptr(),
					},
				}),
			},
			remove: []*srcRouteTuple{
				{
					src: "a",
					route: &routeapi.Route{
						Pfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).ToProto(),
						Paths: []*routeapi.Path{
							{
								Type: routeapi.Path_Static,
								StaticPath: &routeapi.StaticPath{
									NextHop: bnet.IPv4FromOctets(1, 1, 1, 1).ToProto(),
								},
							},
						},
					},
				},
			},
			expectedAfterRemove: []*route.Route{},
		},
		{
			name: "Test #2: Multiple source, single delete",
			add: []*srcRouteTuple{
				{
					src: "a",
					route: &routeapi.Route{
						Pfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).ToProto(),
						Paths: []*routeapi.Path{
							{
								Type: routeapi.Path_Static,
								StaticPath: &routeapi.StaticPath{
									NextHop: bnet.IPv4FromOctets(1, 1, 1, 1).ToProto(),
								},
							},
						},
					},
				},
				{
					src: "b",
					route: &routeapi.Route{
						Pfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).ToProto(),
						Paths: []*routeapi.Path{
							{
								Type: routeapi.Path_Static,
								StaticPath: &routeapi.StaticPath{
									NextHop: bnet.IPv4FromOctets(1, 1, 1, 1).ToProto(),
								},
							},
						},
					},
				},
			},
			expectedAfterAdd: []*route.Route{
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.StaticPathType,
					StaticPath: &route.StaticPath{
						NextHop: bnet.IPv4FromOctets(1, 1, 1, 1).Ptr(),
					},
				}),
			},
			remove: []*srcRouteTuple{
				{
					src: "a",
					route: &routeapi.Route{
						Pfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).ToProto(),
						Paths: []*routeapi.Path{
							{
								Type: routeapi.Path_Static,
								StaticPath: &routeapi.StaticPath{
									NextHop: bnet.IPv4FromOctets(1, 1, 1, 1).ToProto(),
								},
							},
						},
					},
				},
			},
			expectedAfterRemove: []*route.Route{
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.StaticPathType,
					StaticPath: &route.StaticPath{
						NextHop: bnet.IPv4FromOctets(1, 1, 1, 1).Ptr(),
					},
				}),
			},
		},
		{
			name: "Test #3: Multiple source, double delete",
			add: []*srcRouteTuple{
				{
					src: "a",
					route: &routeapi.Route{
						Pfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).ToProto(),
						Paths: []*routeapi.Path{
							{
								Type: routeapi.Path_Static,
								StaticPath: &routeapi.StaticPath{
									NextHop: bnet.IPv4FromOctets(1, 1, 1, 1).ToProto(),
								},
							},
						},
					},
				},
				{
					src: "b",
					route: &routeapi.Route{
						Pfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).ToProto(),
						Paths: []*routeapi.Path{
							{
								Type: routeapi.Path_Static,
								StaticPath: &routeapi.StaticPath{
									NextHop: bnet.IPv4FromOctets(1, 1, 1, 1).ToProto(),
								},
							},
						},
					},
				},
			},
			expectedAfterAdd: []*route.Route{
				route.NewRoute(bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.StaticPathType,
					StaticPath: &route.StaticPath{
						NextHop: bnet.IPv4FromOctets(1, 1, 1, 1).Ptr(),
					},
				}),
			},
			remove: []*srcRouteTuple{
				{
					src: "a",
					route: &routeapi.Route{
						Pfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).ToProto(),
						Paths: []*routeapi.Path{
							{
								Type: routeapi.Path_Static,
								StaticPath: &routeapi.StaticPath{
									NextHop: bnet.IPv4FromOctets(1, 1, 1, 1).ToProto(),
								},
							},
						},
					},
				},
				{
					src: "b",
					route: &routeapi.Route{
						Pfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).ToProto(),
						Paths: []*routeapi.Path{
							{
								Type: routeapi.Path_Static,
								StaticPath: &routeapi.StaticPath{
									NextHop: bnet.IPv4FromOctets(1, 1, 1, 1).ToProto(),
								},
							},
						},
					},
				},
			},
			expectedAfterRemove: []*route.Route{},
		},
	}

	for _, test := range tests {
		rt := routingtable.NewRoutingTable()
		rtm := New(rt)

		for _, a := range test.add {
			rtm.AddRoute(a.src, a.route)
		}

		assert.Equal(t, test.expectedAfterAdd, rt.Dump(), test.name)

		/*for _, r := range test.add {
			rtm.RemoveRoute(r.src, r.route)
		}

		assert.Equal(t, test.expectedAfterAdd, rt.Dump(), test.name)*/
	}
}
