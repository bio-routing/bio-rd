package server

import (
	"fmt"
	"sync"
	"time"

	"github.com/bio-routing/bio-rd/protocols/device"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	btime "github.com/bio-routing/bio-rd/util/time"
)

// ISISServer is generic ISIS server interface
type ISISServer interface {
	AddInterface(*InterfaceConfig) error
	//GetInterfaceConfig(string) *InterfaceConfig
	//DisposeInterface(string)
	Start() error
}

//Server represents an ISIS server
type Server struct {
	running        bool
	runningMu      sync.Mutex
	nets           []*types.NET
	sequenceNumber uint32
	netIfaManager  *netIfaManager
	lsdb           *lsdb
	stop           chan struct{}
	ds             device.Updater
}

// Start starts the ISIS server
func (s *Server) Start() error {
	s.runningMu.Lock()
	defer s.runningMu.Unlock()

	if !s.running {
		s.running = true
		s.start()
	}

	return nil
}

// New creates a new ISIS server
func New(nets []*types.NET, ds device.Updater) (*Server, error) {
	if len(nets) == 0 {
		return nil, fmt.Errorf("No NETs given. One is minimum")
	}

	if !netsCompatible(nets) {
		return nil, fmt.Errorf("Incompatible NETs. System IDs must be equal")
	}

	s := &Server{
		nets:           nets,
		ds:             ds,
		sequenceNumber: 1,
		stop:           make(chan struct{}),
	}

	s.netIfaManager = newNetIfaManager(s)

	s.lsdb = newLSDB(s)
	return s, nil
}

func (s *Server) start() {
	s.lsdb.start(btime.NewBIOTicker(time.Second))
}

func (s *Server) dispose() {
	s.lsdb.dispose()
	s.lsdb = nil
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
