package metrics

type BMPMetrics struct {
	Routers []*BMPRouterMetrics
}

type BMPRouterMetrics struct {
	// Name of the monitored routers
	Name string

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
}
