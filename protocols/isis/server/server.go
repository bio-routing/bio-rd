package server

import (
	"time"

	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/protocols/device"
	btime "github.com/bio-routing/bio-rd/util/time"
	log "github.com/sirupsen/logrus"
)

//Server represents an ISIS server
type Server struct {
	config         *config.ISISConfig
	sequenceNumber uint32
	devices        *devices
	lsdb           *lsdb
	stop           chan struct{}
	ds             device.Updater
	log            *log.Logger
}

func New(cfg *config.ISISConfig, ds device.Updater, l *log.Logger) *Server {
	s := &Server{
		config:         cfg,
		ds:             ds,
		sequenceNumber: 1,
		stop:           make(chan struct{}),
		log:            l,
	}

	s.devices = newDevices(s)
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
	s.devices.addDevice(ifcfg)
}

// RemoveInterface removes an interface from the ISIS Server
func (s *Server) RemoveInterface(name string) {
	s.devices.removeDevice(name)
}
