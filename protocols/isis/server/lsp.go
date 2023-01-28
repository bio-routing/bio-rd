package server

import (
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

// TODO: Call this function
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
			// TODO: Add ExtendedISReachabilityTLV
			// TODO: Add ExtendedIPReachabilityTLV
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
