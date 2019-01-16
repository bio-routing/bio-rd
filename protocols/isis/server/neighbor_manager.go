package server

import (
	"sync"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
)

type neighborEntry struct {
	addr types.MACAddress
	n    *neighbor
}

type neighborManager struct {
	db   []neighborEntry
	dbMu sync.RWMutex
}

func newNeighborManager() *neighborManager {
	return &neighborManager{
		db: make([]neighborEntry, 0, 1),
	}
}

func (nm *neighborManager) setNeighbor(addr types.MACAddress, n *neighbor) {
	nm.dbMu.RLock()
	defer nm.dbMu.RUnlock()

	for i := range nm.db {
		if nm.db[i].addr != addr {
			continue
		}

		// FIXME: Update
		return
	}

	// FIXME: Add
	nm.db = append(nm.db, neighborEntry{
		addr: addr,
		n:    n,
	})
}

func (nm *neighborManager) getNeighbor(addr types.MACAddress) *neighbor {
	nm.dbMu.RLock()
	defer nm.dbMu.RUnlock()

	for i := range nm.db {
		if nm.db[i].addr != addr {
			continue
		}

		return nm.db[i].n
	}

	return nil
}
