package server

import (
	"time"

	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/protocols/device"
	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	btime "github.com/bio-routing/bio-rd/util/time"
	log "github.com/sirupsen/logrus"
)

//Server represents an ISIS server
type Server struct {
	config             *config.ISISConfig
	sequenceNumber     uint32
	dm                 *devicesManager
	nm                 *neighborManager
	lsdb               *lsdb
	stop               chan struct{}
	ds                 device.Updater
	supportedProtocols []uint8
	log                *log.Logger
}

// New creates a new IS-IS server
func New(cfg *config.ISISConfig, ds device.Updater, l *log.Logger) *Server {
	s := &Server{
		config:             cfg,
		ds:                 ds,
		sequenceNumber:     1,
		stop:               make(chan struct{}),
		supportedProtocols: []uint8{packet.NLPIDIPv4, packet.NLPIDIPv6},
		log:                l,
	}

	s.nm = newNeighborManager(s)
	s.dm = newDevicesManager(s)
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

// AddInterface adds an interface to the ISIS Server
func (s *Server) AddInterface(ifcfg *config.ISISInterfaceConfig) {
	s.dm.addDevice(ifcfg)
}

// RemoveInterface removes an interface from the ISIS Server
func (s *Server) RemoveInterface(name string) {
	s.dm.removeDevice(name)
}

func (s *Server) systemID() [6]byte {
	return s.config.NETs[0].SystemID
}

func (s *Server) getAreas() []types.AreaID {
	areas := make([]types.AreaID, len(s.config.NETs))
	for i, NET := range s.config.NETs {
		areas[i] = append([]byte{NET.AFI}, NET.AreaID...)
	}

	return areas
}

// GetDatabase get's the LSDBs save LSPDUs
func (s *Server) GetDatabase() []*packet.LSPDU {
	return s.lsdb.dump()
}
