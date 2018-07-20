package server

import (
	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
)

const (
	maxEtherFrameSize = 9216
)

//ISISServer represents an ISIS speaker
type ISISServer struct {
	config     config.ISISConfig
	interfaces map[string]*netIf
}

type isisNeighbor struct {
	SystemID  types.SystemID
	Interface string
}

// NewISISServer creates and initializes a new ISIS speaker
func NewISISServer(cfg config.ISISConfig) *ISISServer {
	return &ISISServer{
		config:     cfg,
		interfaces: make(map[string]*netIf),
	}
}

// Start starts an ISIS speaker
func (isis *ISISServer) Start() error {
	for _, ifs := range isis.config.Interfaces {
		isis.interfaces[ifs.Name] = newNetIf(ifs)
	}

	return nil
}

// Stop stops an ISIS speaker
func (isis *ISISServer) Stop() error {

	return nil
}
