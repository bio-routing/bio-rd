package server

import (
	"fmt"
	
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
		i, err := newNetIf(isis, ifs)
		if err != nil {
			return fmt.Errorf("Unable to enable ISIS on %s: %v", ifs.Name, err)
		}
		isis.interfaces[ifs.Name] = i

		go func(ifaName string) {
			for {
				isis.interfaces[ifaName].readPacket()
			}
		}(ifs.Name)
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
