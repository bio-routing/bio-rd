package device

import (
	"fmt"
	"sync"
)

// Server represents a device server
type Server struct {
	devices           map[uint64]*Device
	devicesMu         sync.RWMutex
	clientsByDevice   map[string][]Client
	clientsByDeviceMu sync.RWMutex
	osAdapter         *osAdapter
	done              chan struct{}
}

// Client represents a client of the device server
type Client interface {
	DeviceUpdate(*Device)
}

// New creates a new device server
func New() (*Server, error) {
	srv := &Server{
		devices:         make(map[uint64]*Device),
		clientsByDevice: make(map[string][]Client),
	}

	o, err := newOSAdapter(srv)
	if err != nil {
		return nil, fmt.Errorf("Unable to create OS adapter: %v", err)
	}

	srv.osAdapter = o
	return srv, nil
}

// Start starts the device server
func (ds *Server) Start() error {
	err := ds.osAdapter.start()
	if err != nil {
		return fmt.Errorf("Unable to start osAdapter: %v", err)
	}

	return nil
}

// Stop stops the device server
func (ds *Server) Stop() {
	ds.done <- struct{}{}
}

// Subscribe allows a client to subscribe for status updates on interface `devName`
func (ds *Server) Subscribe(client Client, devName string) {
	d := ds.getLinkState(devName)
	if d != nil {
		client.DeviceUpdate(d)
	}

	ds.clientsByDeviceMu.RLock()
	defer ds.clientsByDeviceMu.RUnlock()

	if _, ok := ds.clientsByDevice[devName]; !ok {
		ds.clientsByDevice[devName] = make([]Client, 0)
	}

	ds.clientsByDevice[devName] = append(ds.clientsByDevice[devName], client)
}

func (ds *Server) addDevice(d *Device) {
	ds.devicesMu.Lock()
	defer ds.devicesMu.Unlock()

	ds.devices[d.Index] = d
}

func (ds *Server) getLinkState(name string) *Device {
	ds.devicesMu.RLock()
	defer ds.devicesMu.RUnlock()

	for _, d := range ds.devices {
		if d.Name != name {
			continue
		}

		return d.copy()
	}

	return nil
}
