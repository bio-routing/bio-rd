package metrics

import (
	"net"

	vrf_metrics "github.com/bio-routing/bio-rd/routingtable/vrf/metrics"
)

// BMPMetrics contains per router BMP metrics
type BMPMetrics struct {
	Routers []*BMPRouterMetrics
}

// BMPRouterMetrics contains a routers BMP metrics
type BMPRouterMetrics struct {
	// Routers IP Address
	Address net.IP

	// SysName of the monitored router
	SysName string

	// Status of TCP session
	Established bool

	// Count of received RouteMonitoringMessages
	RouteMonitoringMessages uint64

	// Count of received StatisticsReportMessages
	StatisticsReportMessages uint64

	// Count of received PeerDownNotificationMessages
	PeerDownNotificationMessages uint64

	// Count of received PeerUpNotificationMessages
	PeerUpNotificationMessages uint64

	// Count of received InitiationMessages
	InitiationMessages uint64

	// Count of received TerminationMessages
	TerminationMessages uint64

	// Count of received RouteMirroringMessages
	RouteMirroringMessages uint64

	// VRFMetrics represent per VRF metrics
	VRFMetrics []*vrf_metrics.VRFMetrics

	// PeerMetrics contains BGP per peer metrics
	PeerMetrics []*BGPPeerMetrics
}
