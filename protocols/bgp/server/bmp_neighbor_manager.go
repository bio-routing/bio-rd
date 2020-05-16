package server

import (
	"fmt"
	"sync"
)

type neighborManager struct {
	neighbors   []*Neighbor
	neighborsMu sync.Mutex
}

func newNeighborManager() *neighborManager {
	return &neighborManager{
		neighbors: make([]*Neighbor, 0),
	}
}

func (nm *neighborManager) addNeighbor(n *Neighbor) error {
	nm.neighborsMu.Lock()
	defer nm.neighborsMu.Unlock()

	for i := range nm.neighbors {
		if nm.neighbors[i].vrfID == n.vrfID && nm.neighbors[i].peerAddress == n.peerAddress {
			return fmt.Errorf("Unable to add neighbor %s on VRF %d: exists", n.peerAddress, n.vrfID)
		}
	}

	nm.neighbors = append(nm.neighbors, n)
	return nil
}

func (nm *neighborManager) getNeighbor(vrfID uint64, addr [16]byte) *Neighbor {
	nm.neighborsMu.Lock()
	defer nm.neighborsMu.Unlock()

	for i := range nm.neighbors {
		if nm.neighbors[i].vrfID == vrfID && nm.neighbors[i].peerAddress == addr {
			return nm.neighbors[i]
		}
	}

	return nil
}

func (nm *neighborManager) neighborDown(vrfID uint64, addr [16]byte) error {
	nm.neighborsMu.Lock()
	defer nm.neighborsMu.Unlock()

	return nm._neighborDown(vrfID, addr)
}

func (nm *neighborManager) _neighborDown(vrfID uint64, addr [16]byte) error {
	for i := range nm.neighbors {
		if nm.neighbors[i].vrfID != vrfID || nm.neighbors[i].peerAddress != addr {
			continue
		}

		if nm.neighbors[i].fsm.ipv4Unicast != nil {
			nm.neighbors[i].fsm.ipv4Unicast.bmpDispose()
		}

		if nm.neighbors[i].fsm.ipv6Unicast != nil {
			nm.neighbors[i].fsm.ipv6Unicast.bmpDispose()
		}

		nm.neighbors = append(nm.neighbors[:i], nm.neighbors[i+1:]...)
		return nil
	}

	return fmt.Errorf("Neighbor %d/%v not found", vrfID, addr)
}

func (nm *neighborManager) disposeAll() {
	nm.neighborsMu.Lock()
	defer nm.neighborsMu.Unlock()

	for len(nm.neighbors) > 0 {
		nm._neighborDown(nm.neighbors[0].vrfID, nm.neighbors[0].peerAddress)
	}
}

func (nm *neighborManager) list() []*Neighbor {
	nm.neighborsMu.Lock()
	defer nm.neighborsMu.Unlock()

	ret := make([]*Neighbor, len(nm.neighbors))
	for i := range nm.neighbors {
		ret[i] = nm.neighbors[i]
	}

	return ret
}
