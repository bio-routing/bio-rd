package metrics

import "github.com/prometheus/client_golang/prometheus"

func RegisterMetrics(registry prometheus.Registerer) {
	registry.MustRegister(PathUpdates)
	registry.MustRegister(CurrentFSMState)
}
