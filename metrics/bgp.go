package metrics

import "github.com/prometheus/client_golang/prometheus"

const (
	metrics_namespace     = "bio_routing"
	metrics_subsystem_bgp = "bgp"
	AddPathAction         = "add_path"
	RemovePathAction      = "remove_path"
)

var (
	PathUpdates = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics_namespace,
			Subsystem: metrics_subsystem_bgp,
			Name:      "path_updates_total",
			Help:      "Processed path updated by component",
		},
		[]string{"component", "action"},
	)
	CurrentFSMState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metrics_namespace,
			Subsystem: metrics_subsystem_bgp,
			Name:      "current_fsm_state",
			Help:      "Numeric representation of the current fsm state",
		},
		[]string{"peer"},
	)
)
