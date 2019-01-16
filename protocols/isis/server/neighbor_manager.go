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
	dev  *dev
	db   []*neighbor
	dbMu sync.RWMutex
	p2p  bool
}

func newNeighborManager(d *dev) *neighborManager {
	return &neighborManager{
		dev: d,
		db:  make([]*neighbor, 0, 1),
	}
}

func (nm *neighborManager) setNeighbor(n *neighbor) {
	nm.dbMu.RLock()
	defer nm.dbMu.RUnlock()

	for i := range nm.db {
		if nm.db[i].macAddress != n.macAddress {
			continue
		}

		// TODO: Verfiy if hello is valid for us
		nm.db[i] = n
		return
	}

	// TODO: Verfiy if hello is valid for us
	nm.db = append(nm.db, n)
}

// IS THIS FUNCION NEEDED?
func (nm *neighborManager) getNeighbor(mac types.MACAddress) *neighbor {
	nm.dbMu.RLock()
	defer nm.dbMu.RUnlock()

	for i := range nm.db {
		if nm.db[i].macAddress != mac {
			continue
		}

		return nm.db[i]
	}

	return nil
}
