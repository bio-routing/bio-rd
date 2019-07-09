package vrf

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable/vrf/metrics"

	bnet "github.com/bio-routing/bio-rd/net"
)

func TestMetrics(t *testing.T) {
	r := NewVRFRegistry()
	green := r.CreateVRFIfNotExists("green", 0)
	green.IPv4UnicastRIB().AddPath(bnet.NewPfx(bnet.IPv4FromOctets(8, 0, 0, 0), 8), &route.Path{})
	green.IPv4UnicastRIB().AddPath(bnet.NewPfx(bnet.IPv4FromOctets(8, 0, 0, 0), 16), &route.Path{})
	green.IPv6UnicastRIB().AddPath(bnet.NewPfx(bnet.IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0, 0, 0, 0, 0), 48), &route.Path{})

	red := r.CreateVRFIfNotExists("red", 1)
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

	actual := Metrics(r)
	sortResult(actual)

	assert.Equal(t, expected, actual)
	_ = green
}

func sortResult(vrfMetrics []*metrics.VRFMetrics) {
	sort.Slice(vrfMetrics, func(i, j int) bool {
		return vrfMetrics[i].Name < vrfMetrics[j].Name
	})

	for _, vrfMetric := range vrfMetrics {
		sort.Slice(vrfMetric.RIBs, func(i, j int) bool {
			return vrfMetric.RIBs[i].Name < vrfMetric.RIBs[j].Name
		})
	}
}
