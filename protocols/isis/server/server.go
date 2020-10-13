package server

import (
	"fmt"
	"sync"
	"time"

	"github.com/bio-routing/bio-rd/protocols/device"
	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	btime "github.com/bio-routing/bio-rd/util/time"
)

const (
	minimumLSPTransmissionIntervalMS = 100
	initialLSPSequenceNumber         = 1
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
func New(nets []*types.NET, ds device.Updater) (*Server, error) {
	if len(nets) == 0 {
		return nil, fmt.Errorf("No NETs given. One is minimum")
	}

	if !netsCompatible(nets) {
		return nil, fmt.Errorf("Incompatible NETs. System IDs must be equal")
	}

	s := &Server{
		nets:             nets,
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
	s.generateL2LSP(initialLSPSequenceNumber)
	// TODO: Start L1 LSDBs and create their LSPs
	s.lsdbL1.start(btime.NewBIOTicker(time.Second), btime.NewBIOTicker(time.Millisecond*minimumLSPTransmissionIntervalMS), btime.NewBIOTicker(time.Millisecond*minimumLSPTransmissionIntervalMS/3))
	s.lsdbL2.start(btime.NewBIOTicker(time.Second), btime.NewBIOTicker(time.Millisecond*minimumLSPTransmissionIntervalMS), btime.NewBIOTicker(time.Millisecond*minimumLSPTransmissionIntervalMS/3))
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

func (s *Server) generateL2LSP(sequenceNumber uint32) {
	lsp := &packet.LSPDU{
		RemainingLifetime: 3600,
		LSPID: packet.LSPID{
			SystemID:     s.nets[0].SystemID,
			PseudonodeID: 0,
			LSPNumber:    0,
		},
		SequenceNumber: sequenceNumber,
	}

	lsp.TypeBlock |= 0x3 // level2, last two bits
	lsp.UpdateLength()
	lsp.SetChecksum()

	s.lsdbL2.lspsMu.Lock()
	defer s.lsdbL2.lspsMu.Unlock()

	s.lsdbL2.lsps[lsp.LSPID] = newLSDBEntry(lsp)

	// TODO: Set SRM?
}
