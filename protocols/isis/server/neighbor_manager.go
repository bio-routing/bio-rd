package server

import (
	"sync"

	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	log "github.com/sirupsen/logrus"
)

type neighborManager struct {
	netIfa      *netIfa
	level       uint8
	neighbors   map[types.SystemID]*neighbor
	neighborsMu sync.Mutex
}

func newNeighborManager(netIfa *netIfa, level uint8) *neighborManager {
	return &neighborManager{
		netIfa:    netIfa,
		level:     level,
		neighbors: make(map[types.SystemID]*neighbor),
	}
}

func (nm *neighborManager) getNeighbors() []*neighbor {
	ret := make([]*neighbor, 0)
	nm.neighborsMu.Lock()
	defer nm.neighborsMu.Unlock()

	for _, v := range nm.neighbors {
		ret = append(ret, v)
	}

	return ret
}

// TODO: Catch if P2P Adj. State is DOWN. What to do then? Drop the neighbor?
func (nm *neighborManager) processP2PHello(hello *packet.P2PHello) error {
	nm.neighborsMu.Lock()
	defer nm.neighborsMu.Unlock()

	// TODO: Validate hello packet (check if all necessary TLVs are there with decent values)

	if _, found := nm.neighbors[hello.SystemID]; !found {
		n := nm.neighborFromP2PHello(hello)
		nm.neighbors[hello.SystemID] = n

		n.wg.Add(1)
		go n.adjChecker()

		log.Infof("IS-IS: NM L%d: Adding new neighbor %q", nm.level, hello.SystemID.String())
		return nil
	}

	n := nm.neighbors[hello.SystemID]
	return n.processP2PHello(hello)
}
