package metrics

// BGPAddressFamilyMetrics provides metrics on AFI/SAFI level for one session
type BGPAddressFamilyMetrics struct {
	// AFI is the identifier for the address family
	AFI uint16

	// SAFI is the identifier for the sub address family
	SAFI uint8

	// RoutesReceived is the number of routes we recevied
	RoutesReceived uint64

	// RoutesAccepted is the number of routes we sent
	RoutesSent uint64

	// EndOfRIBMarkerReceived indicates if a BGP End of RIB marker was received for this AFI/SAFI from the peer
	EndOfRIBMarkerReceived bool
}
