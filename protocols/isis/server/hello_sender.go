package server

import (
	"bytes"

	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/bio-routing/tflow2/convert"
	log "github.com/sirupsen/logrus"
)

func (nifa *netIfa) p2pHelloSender() {
	log.WithFields(nifa.fields()).Debug("Starting hello sender")

	for {
		select {
		case <-nifa.done:
			nifa.helloTicker.Stop()
			nifa.wg.Done()
			return
		case <-nifa.helloTicker.C():
			hello := nifa.p2pHello()
			helloBuf := bytes.NewBuffer(nil)
			hello.Serialize(helloBuf)

			hdr := getHeader(packet.P2P_HELLO)
			hdrBuf := bytes.NewBuffer(nil)
			hdr.Serialize(hdrBuf)
			hdrBuf.Write(helloBuf.Bytes())

			_, err := nifa.isP2PHelloCon.Write(hdrBuf.Bytes())
			if err != nil {
				panic(err) // TODO takt: Check if we really really want this?
			}
		}

	}
}

func (nifa *netIfa) p2pHello() *packet.P2PHello {
	circuitType := uint8(0)
	if nifa.cfg.Level1 != nil {
		circuitType++
	}
	if nifa.cfg.Level2 != nil {
		circuitType += 2
	}

	h := &packet.P2PHello{
		CircuitType:    circuitType,
		SystemID:       nifa.srv.nets[0].SystemID,
		HoldingTimer:   nifa.cfg.holdingTimer(),
		PDULength:      packet.P2PHelloMinLen,
		LocalCircuitID: 1,
		TLVs:           make([]packet.TLV, 0, 5),
	}

	n := nifa.getP2PNeighbor()
	if n == nil {
		h.TLVs = append(h.TLVs, packet.NewP2PAdjacencyStateTLV(packet.DOWN_STATE, uint32(nifa.devStatus.GetIndex())))
	} else {
		p2pAdjTLV := packet.NewP2PAdjacencyStateTLV(n.getState(), uint32(nifa.devStatus.GetIndex()))
		p2pAdjTLV.NeighborSystemID = n.sysID
		p2pAdjTLV.NeighborExtendedLocalCircuitID = n.extendedLocalCircuitID
		p2pAdjTLV.TLVLength = packet.P2PAdjacencyStateTLVLenWithNeighbor
		h.TLVs = append(h.TLVs, p2pAdjTLV)
	}

	h.TLVs = append(h.TLVs, nifa.srv.getProtocolsSupportedTLV())

	ipv4Addrs := make([]uint32, 0)
	for _, a := range nifa.devStatus.GetAddrs() {
		if !a.Addr().IsIPv4() {
			continue
		}

		ipv4Addrs = append(ipv4Addrs, convert.Uint32(convert.Reverse(a.Addr().Bytes())))
	}
	h.TLVs = append(h.TLVs, packet.NewIPInterfaceAddressesTLV(ipv4Addrs))

	// TODO: Add IPv6 Interface Addresses TLV

	areas := make([]types.AreaID, 0)
	for _, net := range nifa.srv.nets {
		areas = append(areas, append([]byte{net.AFI}, net.AreaID...))
	}
	h.TLVs = append(h.TLVs, packet.NewAreaAddressesTLV(areas))

	return h
}

func (nifa *netIfa) getP2PNeighbor() *neighbor {
	var l1n *neighbor
	if nifa.neighborManagerL1 != nil {
		l1Neighbors := nifa.neighborManagerL1.getNeighbors()
		if len(l1Neighbors) > 1 {
			log.WithFields(nifa.fields()).Errorf("IS-IS: p2p interface with more than one L1 neighbor")
		}

		if len(l1Neighbors) == 1 {
			l1n = l1Neighbors[0]
		}
	}

	var l2n *neighbor
	if nifa.neighborManagerL2 != nil {
		l2Neighbors := nifa.neighborManagerL2.getNeighbors()
		if len(l2Neighbors) > 1 {
			log.WithFields(nifa.fields()).Errorf("IS-IS: p2p interface with more than one L2 neighbor")
		}

		if len(l2Neighbors) == 1 {
			l2n = l2Neighbors[0]
		}
	}

	if l1n != nil && l2n != nil {
		if l1n.sysID != l2n.sysID {
			log.WithFields(nifa.fields()).Errorf("BUG: Seeing different system IDs for L1 and L2 on a p2p interface")
		}
	}

	if l1n != nil {
		return l1n
	}

	if l2n != nil {
		return l2n
	}

	return nil
}
