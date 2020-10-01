package prom

import (
	"github.com/bio-routing/bio-rd/util/grpc/clientmanager"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	prefix = "bio_grpc_clientmanager_"
)

var (
	connectionStateDesc *prometheus.Desc
)

func init() {
	labels := []string{"target"}
	connectionStateDesc = prometheus.NewDesc(prefix+"connection_state", "Connection state, 0=IDLE,1=CONNECTING,2=READY,3=TRANSIENT_FAILURE,4=SHUTDOWN", labels, nil)
}

// NewCollector creates a new collector instance for the given clientmanager
func NewCollector(cm *clientmanager.ClientManager) prometheus.Collector {
	return &grpcClientManagerCollector{
		cm: cm,
	}
}

// grpcClientManagerCollector provides a collector for RIS metrics of BIO to use with Prometheus
type grpcClientManagerCollector struct {
	cm *clientmanager.ClientManager
}

// Describe conforms to the prometheus collector interface
func (c *grpcClientManagerCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- connectionStateDesc
}

// Collect conforms to the prometheus collector interface
func (c *grpcClientManagerCollector) Collect(ch chan<- prometheus.Metric) {
	for _, con := range c.cm.Metrics().Connections {
		l := []string{con.Target}
		ch <- prometheus.MustNewConstMetric(connectionStateDesc, prometheus.GaugeValue, float64(con.State), l...)
	}
}
