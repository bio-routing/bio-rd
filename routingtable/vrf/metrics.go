package vrf

import (
	"github.com/bio-routing/bio-rd/routingtable/vrf/metrics"
)

// Metrics returns metrics for all VRFs
func Metrics(r *VRFRegistry) []*metrics.VRFMetrics {
	vrfs := r.List()

	m := make([]*metrics.VRFMetrics, len(vrfs))
	i := 0
	for _, v := range vrfs {
		m[i] = MetricsForVRF(v)
		i++
	}

	return m
}

func MetricsForVRF(v *VRF) *metrics.VRFMetrics {
	m := &metrics.VRFMetrics{
		Name: v.Name(),
		RIBs: make([]*metrics.RIBMetrics, 0),
	}

	for family, rib := range v.ribs {
		m.RIBs = append(m.RIBs, &metrics.RIBMetrics{
			Name:       v.nameForRIB(rib),
			AFI:        family.afi,
			SAFI:       family.safi,
			RouteCount: rib.Count(),
		})
	}

	return m
}
