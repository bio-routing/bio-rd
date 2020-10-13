package server

import (
	"fmt"

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
	lsp.TLVs = append(lsp.TLVs, s.getIPInternalReachabilityInformationTLV())
	lsp.TLVs = append(lsp.TLVs, s.getProtocolsSupportedTLV())
	lsp.TLVs = append(lsp.TLVs, s.getIPInterfaceAddressesTLV())

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

func (s *Server) getIPInternalReachabilityInformationTLV() *packet.ISReachabilityTLV {
	neighbors := make([][7]byte, 0)
	for _, ifa := range s.netIfaManager.getAllInterfaces() {
		fmt.Printf("Getting neighbors\n")
		for _, n := range ifa.neighborManagerL2.getNeighbors() {
			neighbors = append(neighbors, [7]byte{
				n.sysID[0],
				n.sysID[1],
				n.sysID[2],
				n.sysID[3],
				n.sysID[4],
				n.sysID[5],
				0,
			})
		}
		fmt.Printf("Getting neighbors: Done\n")
	}

	return packet.NewISReachabilityTLV(neighbors)
}
