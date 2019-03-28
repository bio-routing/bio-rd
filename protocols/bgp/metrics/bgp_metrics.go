package metrics

// BGPMetrics provides metrics for a single BGP server instance
type BGPMetrics struct {
	// Peers is the collection of per peer metrics
	Peers []*BGPPeerMetrics
}
