package prom

import (
	"github.com/bio-routing/bio-rd/protocols/bgp/metrics"
	"github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/prometheus/client_golang/prometheus"

	bgp_prom "github.com/bio-routing/bio-rd/metrics/bgp/adapter/prom"
	vrf_prom "github.com/bio-routing/bio-rd/metrics/vrf/adapter/prom"
	"github.com/bio-routing/bio-rd/util/log"
)

const (
	prefix = "bio_bmp_"
)

var (
	bmpSessionEstablishedDesc    *prometheus.Desc
	routeMonitoringMessagesDesc  *prometheus.Desc
	statisticsReportMessages     *prometheus.Desc
	peerDownNotificationMessages *prometheus.Desc
	peerUpNotificationMessages   *prometheus.Desc
	initiationMessages           *prometheus.Desc
	terminationMessages          *prometheus.Desc
	routeMirroringMessages       *prometheus.Desc
)

func init() {
	labels := []string{"sys_name", "agent_address"}

	bmpSessionEstablishedDesc = prometheus.NewDesc(prefix+"session_established", "Indicates if a BMP session is established", labels, nil)
	routeMonitoringMessagesDesc = prometheus.NewDesc(prefix+"route_monitoring_messages", "Returns number of received route monitoring messages", labels, nil)
	statisticsReportMessages = prometheus.NewDesc(prefix+"statistics_report_messages", "Returns number of received statistics report messages", labels, nil)
	peerDownNotificationMessages = prometheus.NewDesc(prefix+"peer_down_messages", "Returns number of received peer down notification messages", labels, nil)
	peerUpNotificationMessages = prometheus.NewDesc(prefix+"peer_up_messages", "Returns number of received peer up notification messages", labels, nil)
	initiationMessages = prometheus.NewDesc(prefix+"initiation_messages", "Returns number of received initiation messages", labels, nil)
	terminationMessages = prometheus.NewDesc(prefix+"termination_messages", "Returns number of received termination messages", labels, nil)
	routeMirroringMessages = prometheus.NewDesc(prefix+"route_mirroring_messages", "Returns number of received route mirroring messages", labels, nil)
}

// NewCollector creates a new collector instance for the given BMP server
func NewCollector(server *server.BMPReceiver) prometheus.Collector {
	return &bmpCollector{server}
}

// bmpCollector provides a collector for BGP metrics of BIO to use with Prometheus
type bmpCollector struct {
	server *server.BMPReceiver
}

// Describe conforms to the prometheus collector interface
func (c *bmpCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- bmpSessionEstablishedDesc
	ch <- routeMonitoringMessagesDesc
	ch <- statisticsReportMessages
	ch <- peerDownNotificationMessages
	ch <- peerUpNotificationMessages
	ch <- initiationMessages
	ch <- terminationMessages
	ch <- routeMirroringMessages

	vrf_prom.DescribeRouter(ch)
	bgp_prom.DescribeRouter(ch)
}

// Collect conforms to the prometheus collector interface
func (c *bmpCollector) Collect(ch chan<- prometheus.Metric) {
	m, err := c.server.Metrics()
	if err != nil {
		log.WithError(err).Error("Could not retrieve metrics from BMP server")
		return
	}

	for _, rtr := range m.Routers {
		c.collectForRouter(ch, rtr)
	}
}

func (c *bmpCollector) collectForRouter(ch chan<- prometheus.Metric, rtr *metrics.BMPRouterMetrics) {
	l := []string{rtr.SysName, rtr.Address.String()}

	established := 0
	if rtr.Established {
		established = 1
	}

	ch <- prometheus.MustNewConstMetric(bmpSessionEstablishedDesc, prometheus.GaugeValue, float64(established), l...)
	ch <- prometheus.MustNewConstMetric(routeMonitoringMessagesDesc, prometheus.CounterValue, float64(rtr.RouteMonitoringMessages), l...)
	ch <- prometheus.MustNewConstMetric(statisticsReportMessages, prometheus.CounterValue, float64(rtr.StatisticsReportMessages), l...)
	ch <- prometheus.MustNewConstMetric(peerDownNotificationMessages, prometheus.CounterValue, float64(rtr.PeerDownNotificationMessages), l...)
	ch <- prometheus.MustNewConstMetric(peerUpNotificationMessages, prometheus.CounterValue, float64(rtr.PeerUpNotificationMessages), l...)
	ch <- prometheus.MustNewConstMetric(initiationMessages, prometheus.CounterValue, float64(rtr.InitiationMessages), l...)
	ch <- prometheus.MustNewConstMetric(terminationMessages, prometheus.CounterValue, float64(rtr.TerminationMessages), l...)
	ch <- prometheus.MustNewConstMetric(routeMirroringMessages, prometheus.CounterValue, float64(rtr.RouteMirroringMessages), l...)

	for _, vrfMetric := range rtr.VRFMetrics {
		vrf_prom.CollectForVRFRouter(ch, rtr.SysName, rtr.Address.String(), vrfMetric)
	}

	for _, peerMetric := range rtr.PeerMetrics {
		bgp_prom.CollectForPeerRouter(ch, rtr.SysName, rtr.Address.String(), peerMetric)
	}
}
