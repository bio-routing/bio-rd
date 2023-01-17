package server

import (
	"fmt"
	"sync"
	"time"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/net/ethernet"
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
	GetAdjacencies() []*Adjacency
	GetLSDB() []*LSDBEntry
}

// Server represents an ISIS server
type Server struct {
	running                  bool
	runningMu                sync.Mutex
	nets                     []*types.NET
	lspLifetime              uint16
	sequenceNumberL1         uint32
	sequenceNumberL1Mu       sync.Mutex
	sequenceNumberL2         uint32
	sequenceNumberL2Mu       sync.Mutex
	netIfaManager            *netIfaManager
	lsdbL1                   *lsdb
	lsdbL2                   *lsdb
	stop                     chan struct{}
	ds                       device.Updater
	ethernetInterfaceFactory ethernet.EthernetInterfaceFactoryI
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

type Adjacency struct {
	Name            string
	SystemID        types.SystemID
	Address         ethernet.MACAddr
	InterfaceName   string
	Level           uint8
	Priority        uint8
	IPAddresses     []bnet.IP
	LastStateChange time.Time
	Timeout         time.Time
	Status          uint8
}

func (s *Server) GetAdjacencies() []*Adjacency {
	ret := make([]*Adjacency, 0)

	for _, ifa := range s.netIfaManager.getAllInterfaces() {
		for _, n := range ifa.neighborManagerL2.getNeighbors() {
			ret = append(ret, n.getAdjacency())
		}
	}

	return ret
}

func (s *Server) GetLSDB() []*LSDBEntry {
	s.lsdbL2.lspsMu.RLock()
	defer s.lsdbL2.lspsMu.RUnlock()

	ret := make([]*LSDBEntry, 0)
	for _, lspEntry := range s.lsdbL2.lsps {
		ret = append(ret, lspEntry.Export())
	}

	return ret
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
		nets:                     nets,
		lspLifetime:              lspLifetime,
		ds:                       ds,
		stop:                     make(chan struct{}),
		ethernetInterfaceFactory: ethernet.NewEthernetInterfaceFactory(),
	}

	s.netIfaManager = newNetIfaManager(s)
	s.lsdbL2 = newLSDB(s)

	return s, nil
}

func (s *Server) SetEthernetInterfaceFactory(f ethernet.EthernetInterfaceFactoryI) {
	s.ethernetInterfaceFactory = f
}

func (s *Server) GetEthernetInterface(name string) ethernet.EthernetInterfaceI {
	ifa := s.netIfaManager.getInterface(name)
	if ifa == nil {
		return nil
	}

	return ifa.ethernetInterface
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

func (s *Server) systemID() types.SystemID {
	return s.nets[0].SystemID
}

func (s *Server) areaIDs() []types.AreaID {
	areaIDs := make([]types.AreaID, 0, len(s.nets))
	for _, net := range s.nets {
		areaIDs = append(areaIDs, net.AreaID)
	}

	return areaIDs
}

// GetInterfaceNames gets names of all configured interfaces
func (s *Server) GetInterfaceNames() []string {
	ret := make([]string, 0)
	for _, x := range s.netIfaManager.getAllInterfaces() {
		ret = append(ret, x.name)
	}

	return ret
}
