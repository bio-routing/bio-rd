package metrics

type BGPAddressFamilyMetrics struct {
	AFI            uint16
	SAFI           uint8
	RoutesReceived uint64
	RoutesRejected uint64
	RoutesAccepted uint64
	RoutesSent     uint64
}
