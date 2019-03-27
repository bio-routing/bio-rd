package vrf

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable/vrf/metrics"
)

func TestMetrics(t *testing.T) {
	green, err := New("green")
	if err != nil {
		t.Fatal(err)
	}
	green.IPv4UnicastRIB().AddPath(bnet.NewPfx(bnet.IPv4FromOctets(8, 0, 0, 0), 8), &route.Path{})
	green.IPv4UnicastRIB().AddPath(bnet.NewPfx(bnet.IPv4FromOctets(8, 0, 0, 0), 16), &route.Path{})
	green.IPv6UnicastRIB().AddPath(bnet.NewPfx(bnet.IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0, 0, 0, 0, 0), 48), &route.Path{})

	red, err := New("red")
	if err != nil {
		t.Fatal(err)
	}
	red.IPv6UnicastRIB().AddPath(bnet.NewPfx(bnet.IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0x100, 0, 0, 0, 0), 64), &route.Path{})
	red.IPv6UnicastRIB().AddPath(bnet.NewPfx(bnet.IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0x200, 0, 0, 0, 0), 64), &route.Path{})

	expected := []*metrics.VRFMetrics{
		{
			Name: "green",
			RIBs: []*metrics.RIBMetrics{
				{
					Name:       "inet.0",
					AFI:        afiIPv4,
					SAFI:       safiUnicast,
					RouteCount: 2,
				},
				{
					Name:       "inet6.0",
					AFI:        afiIPv6,
					SAFI:       safiUnicast,
					RouteCount: 1,
				},
			},
		},
		{
			Name: "red",
			RIBs: []*metrics.RIBMetrics{
				{
					Name:       "inet.0",
					AFI:        afiIPv4,
					SAFI:       safiUnicast,
					RouteCount: 0,
				},
				{
					Name:       "inet6.0",
					AFI:        afiIPv6,
					SAFI:       safiUnicast,
					RouteCount: 2,
				},
			},
		},
	}

	actual := Metrics()
	sortResult(actual)

	assert.Equal(t, expected, actual)
}

func sortResult(m []*metrics.VRFMetrics) {
	sort.Slice(m, func(i, j int) bool {
		return m[i].Name < m[j].Name
	})

	for _, v := range m {
		sort.Slice(v.RIBs, func(i, j int) bool {
			return m[i].Name < m[j].Name
		})
	}
}
