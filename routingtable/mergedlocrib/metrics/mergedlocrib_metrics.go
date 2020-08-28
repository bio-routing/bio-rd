package metrics

// MergedLocRIBMetrics represents merged local rib metrics
type MergedLocRIBMetrics struct {
	RIBName                     string
	UniqueRouteCount            uint64
	RoutesWithSingleSourceCount uint64
}
