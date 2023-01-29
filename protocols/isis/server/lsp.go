package server

import (
	"github.com/bio-routing/bio-rd/protocols/device"
	"github.com/bio-routing/bio-rd/protocols/isis/packet"
)

const defaultLifetimeSeconds = 1800

func (s *Server) getProtocolsSupportedTLV() packet.ProtocolsSupportedTLV {
	return packet.NewProtocolsSupportedTLV([]uint8{
		packet.NLPIDIPv4,
		packet.NLPIDIPv6,
	})
}

func (s *Server) nextL2SequencenNumber() uint32 {
	s.sequenceNumberL2Mu.Lock()
	defer s.sequenceNumberL2Mu.Unlock()

	s.sequenceNumberL2++
	if s.sequenceNumberL2 == 0 {
		s.sequenceNumberL2++
	}

	return s.sequenceNumberL2
}

func (s *Server) generateLocalLSP() *packet.LSPDU {
	l := &packet.LSPDU{
		RemainingLifetime: defaultLifetimeSeconds,
		LSPID: packet.LSPID{
			SystemID:     s.systemID(),
			PseudonodeID: 0,
			LSPNumber:    0, // FIXME: We may need to use multiple LSPDUs
		},
		SequenceNumber: s.nextL2SequencenNumber(),
		TLVs: []packet.TLV{
			packet.NewAreaAddressesTLV(s.areaIDs()),
			s.getProtocolsSupportedTLV(),
			packet.NewIPInterfaceAddressesTLV(s.netIfaManager.getAddressesIPv4()),
			s.extendedIPReachabilityTLV(),
			s.extendedISReachabilityTLV(),
		},
	}

	hostname, err := s.hostname()
	if err == nil {
		l.TLVs = append(l.TLVs, packet.NewDynamicHostnameTLV([]byte(hostname)))
	}

	l.UpdateLength()
	l.SetChecksum()

	return l
}

func (s *Server) extendedIPReachabilityTLV() *packet.ExtendedIPReachabilityTLV {
	eipr := packet.NewExtendedIPReachabilityTLV()
	for _, ifa := range s.netIfaManager.getAllInterfaces() {
		if ifa.devStatus.GetOperState() != device.IfOperUp {
			continue
		}

		for _, addr := range ifa.ipv4Addrs() {
			eipr.AddExtendedIPReachability(
				packet.NewExtendedIPReachability(
					ifa.cfg.Level2.Metric,
					addr.Len(),
					addr.Addr().ToUint32()),
			)
		}
	}

	return eipr
}

func (s *Server) extendedISReachabilityTLV() *packet.ExtendedISReachabilityTLV {
	eir := packet.NewExtendedISReachabilityTLV()
	for _, ifa := range s.netIfaManager.getAllInterfaces() {
		for _, n := range ifa.neighborManagerL2.getNeighborsUp() {
			eir.AddNeighbor(n.extendedISReachabilityNeighbor())
		}
	}

	return eir
}
