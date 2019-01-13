package server

import (
	"sync"

	"github.com/bio-routing/bio-rd/protocols/device"
)

type dev struct {
	srv                *Server
	name               string
	socket             int
	passive            bool
	p2p                bool
	level2             *level
	supportedProtocols []uint8
	stop               chan struct{}
	phy                *device.Device
	phyMu              sync.RWMutex
}

type level struct {
	HelloInterval uint16
	HoldTime      uint16
	Metric        uint32
	neighbors     *neighbors
}
