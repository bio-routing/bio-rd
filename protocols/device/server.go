package device

import (
	"fmt"
	"sync"
)

// Server represents a device server
type Server struct {
	devices   []*device
	devicesMu sync.RWMutex
	done      chan struct{}
}

// Client represents a client of the device server
type Client interface {
	LinkUpdate(LinkUpdate)
}

// New creates a new device server
func New() *Server {
	return &Server{
		devices: make([]*device, 0),
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
	ds.devicesMu.RLock()
	defer ds.devicesMu.RUnlock()

	for _, d := range ds.devices {
		if d.Name != devName {
			fmt.Printf("Skipped> %s", d.Name)
			continue
		}

		panic("found")
		d.clientsMu.Lock()
		defer d.clientsMu.Unlock()

		for _, c := range d.clients {
			if c == client {
				return
			}
		}

		d.clients = append(d.clients, client)
		return
	}

	ds.devices = append(ds.devices, newDevice(devName, 0, IfUnknown))
}
