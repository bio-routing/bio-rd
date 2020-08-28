package clientmanager

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// ClientManager manages GRPC client connections
type ClientManager struct {
	connections   map[string]*grpc.ClientConn
	connectionsMu sync.RWMutex
}

// New creates a new ClientManager
func New() *ClientManager {
	return &ClientManager{
		connections: make(map[string]*grpc.ClientConn),
	}
}

// Get gets a target connection
func (cm *ClientManager) Get(target string) *grpc.ClientConn {
	cm.connectionsMu.RLock()
	defer cm.connectionsMu.RUnlock()

	if _, exists := cm.connections[target]; !exists {
		return nil
	}

	return cm.connections[target]
}

// Add adds a target
func (cm *ClientManager) Add(target string, opts ...grpc.DialOption) error {
	cm.connectionsMu.Lock()
	defer cm.connectionsMu.Unlock()

	if _, exists := cm.connections[target]; exists {
		return fmt.Errorf("Target exists already")
	}

	cc, err := grpc.Dial(target, opts...)
	if err != nil {
		return errors.Wrap(err, "grpc.Dial failed")
	}

	cm.connections[target] = cc
	return nil
}
