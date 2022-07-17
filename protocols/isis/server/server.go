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
	minimumLSPTransmissionInterval = time.Second * 5
	csnpTransmissionInterval       = time.Second * 10
)

// ISISServer is generic ISIS server interface
type ISISServer interface {
	AddInterface(*InterfaceConfig) error
	RemoveInterface(name string) error
	GetInterfaceNames() []string
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
	lsdbL1             *lsdb
	lsdbL2             *lsdb
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

	decrementTicker := btime.NewBIOTicker(time.Second)
	minLSPTransTicker := btime.NewBIOTicker(minimumLSPTransmissionInterval)
	psnpTransTicker := btime.NewBIOTicker(time.Second * 5)
	csnpTransTicker := btime.NewBIOTicker(csnpTransmissionInterval)
	s.lsdbL2.start(decrementTicker, minLSPTransTicker, psnpTransTicker, csnpTransTicker)

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
	s.lsdbL2 = newLSDB(s)

	return s, nil
}

// netsCompatible verifies if the system id is equal in all NETs
func netsCompatible(nets []*types.NET) bool {
	first := nets[0].SystemID
	for _, net := range nets {
		if first != net.SystemID {
			return false
		}
	}

	return true
}

// GetInterfaceNames gets names of all configured interfaces
func (s *Server) GetInterfaceNames() []string {
	ret := make([]string, 0)
	for _, x := range s.netIfaManager.getAllInterfaces() {
		ret = append(ret, x.name)
	}

	return ret
}
