package server

import (
	"fmt"
	"sync"

	"github.com/bio-routing/bio-rd/protocols/device"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
)

const (
	minimumLSPTransmissionIntervalS = 5
	csnpTransmissionIntervalS       = 40
)

// ISISServer is generic ISIS server interface
type ISISServer interface {
	AddInterface(*InterfaceConfig) error
	Start() error
}

//Server represents an ISIS server
type Server struct {
	running            bool
	runningMu          sync.Mutex
	nets               []*types.NET
	lspLifetime        uint16
	sequenceNumberL1   uint32
	sequenceNumberL1Mu sync.Mutex
	sequenceNumberL2   uint32
	sequenceNumberL2Mu sync.Mutex
	netIfaManager      *netIfaManager
	stop               chan struct{}
	ds                 device.Updater
}

// Start starts the ISIS server
func (s *Server) Start() error {
	s.runningMu.Lock()
	defer s.runningMu.Unlock()

	if !s.running {
		s.running = true
	}

	return nil
}

// New creates a new ISIS server
func New(nets []*types.NET, ds device.Updater, lspLifetime uint16) (*Server, error) {
	if len(nets) == 0 {
		return nil, fmt.Errorf("No NETs given. One is minimum")
	}

	if !netsCompatible(nets) {
		return nil, fmt.Errorf("Incompatible NETs. System IDs must be equal")
	}

	s := &Server{
		nets:        nets,
		lspLifetime: lspLifetime,
		ds:          ds,
		stop:        make(chan struct{}),
	}

	s.netIfaManager = newNetIfaManager(s)

	return s, nil
}

// netsCompatible verifies if the system id is equal in all NETs
func netsCompatible(nets []*types.NET) bool {
	if len(nets) <= 1 {
		return true
	}

	first := nets[0].SystemID
	for _, net := range nets {
		if first != net.SystemID {
			return false
		}
	}

	return true
}
