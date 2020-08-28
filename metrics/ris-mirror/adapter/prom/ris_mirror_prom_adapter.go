package prom

import (
	"github.com/bio-routing/bio-rd/cmd/ris-mirror/rismirror"
	"github.com/bio-routing/bio-rd/protocols/ris/metrics"
	"github.com/prometheus/client_golang/prometheus"

	vrf_prom "github.com/bio-routing/bio-rd/metrics/vrf/adapter/prom"
)

const (
	prefix = "bio_rismirror_"
)

var (
	risMirrorSessionEstablishedDesc *prometheus.Desc
	risMirrorObserveRIBMessages     *prometheus.Desc
)

func init() {
	labels := []string{"sys_name", "agent_address"}

	risMirrorSessionEstablishedDesc = prometheus.NewDesc(prefix+"session_established", "Indicates if a RIS session is established", labels, nil)
	risMirrorObserveRIBMessages = prometheus.NewDesc(prefix+"observe_rib_messages", "Returns number of received rib monitoring messages", labels, nil)
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
	/*ch <- risMirrorSessionEstablishedDesc
	ch <- risMirrorObserveRIBMessages*/

	vrf_prom.DescribeRouter(ch)
}

// Collect conforms to the prometheus collector interface
func (c *risCollector) Collect(ch chan<- prometheus.Metric) {
	for _, rtr := range c.risMirror.Metrics().Routers {
		c.collectForRouter(ch, rtr)
	}
}

func (c *risCollector) collectForRouter(ch chan<- prometheus.Metric, rtr *metrics.RISMirrorRouterMetrics) {

	/*l := []string{rtr.SysName, rtr.Address.String()}

	ch <- prometheus.MustNewConstMetric(routeMonitoringMessagesDesc, prometheus.CounterValue, float64(rtr.RouteMonitoringMessages), l...)
	ch <- prometheus.MustNewConstMetric(statisticsReportMessages, prometheus.CounterValue, float64(rtr.StatisticsReportMessages), l...)
	ch <- prometheus.MustNewConstMetric(peerDownNotificationMessages, prometheus.CounterValue, float64(rtr.PeerDownNotificationMessages), l...)
	ch <- prometheus.MustNewConstMetric(peerUpNotificationMessages, prometheus.CounterValue, float64(rtr.PeerUpNotificationMessages), l...)
	ch <- prometheus.MustNewConstMetric(initiationMessages, prometheus.CounterValue, float64(rtr.InitiationMessages), l...)
	ch <- prometheus.MustNewConstMetric(terminationMessages, prometheus.CounterValue, float64(rtr.TerminationMessages), l...)
	ch <- prometheus.MustNewConstMetric(routeMirroringMessages, prometheus.CounterValue, float64(rtr.RouteMirroringMessages), l...)*/

	for _, vrfMetric := range rtr.VRFMetrics {
		vrf_prom.CollectForVRFRouter(ch, rtr.SysName, rtr.Address.String(), vrfMetric)
	}
}
