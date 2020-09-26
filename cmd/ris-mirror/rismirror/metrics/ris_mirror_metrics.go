package metrics

import (
	"net"

	mlrib_metrics "github.com/bio-routing/bio-rd/routingtable/mergedlocrib/metrics"
	vrf_metrics "github.com/bio-routing/bio-rd/routingtable/vrf/metrics"
)

// RISMirrorMetrics contains per router BMP metrics
type RISMirrorMetrics struct {
	Routers []*RISMirrorRouterMetrics
}

// RISMirrorRouterMetrics contains a routers RIS mirror metrics
type RISMirrorRouterMetrics struct {
	Address            net.IP
	SysName            string
	VRFMetrics         []*vrf_metrics.VRFMetrics
	InternalVRFMetrics []*InternalVRFMetrics
}

// InternalVRFMetrics represents internal VRF metrics (_vrf)
type InternalVRFMetrics struct {
	RD                             uint64
	MergedLocRIBMetricsIPv4Unicast *mlrib_metrics.MergedLocRIBMetrics
	MergedLocRIBMetricsIPv6Unicast *mlrib_metrics.MergedLocRIBMetrics
}
