package prom

import (
	"strconv"
	"time"

	"github.com/bio-routing/bio-rd/protocols/bgp/metrics"
	"github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const (
	prefix = "bio_bgp_"
)

var (
	upDesc              *prometheus.Desc
	stateDesc           *prometheus.Desc
	uptimeDesc          *prometheus.Desc
	updatesReceivedDesc *prometheus.Desc
	updatesSentDesc     *prometheus.Desc
	routesReceivedDesc  *prometheus.Desc
	routesSentDesc      *prometheus.Desc
	routesRejectedDesc  *prometheus.Desc
	routesAcceptedDesc  *prometheus.Desc
)

func init() {
	labels := []string{"peer_ip", "local_asn", "peer_asn", "vrf"}
	upDesc = prometheus.NewDesc(prefix+"up", "Returns if the session is up", labels, nil)
	stateDesc = prometheus.NewDesc(prefix+"state", "State of the BGP session (Down = 0, Idle = 1, Connect = 2, Active = 3, OpenSent = 4, OpenConfirm = 5, Established = 6)", labels, nil)
	uptimeDesc = prometheus.NewDesc(prefix+"uptime_second", "Time since the session was established in seconds", labels, nil)
	updatesReceivedDesc = prometheus.NewDesc(prefix+"update_received_count", "Number of updates received", labels, nil)
	updatesSentDesc = prometheus.NewDesc(prefix+"update_sent_count", "Number of updates sent", labels, nil)

	labels = append(labels, "afi", "safi")
	routesReceivedDesc = prometheus.NewDesc(prefix+"route_received_count", "Number of routes received", labels, nil)
	routesSentDesc = prometheus.NewDesc(prefix+"route_sent_count", "Number of routes sent", labels, nil)
	routesRejectedDesc = prometheus.NewDesc(prefix+"route_rejected_count", "Number of routes rejected", labels, nil)
	routesAcceptedDesc = prometheus.NewDesc(prefix+"route_accepted_count", "Number of routes accepted", labels, nil)
}

// NewCollector creates a new collector instance for the given BGP server
func NewCollector(server server.BGPServer) prometheus.Collector {
	return &bgpCollector{server}
}

// BGPCollector provides a collector for BGP metrics of BIO to use with Prometheus
type bgpCollector struct {
	server server.BGPServer
}

// Describe conforms to the prometheus collector interface
func (c *bgpCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- upDesc
	ch <- stateDesc
	ch <- uptimeDesc
	ch <- updatesReceivedDesc
	ch <- updatesSentDesc
	ch <- routesReceivedDesc
	ch <- routesSentDesc
	ch <- routesRejectedDesc
	ch <- routesAcceptedDesc
}

// Collect conforms to the prometheus collector interface
func (c *bgpCollector) Collect(ch chan<- prometheus.Metric) {
	m, err := c.server.Metrics()
	if err != nil {
		log.Error(errors.Wrap(err, "Could not retrieve metrics from BGP server"))
		return
	}

	for _, peer := range m.Peers {
		c.collectForPeer(ch, peer)
	}
}

func (c *bgpCollector) collectForPeer(ch chan<- prometheus.Metric, peer *metrics.BGPPeerMetrics) {
	l := []string{
		peer.IP.String(),
		strconv.Itoa(int(peer.LocalASN)),
		strconv.Itoa(int(peer.ASN)),
		peer.VRF}

	var up float64
	var uptime float64
	if peer.Up {
		up = 1
		uptime = float64(time.Since(peer.Since) * time.Second)
	}
	ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, up, l...)
	ch <- prometheus.MustNewConstMetric(uptimeDesc, prometheus.GaugeValue, uptime, l...)
	ch <- prometheus.MustNewConstMetric(stateDesc, prometheus.GaugeValue, float64(peer.State), l...)

	ch <- prometheus.MustNewConstMetric(updatesReceivedDesc, prometheus.CounterValue, float64(peer.UpdatesReceived), l...)
	ch <- prometheus.MustNewConstMetric(updatesSentDesc, prometheus.CounterValue, float64(peer.UpdatesSent), l...)

	for _, family := range peer.AddressFamilies {
		c.collectForFamily(ch, family, l)
	}
}

func (c *bgpCollector) collectForFamily(ch chan<- prometheus.Metric, family *metrics.BGPAddressFamilyMetrics, l []string) {
	l = append(l, strconv.Itoa(int(family.AFI)), strconv.Itoa(int(family.SAFI)))

	ch <- prometheus.MustNewConstMetric(routesReceivedDesc, prometheus.CounterValue, float64(family.RoutesReceived), l...)
	ch <- prometheus.MustNewConstMetric(routesSentDesc, prometheus.CounterValue, float64(family.RoutesSent), l...)
}
