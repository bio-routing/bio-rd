package server

import (
	"unsafe"

	"github.com/bio-routing/bio-rd/protocols/device"
	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/types"

	log "github.com/sirupsen/logrus"
)

func (s *Server) regenerateL2LSP() {
	log.Info("Generating L2 LSP")
	defer log.Info("Generating L2 LSP: Done")

	s.sequenceNumberL2++
	lsp := &packet.LSPDU{
		RemainingLifetime: 3600,
		LSPID: packet.LSPID{
			SystemID:     s.nets[0].SystemID,
			PseudonodeID: 0,
			LSPNumber:    0,
		},
		SequenceNumber: s.sequenceNumberL2,
		TLVs:           make([]packet.TLV, 0),
	}

	lsp.TypeBlock |= 0x3 // level2, last two bits

	lsp.TLVs = append(lsp.TLVs, s.getAreaAddressTLV())
	lsp.TLVs = append(lsp.TLVs, s.getISReachabilityTLV())
	lsp.TLVs = append(lsp.TLVs, s.getProtocolsSupportedTLV())
	lsp.TLVs = append(lsp.TLVs, s.getIPInterfaceAddressesTLV())
	lsp.TLVs = append(lsp.TLVs, s.getExtendedISReachabilityTLV())

	lsp.UpdateLength()
	lsp.SetChecksum()

	s.lsdbL2.lspsMu.Lock()
	defer s.lsdbL2.lspsMu.Unlock()

	s.lsdbL2.lsps[lsp.LSPID] = newLSDBEntry(lsp)

	// TODO: Set SRM?

}

func (s *Server) getAreaAddressTLV() *packet.AreaAddressesTLV {
	areas := make([]types.AreaID, 0)
	for _, NET := range s.nets {
		fullAreaID := []byte{NET.AFI}
		areas = append(areas, append(fullAreaID, NET.AreaID...))
	}

	return packet.NewAreaAddressesTLV(areas)
}

func (s *Server) getISNeighborsTLV() *packet.ISNeighborsTLV {
	return nil
}

func (s *Server) getProtocolsSupportedTLV() packet.ProtocolsSupportedTLV {
	return packet.NewProtocolsSupportedTLV([]uint8{
		packet.NLPIDIPv4,
		packet.NLPIDIPv6,
	})
}

func (s *Server) getIPInterfaceAddressesTLV() *packet.IPInterfaceAddressesTLV {
	addrs := make([]uint32, 0)
	for _, ifa := range s.netIfaManager.getAllInterfaces() {
		if ifa.devStatus.GetOperState() != device.IfOperUp {
			continue
		}

		for _, addr := range ifa.devStatus.GetAddrs() {
			if !addr.Addr().IsIPv4() {
				continue
			}

			addrs = append(addrs, addr.Addr().ToUint32())
		}
	}

	return packet.NewIPInterfaceAddressesTLV(addrs)
}

func (s *Server) getISReachabilityTLV() *packet.ISReachabilityTLV {
	neighbors := make([]types.SourceID, 0)
	for _, ifa := range s.netIfaManager.getAllInterfaces() {
		if ifa.devStatus.GetOperState() != device.IfOperUp {
			continue
		}

		for _, n := range ifa.neighborManagerL2.getNeighbors() {
			neighbors = append(neighbors, n.sysID.ToSourceID(0))
		}
	}

	return packet.NewISReachabilityTLV(neighbors)
}

func (s *Server) getExtendedISReachabilityTLV() *packet.ExtendedISReachabilityTLV {
	t := packet.NewExtendedISReachabilityTLV()
	for _, ifa := range s.netIfaManager.getAllInterfaces() {
		for _, n := range ifa.neighborManagerL2.getNeighbors() {
			m := metricToThreeBytes(ifa.cfg.Level2.Metric)
			ntlv := packet.NewExtendedISReachabilityNeighbor(n.sysID.ToSourceID(0), m)

			for _, addr := range ifa.devStatus.GetAddrs() {
				if !addr.Addr().IsIPv4() {
					// TODO: What about IPv6?
					continue
				}

				ipv4LocalTLV := packet.NewIPv4InterfaceAddressSubTLV(addr.Addr().ToUint32())
				ntlv.AddSubTLV(ipv4LocalTLV)
			}

			for _, nAddr := range n.ipAddresses {
				if !nAddr.IsIPv4() {
					// TODO: What about IPv6?
					continue
				}

				ipv4RemoteTLV := packet.NewIPv4NeighborAddressSubTLV(nAddr.ToUint32())
				ntlv.AddSubTLV(ipv4RemoteTLV)
			}

			llriTLV := packet.NewLinkLocalRemoteIdentifiersSubTLV(uint32(ifa.devStatus.GetIndex()), n.extendedLocalCircuitID)
			ntlv.AddSubTLV(llriTLV)

			t.Neighbors = append(t.Neighbors, ntlv)
		}
	}

	t.UpdateLength()
	return t
}

func metricToThreeBytes(m uint32) [3]byte {
	// TODO: Check if this is affected by endian issues
	x := (*[4]byte)(unsafe.Pointer(&m))
	return [3]byte{x[1], x[2], x[3]}
}
