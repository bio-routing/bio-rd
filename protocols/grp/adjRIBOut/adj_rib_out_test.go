package adjRIBOut

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/routingtable/filter"

	"github.com/bio-routing/bio-rd/route"
)

func TestAddRemovePath(t *testing.T) {
	testSteps := []struct {
		name         string
		routesAdd    []*route.Route
		routesRemove []*route.Route
		metaData     map[string]string
		expected     []*route.Route
	}{
		{
			name: "Add a simple route w/o meta data",
			routesAdd: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.GRPPathType,
					GRPPath: &route.GRPPath{
						NextHop: net.IPv4(0).Ptr(),
						Device:  "foo",
					},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.GRPPathType,
					GRPPath: &route.GRPPath{
						NextHop:  net.IPv4(0).Ptr(),
						Device:   "foo",
						MetaData: map[string]string{},
					},
				}),
			},
		},
		{
			name: "Add a simple route w/ meta data",
			routesAdd: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.GRPPathType,
					GRPPath: &route.GRPPath{
						NextHop: net.IPv4(0).Ptr(),
					},
				}),
			},
			metaData: map[string]string{
				"foo": "bar",
				"key": "value",
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.GRPPathType,
					GRPPath: &route.GRPPath{
						NextHop: net.IPv4(0).Ptr(),
						MetaData: map[string]string{
							"foo": "bar",
							"key": "value",
						},
					},
				}),
			},
		},
	}

	for _, test := range testSteps {
		adjRIBOut := New(nil, filter.NewAcceptAllFilterChain(), test.metaData)

		for _, route := range test.routesAdd {
			adjRIBOut.AddPath(route.Prefix().Ptr(), route.Paths()[0])
		}

		for _, route := range test.routesRemove {
			adjRIBOut.RemovePath(route.Prefix().Ptr(), route.Paths()[0])
		}

		assert.Equal(t, test.expected, adjRIBOut.rt.Dump())

		actualCount := adjRIBOut.RouteCount()
		expectedCount := int64(len(test.expected))
		if actualCount != expectedCount {
			t.Errorf("Expected route count %d differs from actual route count %d!\n", expectedCount, actualCount)
		}
	}
}

func TestRedistributeIntoGRP(t *testing.T) {
	testSteps := []struct {
		name         string
		routesAdd    []*route.Route
		routesRemove []*route.Route
		metaData     map[string]string
		expected     []*route.Route
	}{
		{
			name: "Redistribute from Static",
			routesAdd: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.StaticPathType,
					StaticPath: &route.StaticPath{
						NextHop: net.IPv4(0).Ptr(),
					},
				}),
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.GRPPathType,
					GRPPath: &route.GRPPath{
						NextHop:  net.IPv4(0).Ptr(),
						MetaData: map[string]string{},
					},
				}),
			},
		},
		{
			name: "Redistribute from Static w/ meta data",
			routesAdd: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.StaticPathType,
					StaticPath: &route.StaticPath{
						NextHop: net.IPv4(0).Ptr(),
					},
				}),
			},
			metaData: map[string]string{
				"foo": "bar",
				"key": "value",
			},
			expected: []*route.Route{
				route.NewRoute(net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 8).Ptr(), &route.Path{
					Type: route.GRPPathType,
					GRPPath: &route.GRPPath{
						NextHop: net.IPv4(0).Ptr(),
						MetaData: map[string]string{
							"foo": "bar",
							"key": "value",
						},
					},
				}),
			},
		},
	}

	for _, test := range testSteps {
		adjRIBOut := New(nil, filter.NewAcceptAllFilterChain(), test.metaData)

		for _, route := range test.routesAdd {
			adjRIBOut.AddPath(route.Prefix().Ptr(), route.Paths()[0])
		}

		for _, route := range test.routesRemove {
			adjRIBOut.RemovePath(route.Prefix().Ptr(), route.Paths()[0])
		}

		assert.Equal(t, test.expected, adjRIBOut.rt.Dump())

		actualCount := adjRIBOut.RouteCount()
		expectedCount := int64(len(test.expected))
		if actualCount != expectedCount {
			t.Errorf("Expected route count %d differs from actual route count %d!\n", expectedCount, actualCount)
		}
	}
}
