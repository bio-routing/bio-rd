package server

import (
	"os"

	"github.com/bio-routing/bio-rd/protocols/isis/packet"
)

func (s *ISISServer) createLSPDU() *packet.LSPDU {
	lspID := packet.LSPID{
		SystemID:     s.config.NETs[0].SystemID,
		PseudonodeID: 0x00,
	}

	lspdu := &packet.LSPDU{
		Length:            packet.LSPDUMinLen,
		RemainingLifetime: 3600,
		LSPID:             lspID,
		SequenceNumber:    1,
		Checksum:          0,
		TypeBlock:         0x03,
		TLVs:              make([]packet.TLV, 0),
	}

	areaAddressesTLV := packet.NewAreaAddressesTLV(s.getAreas())
	lspdu.TLVs = append(lspdu.TLVs, areaAddressesTLV)

	protoSupportedTLV := packet.NewProtocolsSupportedTLV([]uint8{0xcc, 0x8e})
	lspdu.TLVs = append(lspdu.TLVs, protoSupportedTLV)

	hostname, err := os.Hostname()
	if err == nil {
		hostnameTLV := packet.NewDynamicHostnameTLV([]byte(hostname))
		lspdu.TLVs = append(lspdu.TLVs, hostnameTLV)
	}

	ipInterfaceAddrTLV := packet.NewIPInterfaceAddressesTLV([]uint32{3232235520})
	lspdu.TLVs = append(lspdu.TLVs, ipInterfaceAddrTLV)

	teRouterIDTLV := packet.NewTrafficEngineeringRouterIDTLV(s.config.TrafficEngineeringRouterID)
	lspdu.TLVs = append(lspdu.TLVs, teRouterIDTLV)

	extISReachTLV := s.createExtendedISReachabilityTLV()
	lspdu.TLVs = append(lspdu.TLVs, extISReachTLV)

	lspdu.SetChecksum()
	return lspdu
}

func (s *ISISServer) createExtendedISReachabilityTLV() *packet.ExtendedISReachabilityTLV {
	tlv := packet.NewExtendedISReachabilityTLV()

	s.interfacesMu.RLock()
	defer s.interfacesMu.RUnlock()
	for i := range s.interfaces {
		for j := range s.interfaces[i].l2.neighbors {
			n := s.interfaces[i].l2.neighbors[j].getExtendedISReachabilityNeighbor()
			tlv.Neighbors = append(tlv.Neighbors, n)
		}
	}

	return tlv
}
