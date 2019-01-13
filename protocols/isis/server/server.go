package server

import (
	"github.com/bio-routing/bio-rd/protocols/device"
	"github.com/bio-routing/isis/config"
)

//Server represents an ISIS server
type Server struct {
	config         *config.ISISConfig
	sequenceNumber uint32
	devices        *devices
	lsdb           *lsdb
	stop           chan struct{}
	ds             *device.Server
}

func New(cfg *config.ISISConfig, ds *device.Server) *Server {
	s := &Server{
		config:         cfg,
		ds:             ds,
		sequenceNumber: 1,
		devices:        newDevices(),
		stop:           make(chan struct{}),
	}

	s.lsdb = newLSDB(s)
	return s
}

func (s *Server) dispose() {
	s.lsdb.dispose()
	s.lsdb = nil
}
