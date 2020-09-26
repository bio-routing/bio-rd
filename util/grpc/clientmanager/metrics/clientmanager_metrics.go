package metrics

// ClientManagerMetrics provides metrics for a single ClientManager instance
type ClientManagerMetrics struct {
	Connections []*GRPCConnectionMetrics
}

// New returns ClientManagerMetrics
func New() *ClientManagerMetrics {
	return &ClientManagerMetrics{
		Connections: make([]*GRPCConnectionMetrics, 0),
	}
}

// GRPCConnectionMetrics represents metrics of an GRPC connection
type GRPCConnectionMetrics struct {
	Target string
	State  int
}
