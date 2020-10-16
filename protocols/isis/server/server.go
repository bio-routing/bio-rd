package server

import (
	"fmt"
	"sync"
	"time"

	"github.com/bio-routing/bio-rd/protocols/device"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	btime "github.com/bio-routing/bio-rd/util/time"
)

const (
	minimumLSPTransmissionIntervalS = 5
	csnpTransmissionIntervalS       = 40
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
	running          bool
	runningMu        sync.Mutex
	nets             []*types.NET
	lspLifetime      uint16
	sequenceNumberL1 uint32
	sequenceNumberL2 uint32
	netIfaManager    *netIfaManager
	lsdbL1           *lsdb
	lsdbL2           *lsdb
	stop             chan struct{}
	ds               device.Updater
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
func New(nets []*types.NET, ds device.Updater, lspLifetime uint16) (*Server, error) {
	if len(nets) == 0 {
		return nil, fmt.Errorf("No NETs given. One is minimum")
	}

	if !netsCompatible(nets) {
		return nil, fmt.Errorf("Incompatible NETs. System IDs must be equal")
	}

	s := &Server{
		nets:             nets,
		lspLifetime:      lspLifetime,
		ds:               ds,
		sequenceNumberL1: 1,
		sequenceNumberL2: 1,
		stop:             make(chan struct{}),
	}

	s.netIfaManager = newNetIfaManager(s)

	s.lsdbL1 = newLSDB(s)
	s.lsdbL2 = newLSDB(s)

	return s, nil
}

func (s *Server) start() {
	s.regenerateL2LSP()

	// TODO: Start L1 LSDB and create their LSP
	lifetimeDecrementTickerL1 := btime.NewBIOTicker(time.Second)
	lspTransmissionTickerL1 := btime.NewBIOTicker(time.Second * minimumLSPTransmissionIntervalS)
	psnpTransmissionTickerL1 := btime.NewBIOTicker(time.Second * minimumLSPTransmissionIntervalS / 3)
	csnpTransmissionTickerL1 := btime.NewBIOTicker(time.Second * csnpTransmissionIntervalS)
	s.lsdbL1.start(lifetimeDecrementTickerL1, lspTransmissionTickerL1, psnpTransmissionTickerL1, csnpTransmissionTickerL1)

	lifetimeDecrementTickerL2 := btime.NewBIOTicker(time.Second)
	lspTransmissionTickerL2 := btime.NewBIOTicker(time.Second * minimumLSPTransmissionIntervalS)
	psnpTransmissionTickerL2 := btime.NewBIOTicker(time.Second * minimumLSPTransmissionIntervalS / 3)
	csnpTransmissionTickerL2 := btime.NewBIOTicker(time.Second * csnpTransmissionIntervalS)
	s.lsdbL2.start(lifetimeDecrementTickerL2, lspTransmissionTickerL2, psnpTransmissionTickerL2, csnpTransmissionTickerL2)
}

func (s *Server) dispose() {
	s.lsdbL1.dispose()
	s.lsdbL2.dispose()
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
