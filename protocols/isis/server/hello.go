package server

import (
	"fmt"

	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/pkg/errors"
)

func (d *dev) helloRoutine() {
	// To be implemented
}

func (d *dev) processP2PHello(h *packet.P2PHello, src types.MACAddress) error {
	n, err := d.p2pHelloToNeighbor(h)
	if err != nil {
		return errors.Wrap(err, "Unable to create neighbor object from hello")
	}

	d.neighborManager.setNeighbor(src, n)
	return nil
}

func (d *dev) p2pHelloToNeighbor(h *packet.P2PHello) (*neighbor, error) {
	p2pAdjTLV := h.GetP2PAdjTLV()
	if p2pAdjTLV == nil {
		return nil, fmt.Errorf("Received a P2P hello PDU without P2P Adjacency TLV on %s", d.name)
	}

	ipIfAddrTLV := h.GetIPInterfaceAddressesesTLV()
	if ipIfAddrTLV == nil {
		return nil, fmt.Errorf("Received a P2P hello PDU without IP Interface Addresses TLV on %s", d.name)
	}

	n := &neighbor{
		systemID:               h.SystemID,
		dev:                    d,
		holdingTime:            h.HoldingTimer,
		localCircuitID:         h.LocalCircuitID,
		extendedLocalCircuitID: p2pAdjTLV.ExtendedLocalCircuitID,
		ipInterfaceAddresses:   ipIfAddrTLV.IPv4Addresses,
	}

	return n, nil
}
