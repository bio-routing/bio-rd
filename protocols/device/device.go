package device

import (
	"net"
	"sync"
)

const (
	IfAdminDown = 0
	IfLinkDown  = 1
	IfUp        = 2
	IfUnknown   = 255
)

// Device represents a network interface
type device struct {
	Name      string
	IfIndex   uint64
	Status    uint8
	clients   []Client
	clientsMu sync.RWMutex
}

type LinkUpdate struct {
	Index        uint64
	MTU          uint16
	Name         string
	HardwareAddr net.HardwareAddr
	Flags        net.Flags
	OperState    uint8
}

func newDevice(name string, ifIndex uint64, status uint8) *device {
	return &device{
		Name:    name,
		IfIndex: ifIndex,
		Status:  status,
		clients: make([]Client, 0, 5),
	}
}
