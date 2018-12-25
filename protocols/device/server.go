package device

import (
	"fmt"
	"sync"
)

// Server represents a device server
type Server struct {
	clientsByDevice   map[string][]Client
	clientsByDeviceMu sync.RWMutex
	done              chan struct{}
}

// Client represents a client of the device server
type Client interface {
	LinkUpdate(LinkUpdate)
}

// New creates a new device server
func New() *Server {
	return &Server{
		clientsByDevice: make(map[string][]Client),
	}
}

// Start starts the device server
func (ds *Server) Start() {
	go func() {
		err := ds.monitorDevices()
		if err != nil {
			panic(fmt.Errorf("Unable to monitor interfaces: %v", err))
		}
	}()
}

// Stop stops the device server
func (ds *Server) Stop() {
	ds.done <- struct{}{}
}

// Subscribe allows a client to subscribe for status updates on interface `devName`
func (ds *Server) Subscribe(client Client, devName string) {
	ds.clientsByDeviceMu.RLock()
	defer ds.clientsByDeviceMu.RUnlock()

	if _, ok := ds.clientsByDevice[devName]; !ok {
		ds.clientsByDevice[devName] = make([]Client, 0)
	}

	ds.clientsByDevice[devName] = append(ds.clientsByDevice[devName], client)
}
