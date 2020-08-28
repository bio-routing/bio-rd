package prom

import (
	"strconv"

	"github.com/bio-routing/bio-rd/routingtable/vrf"
	"github.com/bio-routing/bio-rd/routingtable/vrf/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	prefix = "bio_vrf_"
)

var (
	routeCountDesc       *prometheus.Desc
	routeCountDescRouter *prometheus.Desc
)

func init() {
	labels := []string{"vrf_name", "vrf_rd", "rib", "afi", "safi"}
	routeCountDesc = prometheus.NewDesc(prefix+"route_count", "Number of routes in the RIB", labels, nil)
	routeCountDescRouter = prometheus.NewDesc(prefix+"route_count", "Number of routes in the RIB", append([]string{"sys_name", "agent_address"}, labels...), nil)
}

// NewCollector creates a new collector instance for the given BGP server
func NewCollector(r *vrf.VRFRegistry) prometheus.Collector {
	return &vrfCollector{
		registry: r,
	}
}

// vrfCollector provides a collector for VRF metrics of BIO to use with Prometheus
type vrfCollector struct {
	registry *vrf.VRFRegistry
}

// Describe conforms to the prometheus collector interface
func (c *vrfCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- routeCountDesc
}

// DescribeRouter conforms to the prometheus collector interface (used by BMP Server)
func DescribeRouter(ch chan<- *prometheus.Desc) {
	ch <- routeCountDescRouter
}

// Collect conforms to the prometheus collector interface
func (c *vrfCollector) Collect(ch chan<- prometheus.Metric) {
	for _, v := range vrf.Metrics(c.registry) {
		c.collectForVRF(ch, v)
	}
}

func (c *vrfCollector) collectForVRF(ch chan<- prometheus.Metric, v *metrics.VRFMetrics) {
	for _, rib := range v.RIBs {
		ch <- prometheus.MustNewConstMetric(routeCountDesc, prometheus.GaugeValue, float64(rib.RouteCount),
			v.Name, vrf.RouteDistinguisherHumanReadable(v.RD), rib.Name, strconv.Itoa(int(rib.AFI)), strconv.Itoa(int(rib.SAFI)))
	}
}

// CollectForVRFRouter collects metrics for a certain router (used by BMP Server)
func CollectForVRFRouter(ch chan<- prometheus.Metric, sysName string, agentAddress string, v *metrics.VRFMetrics) {
	for _, rib := range v.RIBs {
		ch <- prometheus.MustNewConstMetric(routeCountDescRouter, prometheus.GaugeValue, float64(rib.RouteCount),
			sysName, agentAddress, v.Name, vrf.RouteDistinguisherHumanReadable(v.RD), rib.Name, strconv.Itoa(int(rib.AFI)), strconv.Itoa(int(rib.SAFI)))
	}
}
