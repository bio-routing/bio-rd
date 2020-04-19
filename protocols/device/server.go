package device

import (
	"sync"

	"github.com/pkg/errors"
)

// Updater is a device updater interface
type Updater interface {
	Subscribe(Client, string)
	Unsubscribe(Client, string)
}

// Server represents a device server
type Server struct {
	devices           map[uint64]*Device
	devicesMu         sync.RWMutex
	clientsByDevice   map[string][]Client
	clientsByDeviceMu sync.RWMutex
	osAdapter         osAdapter
	done              chan struct{}
}

// Client represents a client of the device server
type Client interface {
	DeviceUpdate(DeviceInterface)
}

type osAdapter interface {
	start() error
}

// New creates a new device server
func New() (*Server, error) {
	srv := newWithAdapter(nil)
	err := srv.loadAdapter()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create OS adapter")
	}

	return srv, nil
}

func newWithAdapter(a osAdapter) *Server {
	return &Server{
		devices:         make(map[uint64]*Device),
		clientsByDevice: make(map[string][]Client),
		osAdapter:       a,
		done:            make(chan struct{}),
	}
}

// Start starts the device server
func (ds *Server) Start() error {
	err := ds.osAdapter.start()
	if err != nil {
		return errors.Wrap(err, "Unable to start osAdapter")
	}

	return nil
}

// Stop stops the device server
func (ds *Server) Stop() {
	close(ds.done)
}

// Subscribe allows a client to subscribe for status updates on interface `devName`
func (ds *Server) Subscribe(client Client, devName string) {
	d := ds.getLinkState(devName)
	if d != nil {
		client.DeviceUpdate(d)
	}

	ds.clientsByDeviceMu.Lock()
	defer ds.clientsByDeviceMu.Unlock()

	if _, ok := ds.clientsByDevice[devName]; !ok {
		ds.clientsByDevice[devName] = make([]Client, 0)
	}

	ds.clientsByDevice[devName] = append(ds.clientsByDevice[devName], client)
}

// Unsubscribe unsubscribes a client
func (ds *Server) Unsubscribe(client Client, devName string) {
	ds.clientsByDeviceMu.Lock()
	defer ds.clientsByDeviceMu.Unlock()

	if _, ok := ds.clientsByDevice[devName]; !ok {
		return
	}

	for i := range ds.clientsByDevice[devName] {
		if ds.clientsByDevice[devName][i] != client {
			continue
		}

		ds.clientsByDevice[devName] = append(ds.clientsByDevice[devName][:i], ds.clientsByDevice[devName][i+1:]...)
		return
	}
}

func (ds *Server) addDevice(d *Device) {
	ds.devicesMu.Lock()
	defer ds.devicesMu.Unlock()

	ds.devices[d.index] = d
}

func (ds *Server) delDevice(index uint64) {
	delete(ds.devices, index)
}

func (ds *Server) getLinkState(name string) *Device {
	ds.devicesMu.RLock()
	defer ds.devicesMu.RUnlock()

	for _, d := range ds.devices {
		if d.name != name {
			continue
		}

		return d.copy()
	}

	return nil
}

func (ds *Server) notify(index uint64) {
	ds.clientsByDeviceMu.RLock()
	defer ds.clientsByDeviceMu.RUnlock()

	for i, d := range ds.devices {
		if i != index {
			continue
		}

		for _, c := range ds.clientsByDevice[d.name] {
			c.DeviceUpdate(d.copy())
		}
	}
}
