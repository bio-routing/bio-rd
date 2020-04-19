package server

import (
	"sync"
	"time"

	"github.com/bio-routing/bio-rd/protocols/device"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	btime "github.com/bio-routing/bio-rd/util/time"
)

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
	//devices        *devices
	lsdb *lsdb
	stop chan struct{}
	ds   device.Updater
}

func (s *Server) Start() error {
	s.runningMu.Lock()
	defer s.runningMu.Unlock()

	if !s.running {
		s.running = true
		s.start()
	}

	return nil
}

func New(nets []*types.NET, ds device.Updater) *Server {
	s := &Server{
		nets:           nets,
		ds:             ds,
		sequenceNumber: 1,
		stop:           make(chan struct{}),
	}

	//s.devices = newDevices(s)
	s.lsdb = newLSDB(s)
	return s
}

func (s *Server) start() {
	s.lsdb.start(btime.NewBIOTicker(time.Second))
}

func (s *Server) dispose() {
	s.lsdb.dispose()
	s.lsdb = nil
}
