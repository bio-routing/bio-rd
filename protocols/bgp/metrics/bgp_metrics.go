package metrics

type BGPMetrics struct {
	OpenReceived uint64
	OpenSent     uint64
	Neighbors    []*BGPPeerMetrics
}
