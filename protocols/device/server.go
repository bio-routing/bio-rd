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
func New() *Server {
	srv := &Server{
		devices:         make(map[uint64]*Device),
		clientsByDevice: make(map[string][]Client),
	}

	srv.osAdapter = newOSAdapter(srv)
	return srv
}

// Start starts the device server
func (ds *Server) Start() {
	go func() {
		err := ds.monitorLinks()
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
	/*lu := ds.getLinkState(devName)
	if lu != nil {
		client.DeviceUpdate(lu)
	}

	ds.clientsByDeviceMu.RLock()
	defer ds.clientsByDeviceMu.RUnlock()

	if _, ok := ds.clientsByDevice[devName]; !ok {
		ds.clientsByDevice[devName] = make([]Client, 0)
	}

	ds.clientsByDevice[devName] = append(ds.clientsByDevice[devName], client)*/
}

func (ds *Server) addDevice(d *Device) {
	ds.devicesMu.Lock()
	defer ds.devicesMu.Unlock()

	ds.devices[d.Index] = d
}
