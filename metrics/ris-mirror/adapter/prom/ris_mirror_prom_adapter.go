package prom

import (
	"fmt"

	"github.com/bio-routing/bio-rd/cmd/ris-mirror/rismirror"
	"github.com/bio-routing/bio-rd/cmd/ris-mirror/rismirror/metrics"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
	"github.com/prometheus/client_golang/prometheus"

	vrf_prom "github.com/bio-routing/bio-rd/metrics/vrf/adapter/prom"
)

const (
	prefix = "bio_rismirror_"
)

var (
	mergedLocalRIBRouteCount             *prometheus.Desc
	mergedLocalRIBSingleSourceRouteCount *prometheus.Desc
)

func init() {
	labels := []string{"sys_name", "agent_address", "vrf", "afi", "rib"}
	mergedLocalRIBRouteCount = prometheus.NewDesc(prefix+"merged_locrib_route_count", "Number of unique routes", labels, nil)
	mergedLocalRIBSingleSourceRouteCount = prometheus.NewDesc(prefix+"merged_locrib_single_source_route_count", "Number of routes seen from single source", labels, nil)
}

// NewCollector creates a new collector instance for the given RIS mirror server
func NewCollector(risMirror *rismirror.RISMirror) prometheus.Collector {
	return &risCollector{
		risMirror: risMirror,
	}
}

// risCollector provides a collector for RIS metrics of BIO to use with Prometheus
type risCollector struct {
	risMirror *rismirror.RISMirror
}

// Describe conforms to the prometheus collector interface
func (c *risCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- mergedLocalRIBRouteCount
	ch <- mergedLocalRIBSingleSourceRouteCount

	vrf_prom.DescribeRouter(ch)
}

// Collect conforms to the prometheus collector interface
func (c *risCollector) Collect(ch chan<- prometheus.Metric) {
	for _, rtr := range c.risMirror.Metrics().Routers {
		c.collectForRouter(ch, rtr)
	}
}

func (c *risCollector) collectForRouter(ch chan<- prometheus.Metric, rtr *metrics.RISMirrorRouterMetrics) {
	for _, vrfMetric := range rtr.VRFMetrics {
		vrf_prom.CollectForVRFRouter(ch, rtr.SysName, rtr.Address.String(), vrfMetric)
	}

	for _, x := range rtr.InternalVRFMetrics {
		c.collectMergedLocRIBMetrics(ch, rtr, x)
	}
}

func (c *risCollector) collectMergedLocRIBMetrics(ch chan<- prometheus.Metric, rtr *metrics.RISMirrorRouterMetrics, v *metrics.InternalVRFMetrics) {
	ch <- prometheus.MustNewConstMetric(mergedLocalRIBRouteCount, prometheus.GaugeValue, float64(v.MergedLocRIBMetricsIPv4Unicast.UniqueRouteCount),
		getMergedLocRIBMetricsLabels(rtr, v, packet.IPv4AFI)...)

	ch <- prometheus.MustNewConstMetric(mergedLocalRIBRouteCount, prometheus.GaugeValue, float64(v.MergedLocRIBMetricsIPv6Unicast.UniqueRouteCount),
		getMergedLocRIBMetricsLabels(rtr, v, packet.IPv6AFI)...)

	ch <- prometheus.MustNewConstMetric(mergedLocalRIBSingleSourceRouteCount, prometheus.GaugeValue, float64(v.MergedLocRIBMetricsIPv4Unicast.RoutesWithSingleSourceCount),
		getMergedLocRIBMetricsLabels(rtr, v, packet.IPv4AFI)...)

	ch <- prometheus.MustNewConstMetric(mergedLocalRIBSingleSourceRouteCount, prometheus.GaugeValue, float64(v.MergedLocRIBMetricsIPv6Unicast.RoutesWithSingleSourceCount),
		getMergedLocRIBMetricsLabels(rtr, v, packet.IPv6AFI)...)
}

func getMergedLocRIBMetricsLabels(rtr *metrics.RISMirrorRouterMetrics, v *metrics.InternalVRFMetrics, afi uint8) []string {
	ret := []string{rtr.SysName, rtr.Address.String(), vrf.RouteDistinguisherHumanReadable(v.RD), fmt.Sprintf("%d", afi)}

	if afi == packet.IPv4AFI {
		return append(ret, v.MergedLocRIBMetricsIPv4Unicast.RIBName)
	}

	return append(ret, v.MergedLocRIBMetricsIPv6Unicast.RIBName)
}
