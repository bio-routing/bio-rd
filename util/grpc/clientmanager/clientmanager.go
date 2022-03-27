package clientmanager

import (
	"fmt"
	"sync"

	"github.com/bio-routing/bio-rd/util/grpc/clientmanager/metrics"
	"google.golang.org/grpc"
)

// ClientManager manages GRPC client connections
type ClientManager struct {
	connections   map[string]*conn
	connectionsMu sync.RWMutex
}

type conn struct {
	gc       *grpc.ClientConn
	refCount uint
}

// New creates a new ClientManager
func New() *ClientManager {
	return &ClientManager{
		connections: make(map[string]*conn),
	}
}

// Get gets a target connection and tracks it's usage
func (cm *ClientManager) Get(target string) *grpc.ClientConn {
	cm.connectionsMu.Lock()
	defer cm.connectionsMu.Unlock()

	if _, exists := cm.connections[target]; !exists {
		return nil
	}

	cm.connections[target].refCount++
	return cm.connections[target].gc
}

// AddIfNotExists adds a client if it doesn't exist already
func (cm *ClientManager) AddIfNotExists(target string, opts ...grpc.DialOption) error {
	cm.connectionsMu.Lock()
	defer cm.connectionsMu.Unlock()

	if _, exists := cm.connections[target]; exists {
		return nil
	}

	cc, err := grpc.Dial(target, opts...)
	if err != nil {
		return fmt.Errorf("grpc.Dial failed: %w", err)
	}

	cm.connections[target] = &conn{
		gc: cc,
	}

	return nil
}

// Release releases a connection if refCount reaches 0
func (cm *ClientManager) Release(target string) {
	cm.connectionsMu.Lock()
	defer cm.connectionsMu.Unlock()

	if _, exists := cm.connections[target]; exists {
		return
	}

	cm.connections[target].refCount--
	if cm.connections[target].refCount > 0 {
		return
	}

	cm.connections[target].gc.Close()
	delete(cm.connections, target)
}

// Metrics gets ClientManager metrics
func (cm *ClientManager) Metrics() *metrics.ClientManagerMetrics {
	ret := metrics.New()
	cm.connectionsMu.RLock()
	defer cm.connectionsMu.RUnlock()

	for t, c := range cm.connections {
		ret.Connections = append(ret.Connections, &metrics.GRPCConnectionMetrics{
			Target: t,
			State:  int(c.gc.GetState()),
		})
	}

	return ret
}
