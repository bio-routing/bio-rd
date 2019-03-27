package metrics

// RIBMetrics represents metrics of a RIB in a VRF
type RIBMetrics struct {
	// Name of the RIB
	Name string

	// AFI is the identifier for the address family
	AFI uint16

	// SAFI is the identifier for the sub address family
	SAFI uint8

	// Number of routes in the RIB
	RouteCount uint64
}
