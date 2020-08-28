package prom

import (
	"github.com/bio-routing/bio-rd/cmd/ris-mirror/rismirror"
	"github.com/bio-routing/bio-rd/cmd/ris-mirror/rismirror/metrics"
	"github.com/prometheus/client_golang/prometheus"

	vrf_prom "github.com/bio-routing/bio-rd/metrics/vrf/adapter/prom"
)

const (
	prefix = "bio_rismirror_"
)

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
}
