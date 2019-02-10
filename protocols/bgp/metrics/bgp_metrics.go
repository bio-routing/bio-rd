package metrics

// BGPMetrics provides metrics for a single BGP server instance
type BGPMetrics struct {
	// OpenReceived is the number of open messages recevied
	OpenReceived uint64
	// OpenSent is the number of open messages sent
	OpenSent uint64
	// Peers is the collection of per peer metrics
	Peers []*BGPPeerMetrics
}
