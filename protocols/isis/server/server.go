package server

import (
	"fmt"
	"os"
	"sync"
	"time"

	bbclock "github.com/benbjohnson/clock"
	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/net/ethernet"
	"github.com/bio-routing/bio-rd/protocols/device"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
)

var (
	clock = bbclock.New()
)

func SetClock(c bbclock.Clock) {
	clock = c
}

const (
	minimumLSPTransmissionInterval = time.Second * 5
	csnpTransmissionInterval       = time.Second * 10
)

// ISISServer is generic ISIS server interface
type ISISServer interface {
	AddInterface(*InterfaceConfig) error
	RemoveInterface(name string) error
	GetInterfaceNames() []string
	Start()
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
	hostname                 func() (string, error)
}

// Start starts the ISIS server
func (s *Server) Start() {
	s.runningMu.Lock()
	defer s.runningMu.Unlock()

	if !s.running {
		s.running = true
	}

	s.lsdbL2.updateL2LSP()

	decrementTicker := clock.Ticker(time.Second)
	minLSPTransTicker := clock.Ticker(minimumLSPTransmissionInterval)
	psnpTransTicker := clock.Ticker(time.Second * 5)
	csnpTransTicker := clock.Ticker(csnpTransmissionInterval)
	s.lsdbL2.start(decrementTicker, minLSPTransTicker, psnpTransTicker, csnpTransTicker)
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
		hostname:                 os.Hostname,
	}

	s.netIfaManager = newNetIfaManager(s)
	s.lsdbL2 = newLSDB(s)

	return s, nil
}

func (s *Server) SetEthernetInterfaceFactory(f ethernet.EthernetInterfaceFactoryI) {
	s.ethernetInterfaceFactory = f
}

func (s *Server) SetHostnameFunc(f func() (string, error)) {
	s.hostname = f
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

// updateL2LSP updates the systems L2 LSP. This is triggered when:
// 1. Router starts up (done)
// 2. Periodic refresh timer expired (done)
// 3. A new adjacency is formed (done)
// 4. An adjacency goes down (done)
// 5. A link goes down (done)
// 6. Metric associated with a link or reachable address changes (todo)
// 7. The routers sysID changes  (todo)
// 8. The router is elected or superseded as the DIS (todo)
// 9. An area address associated with the router is added or removed (todo)
// 10. The overload status of the database changes (todo)
func (s *Server) updateL2LSP() {
	s.lsdbL2.requestL2LSPUpdate()
}
