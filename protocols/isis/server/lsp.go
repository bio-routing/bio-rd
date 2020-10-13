package server

import (
	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
)

func (s *Server) generateL2LSP(sequenceNumber uint32) {
	lsp := &packet.LSPDU{
		RemainingLifetime: 3600,
		LSPID: packet.LSPID{
			SystemID:     s.nets[0].SystemID,
			PseudonodeID: 0,
			LSPNumber:    0,
		},
		SequenceNumber: sequenceNumber,
		TLVs:           make([]packet.TLV, 0),
	}

	lsp.TypeBlock |= 0x3 // level2, last two bits

	lsp.TLVs = append(lsp.TLVs, s.getAreaAddressTLV())

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
