package metrics

// VRFMetrics represents a collection of metrics of one VRF
type VRFMetrics struct {
	// Name of the VRF
	Name string

	// RIBs returns the RIB specific metrics
	RIBs []*RIBMetrics
}
