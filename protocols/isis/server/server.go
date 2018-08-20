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
		isis.interfaces[ifs.Name] = newNetIf(isis, ifs)
	}

	return nil
}

// Stop stops an ISIS speaker
func (isis *ISISServer) Stop() error {

	return nil
}

func (isis *ISISServer) systemID() [6]byte {
	n := len(isis.config.NetworkEntityTitle)
	return [6]byte{
		isis.config.NetworkEntityTitle[n-7],
		isis.config.NetworkEntityTitle[n-6],
		isis.config.NetworkEntityTitle[n-5],
		isis.config.NetworkEntityTitle[n-4],
		isis.config.NetworkEntityTitle[n-3],
		isis.config.NetworkEntityTitle[n-2],
	}
}
