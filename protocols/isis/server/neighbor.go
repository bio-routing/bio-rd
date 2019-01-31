package server

import (
	"fmt"
	"sort"
	"time"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
	biotime "github.com/bio-routing/bio-rd/util/time"
)

type neighbor struct {
	macAddress             types.MACAddress
	systemID               types.SystemID
	dev                    *dev
	holdingTime            uint16
	holdingTimer           biotime.Timer
	localCircuitID         uint8
	extendedLocalCircuitID uint32
	ipInterfaceAddresses   []uint32 // This should always be sorted
	fsm                    *fsm
	done                   chan struct{}
}

func newNeighbor() *neighbor {
	return nil
}

func (n *neighbor) dispose(reason error) {
	n.dev.srv.log.Infof("Disposing adjacency with %s on %s: %v", n.macAddress.String(), n.dev.name, reason)
	n.fsm.dispose()
}

func (n *neighbor) ipAddrsEqual(x []uint32) bool {
	if len(n.ipInterfaceAddresses) != len(x) {
		return false
	}

	sort.Slice(x, func(i, j int) bool { return x[i] < x[j] })
	for i := range n.ipInterfaceAddresses {
		if n.ipInterfaceAddresses[i] != x[i] {
			return false
		}
	}

	return true
}

func (n *neighbor) hello(h *neighbor) (dispose bool) {
	validAddrs := n.dev.validateNeighborAddresses(h.ipInterfaceAddresses)
	if len(validAddrs) == 0 {
		n.dispose(fmt.Errorf("Incompatible IP addresses in hello message"))
		return true
	}

	n.holdingTime = h.holdingTime
	if !n.holdingTimer.Reset(time.Duration(n.holdingTime)) {
		n.dispose(fmt.Errorf("Hold timer expired"))
		return true
	}

	if !n.ipAddrsEqual(validAddrs) || n.localCircuitID != h.localCircuitID || n.extendedLocalCircuitID != h.extendedLocalCircuitID {
		n.dev.srv.lsdb.triggerLSPDUGen()
	}

	return false
}
