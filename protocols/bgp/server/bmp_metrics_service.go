package server

import (
	"sync/atomic"

	"github.com/bio-routing/bio-rd/protocols/bgp/metrics"
)

type bmpMetricsService struct {
	server *BMPServer
}

func (b *bmpMetricsService) metrics() *metrics.BMPMetrics {
	return &metrics.BMPMetrics{
		Routers: b.routerMetrics(),
	}
}

func (b *bmpMetricsService) routerMetrics() []*metrics.BMPRouterMetrics {
	routers := make([]*metrics.BMPRouterMetrics, 0)

	for _, rtr := range b.server.getRouters() {
		routers = append(routers, b.metricsForRouter(rtr))
	}

	return routers
}

func (b *bmpMetricsService) metricsForRouter(rtr *Router) *metrics.BMPRouterMetrics {
	rm := &metrics.BMPRouterMetrics{
		Name:                         rtr.name,
		RouteMonitoringMessages:      atomic.LoadUint64(&rtr.counters.routeMonitoringMessages),
		StatisticsReportMessages:     atomic.LoadUint64(&rtr.counters.statisticsReportMessages),
		PeerDownNotificationMessages: atomic.LoadUint64(&rtr.counters.peerDownNotificationMessages),
		PeerUpNotificationMessages:   atomic.LoadUint64(&rtr.counters.peerUpNotificationMessages),
		InitiationMessages:           atomic.LoadUint64(&rtr.counters.initiationMessages),
		TerminationMessages:          atomic.LoadUint64(&rtr.counters.terminationMessages),
		RouteMirroringMessages:       atomic.LoadUint64(&rtr.counters.routeMirroringMessages),
	}

	return rm
}
