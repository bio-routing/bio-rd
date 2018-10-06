package server

import(
	"sync"

	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
)

type lsdb struct {
	server *ISISServer
	lsps map[packet.LSPID]*packet.LSPDU
	lspsMu sync.RWMutex
}

func newLSDB(server *ISISServer) *lsdb {
	lsdb := &lsdb{
		server: server,
		lsps: make(map[packet.LSPID]*packet.LSPDU),
	}

	localLSPID := packet.LSPID{
		SystemID: server.config.NETs[0].SystemID,
		PseudonodeID: 0x00,
	}

	localLSPDU := &packet.LSPDU{
		Length: packet.LSPDUMinLen,
		RemainingLifetime: 3600,
		LSPID: localLSPID,
		SequenceNumber: 0,
		Checksum: 0,
		TypeBlock: 0x03,
		TLVs: make([]packet.TLV, 0),
	}

	areas := make([]types.AreaID, len(lsdb.server.config.NETs))
	for _, NET := range lsdb.server.config.NETs {
		areas = append(areas, NET.AreaID)
	}
	areaAddressesTLV := packet.NewAreaAddressTLV(areas)
	localLSPDU.TLVs = append(localLSPDU.TLVs, areaAddressesTLV)

	protoSupportedTLV := packet.NewProtocolsSupportedTLV([]uint8{0xcc, 0x8e})
	localLSPDU.TLVs = append(localLSPDU.TLVs, protoSupportedTLV)

	lsdb.lsps[localLSPID] = localLSPDU
	return lsdb
}

func (lsdb *lsdb) getCSNP() *packet.CSNP {
	lsdb.lspsMu.RLock()
	defer lsdb.lspsMu.RUnlock()

	p := &packet.CSNP{
		PDULength: packet.CSNPMinLen,
		SourceID: lsdb.server.systemID(),
		StartLSPID: 0,
		EndLSPID: ^uint64(0),
		LSPEntries: make([]packet.LSPEntry, len(lsdb.lsps)),
	}

	i := 0
	for _, lsp := range lsdb.lsps {
		lspEntry := packet.LSPEntry{
			SequenceNumber: lsp.SequenceNumber,
			RemainingLifetime: lsp.RemainingLifetime,
			LSPChecksum: lsp.Checksum,
			LSPID: lsp.LSPID,
		}

		p.LSPEntries[i] = lspEntry
		i++
	}

	return p
}