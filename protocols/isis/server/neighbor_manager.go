package server

import (
	"sync"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
)

type neighborManagerInterface interface {
	setNeighbor(n *neighbor)
	getNeighbor(addr types.MACAddress) *neighbor
}

type neighborManager struct {
	srv         *Server
	neighbors   []*neighbor
	neighborsMu sync.RWMutex
}

func newNeighborManager(s *Server) *neighborManager {
	return &neighborManager{
		srv:       s,
		neighbors: make([]*neighbor, 0, 1),
	}
}

func sameNeighbor(a *neighbor, b *neighbor) bool {
	if a.dev != b.dev || a.macAddress != b.macAddress {
		return false
	}

	return true
}

// hello received from neighbor n
func (nm *neighborManager) hello(n *neighbor) {
	e := nm.getNeighbor(n.dev, n.macAddress)
	if e != nil {
		if e.hello(n) {
			nm.removeNeighbor(e)
		}

		return
	}

	nm.newNeighbor(n)
}

func (nm *neighborManager) newNeighbor(n *neighbor) {
	n.fsm = newFSM(nm.srv, n)
	nm.addNeighbor(n)
	n.fsm.start()
}

func (nm *neighborManager) removeNeighbor(n *neighbor) {
	nm.neighborsMu.Lock()
	defer nm.neighborsMu.Unlock()

	for i := range nm.neighbors {
		if nm.neighbors[i] == n {
			nm.neighbors = append(nm.neighbors[:i], nm.neighbors[i+1:]...)
		}
	}
}

func (nm *neighborManager) getNeighbor(d *dev, m types.MACAddress) *neighbor {
	nm.neighborsMu.RLock()
	defer nm.neighborsMu.RUnlock()

	for i := range nm.neighbors {
		if nm.neighbors[i].macAddress != m {
			continue
		}

		return nm.neighbors[i]
	}

	return nil
}

func (nm *neighborManager) addNeighbor(n *neighbor) {
	nm.neighborsMu.Lock()
	defer nm.neighborsMu.Unlock()
	nm.neighbors = append(nm.neighbors, n)
}
