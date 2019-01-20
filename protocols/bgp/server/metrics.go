package server

type metrics struct {
	server *server
}

func (b *metrics) metrics() *BGPMetrics {
	m := &BGPMetrics{
		RouterID:  b.server.RouterID,
		LocalASN:  b.server.LocalASN,
		Neighbors: b.neighborMetrics(),
	}
}

func (b *metrics) neighborMetrics() []*BGPNeighborMetrics {
	neighbors := make([]*BGPNeighborMetrics, 0)

	for _, peer := range b.server.peers {

	}

	return neighbors
}
