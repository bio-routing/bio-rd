package server

import (
	"bytes"
	"fmt"
	"time"

	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	btime "github.com/bio-routing/bio-rd/util/time"
	"github.com/pkg/errors"
)

func (d *dev) helloRoutine() {
	defer d.wg.Done()

	t := btime.NewBIOTicker(time.Second * 27)
	for {
		select {
		case <-d.done:
			return
		case <-t.C():
			err := d.sendP2PHello()
			if err != nil {
				d.srv.log.Errorf("Unable to send hello packet: %v", err)
				// TODO: Should we trigger an adjacency down here?
				return
			}
		}
	}
}

func (d *dev) sendP2PHello() error {
	p := packet.P2PHello{
		CircuitType:    packet.L2CircuitType,
		SystemID:       d.srv.systemID(),
		HoldingTimer:   d.level2.HoldTime,
		PDULength:      packet.P2PHelloMinLen,
		LocalCircuitID: uint8(d.phy.Index),
		TLVs:           d.p2pHelloTLVs(),
	}

	for _, TLV := range p.TLVs {
		p.PDULength += 2
		p.PDULength += uint16(TLV.Length())
	}

	helloBuf := bytes.NewBuffer(nil)
	p.Serialize(helloBuf)

	hdr := packet.ISISHeader{
		ProtoDiscriminator:  0x83,
		LengthIndicator:     packet.P2PHelloMinLen,
		ProtocolIDExtension: 1,
		IDLength:            0,
		PDUType:             packet.P2P_HELLO,
		Version:             1,
		MaxAreaAddresses:    0,
	}

	hdrBuf := bytes.NewBuffer(nil)
	hdr.Serialize(hdrBuf)
	hdrBuf.Write(helloBuf.Bytes())

	err := d.sys.sendPacket(hdrBuf.Bytes(), packet.AllISS)
	if err != nil {
		return fmt.Errorf("failed to send packet: %v", err)
	}

	return nil
}

func (d *dev) p2pHelloTLVs() []packet.TLV {
	/*l2AdjacencyState, neighborSystemID, neighborExtendedLocalCircuitID := d.p2pL2AdjacencyState()
	p2pAdjStateTLV := packet.NewP2PAdjacencyStateTLV(l2AdjacencyState, uint32(d.phy.Index))

	switch l2AdjacencyState {
	case packet.INITIALIZING_STATE:
		p2pAdjStateTLV.TLVLength = packet.P2PAdjacencyStateTLVLenWithNeighbor
		p2pAdjStateTLV.NeighborSystemID = neighborSystemID
		p2pAdjStateTLV.NeighborExtendedLocalCircuitID = neighborExtendedLocalCircuitID
	}

	protocolsSupportedTLV := packet.NewProtocolsSupportedTLV(d.supportedProtocols)
	areaAddressesTLV := packet.NewAreaAddressesTLV(d.srv.getAreas())

	ipInterfaceAddressesTLV := packet.NewIPInterfaceAddressesTLV([]uint32{3232235523}) //FIXME: Insert address automatically

	return []packet.TLV{
		p2pAdjStateTLV,
		protocolsSupportedTLV,
		areaAddressesTLV,
		ipInterfaceAddressesTLV,
	}*/
	return nil
}

func (d *dev) processP2PHello(h *packet.P2PHello, src types.MACAddress) error {
	if h.CircuitType != 2 {
		return fmt.Errorf("Unsupported P2P Hello: Circuit Type: %d", h.CircuitType)
	}

	n, err := d.newNeighbor(h, src)
	if err != nil {
		return errors.Wrap(err, "Unable to create neighbor object from hello")
	}

	d.srv.nm.hello(n)
	return nil
}

func (d *dev) newNeighbor(h *packet.P2PHello, src types.MACAddress) (*neighbor, error) {
	p2pAdjTLV := h.GetP2PAdjTLV()
	if p2pAdjTLV == nil {
		return nil, fmt.Errorf("Received a P2P hello PDU without P2P Adjacency TLV on %s", d.name)
	}

	ipIfAddrTLV := h.GetIPInterfaceAddressesesTLV()
	if ipIfAddrTLV == nil {
		return nil, fmt.Errorf("Received a P2P hello PDU without IP Interface Addresses TLV on %s", d.name)
	}

	n := &neighbor{
		dev:                    d,
		macAddress:             src,
		systemID:               h.SystemID,
		holdingTime:            h.HoldingTimer,
		localCircuitID:         h.LocalCircuitID,
		extendedLocalCircuitID: p2pAdjTLV.ExtendedLocalCircuitID,
		ipInterfaceAddresses:   ipIfAddrTLV.IPv4Addresses,
		done:                   make(chan struct{}),
	}

	return n, nil
}
