package server

import (
	"bytes"

	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/bio-routing/tflow2/convert"
)

func (nifa *netIfa) p2pHelloSender() {
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

			_, err := nifa.isP2PHelloCon.Write(hdrBuf.Bytes())
			if err != nil {
				panic(err)
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

	h.TLVs = append(h.TLVs, packet.NewP2PAdjacencyStateTLV(nifa.p2pAdjState, uint32(nifa.devStatus.GetIndex())))
	h.TLVs = append(h.TLVs, packet.NewProtocolsSupportedTLV([]uint8{
		packet.NLPIDIPv4,
		packet.NLPIDIPv6,
	}))

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
