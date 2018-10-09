package server

import (
	"sync"

	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
)

type lsdb struct {
	server *ISISServer
	lsps   map[packet.LSPID]*lsdbEntry
	lspsMu sync.RWMutex
}

type lsdbEntry struct {
	lspdu    *packet.LSPDU
	srmFlags map[string]struct{}
	ssnFlags map[string]struct{}
}

func newLSDBEntry(lspdu *packet.LSPDU) *lsdbEntry {
	return &lsdbEntry{
		lspdu:    lspdu,
		srmFlags: make(map[string]struct{}),
		ssnFlags: make(map[string]struct{}),
	}
}

func newLSDB(server *ISISServer) *lsdb {
	lsdb := &lsdb{
		server: server,
		lsps:   make(map[packet.LSPID]*lsdbEntry),
	}

	localLSPID := packet.LSPID{
		SystemID:     server.config.NETs[0].SystemID,
		PseudonodeID: 0x00,
	}

	localLSPDU := &packet.LSPDU{
		Length:            packet.LSPDUMinLen,
		RemainingLifetime: 3600,
		LSPID:             localLSPID,
		SequenceNumber:    0,
		Checksum:          0,
		TypeBlock:         0x03,
		TLVs:              make([]packet.TLV, 0),
	}

	areas := make([]types.AreaID, len(lsdb.server.config.NETs))
	for _, NET := range lsdb.server.config.NETs {
		areas = append(areas, NET.AreaID)
	}
	areaAddressesTLV := packet.NewAreaAddressesTLV(areas)
	localLSPDU.TLVs = append(localLSPDU.TLVs, areaAddressesTLV)

	protoSupportedTLV := packet.NewProtocolsSupportedTLV([]uint8{0xcc, 0x8e})
	localLSPDU.TLVs = append(localLSPDU.TLVs, protoSupportedTLV)

	lsdb.lsps[localLSPID] = newLSDBEntry(localLSPDU)
	return lsdb
}

func lspEntryToLSPDU(lspEntry packet.LSPEntry) *packet.LSPDU {
	return &packet.LSPDU{
		SequenceNumber:    lspEntry.SequenceNumber,
		RemainingLifetime: lspEntry.RemainingLifetime,
		Checksum:          lspEntry.LSPChecksum,
		LSPID:             lspEntry.LSPID,
	}
}

func (lsdb *lsdb) processLSPDU(ifa *netIf, lspdu *packet.LSPDU) {
	lsdb.lspsMu.Lock()
	defer lsdb.lspsMu.Unlock()

	if _, ok := lsdb.lsps[lspdu.LSPID]; ok {
		if lspdu.SequenceNumber > lsdb.lsps[lspdu.LSPID].lspdu.SequenceNumber {
			// Recevied LSP is newer
			lsdb.lsps[lspdu.LSPID].lspdu = lspdu
			lsdb.setSRMAnyIf(lspdu.LSPID)
			lsdb.lsps[lspdu.LSPID].clearSRM(ifa)
			lsdb.lsps[lspdu.LSPID].setSSN(ifa)
			return
		}

		if lsdb.lsps[lspdu.LSPID].lspdu.SequenceNumber > lspdu.SequenceNumber {
			// Received older LSP
			lsdb.lsps[lspdu.LSPID].setSRM(ifa)
			lsdb.lsps[lspdu.LSPID].clearSSN(ifa)
		}

		// Same LSP
		lsdb.lsps[lspdu.LSPID].clearSRM(ifa)
		lsdb.lsps[lspdu.LSPID].setSSN(ifa)
	}

	// New LSP
	lsdb.lsps[lspdu.LSPID].lspdu = lspdu
	lsdb.setSRMAnyIf(lspdu.LSPID)
	lsdb.lsps[lspdu.LSPID].clearSRM(ifa)
	lsdb.lsps[lspdu.LSPID].setSSN(ifa)
}

func (lsdb *lsdb) processPSNP(ifa *netIf, psnp *packet.PSNP) {
	lsdb.lspsMu.Lock()
	defer lsdb.lspsMu.Unlock()

	for _, lspEntry := range psnp.LSPEntries {
		lsdb.lsps[lspEntry.LSPID].clearSRM(ifa)
	}
}

func (lsdb *lsdb) processCSNP(ifa *netIf, csnp *packet.CSNP) {
	lsdb.lspsMu.Lock()
	defer lsdb.lspsMu.Unlock()

	seenLSPIDs := make(map[packet.LSPID]struct{})

	for _, lspEntry := range csnp.LSPEntries {
		seenLSPIDs[lspEntry.LSPID] = struct{}{}

		if _, ok := lsdb.lsps[lspEntry.LSPID]; ok {
			// Same LSP in database
			if lsdb.lsps[lspEntry.LSPID].lspdu.SequenceNumber == lspEntry.SequenceNumber {
				lsdb.lsps[lspEntry.LSPID].clearSRM(ifa)
				continue
			}

			// Newer LSP in database
			if lsdb.lsps[lspEntry.LSPID].lspdu.SequenceNumber > lspEntry.SequenceNumber {
				lsdb.lsps[lspEntry.LSPID].clearSSN(ifa)
				lsdb.lsps[lspEntry.LSPID].setSRM(ifa)
				continue
			}

			// Older LSP in database
			lsdb.lsps[lspEntry.LSPID].clearSRM(ifa)
			lsdb.lsps[lspEntry.LSPID].setSRM(ifa)
			continue
		}

		// LSP not in database
		newLSPEntry := newLSDBEntry(lspEntryToLSPDU(lspEntry))
		newLSPEntry.lspdu.SequenceNumber = 0
		lsdb.lsps[newLSPEntry.lspdu.LSPID] = newLSPEntry
	}

	// Check for LSPs in the database that are not in the CSNP
	for _, lspEntry := range lsdb.lsps {
		if _, ok := seenLSPIDs[lspEntry.lspdu.LSPID]; ok {
			// LSPID found
			continue
		}

		if lspEntry.lspdu.SequenceNumber == 0 || lspEntry.lspdu.RemainingLifetime == 0 {
			// Invalid Sequence Number of Remaining Lifetime
			continue
		}

		if csnp.StartLSPID.Compare(lspEntry.lspdu.LSPID) == -1 {
			// LSPID below Start
			continue
		}

		if csnp.EndLSPID.Compare(lspEntry.lspdu.LSPID) == 1 {
			// LSPID above End
			continue
		}

		lsdb.lsps[lspEntry.lspdu.LSPID].setSRM(ifa)
	}
}

func (lsdb *lsdb) getCSNP() *packet.CSNP {
	lsdb.lspsMu.RLock()
	defer lsdb.lspsMu.RUnlock()

	p := &packet.CSNP{
		PDULength:  packet.CSNPMinLen,
		SourceID:   lsdb.server.systemID(),
		StartLSPID: packet.LSPID{},
		EndLSPID: packet.LSPID{
			SystemID:     types.SystemID{255, 255, 255, 255, 255, 255},
			PseudonodeID: 65535,
		},
		LSPEntries: make([]packet.LSPEntry, len(lsdb.lsps)),
	}

	i := 0
	for _, lspEntry := range lsdb.lsps {
		lspEntry := packet.LSPEntry{
			SequenceNumber:    lspEntry.lspdu.SequenceNumber,
			RemainingLifetime: lspEntry.lspdu.RemainingLifetime,
			LSPChecksum:       lspEntry.lspdu.Checksum,
			LSPID:             lspEntry.lspdu.LSPID,
		}

		p.LSPEntries[i] = lspEntry
		i++
	}

	return p
}

func (lsdb *lsdb) setSRMAny(ifa *netIf) {
	lsdb.lspsMu.Lock()
	defer lsdb.lspsMu.Unlock()

	for _, lsdbEntry := range lsdb.lsps {
		if lsdbEntry.lspdu.SequenceNumber == 0 {
			continue
		}
		lsdbEntry.setSRM(ifa)
	}
}

func (lsdb *lsdb) setSRMAnyIf(lspid packet.LSPID) {
	for _, ifa := range lsdb.server.interfaces {
		lsdb.lsps[lspid].setSRM(ifa)
	}
}

func (e *lsdbEntry) setSRM(ifa *netIf) {
	e.srmFlags[ifa.name] = struct{}{}
}

func (e *lsdbEntry) clearSRM(ifa *netIf) {
	delete(e.srmFlags, ifa.name)
}

func (e *lsdbEntry) setSSN(ifa *netIf) {
	e.ssnFlags[ifa.name] = struct{}{}
}

func (e *lsdbEntry) clearSSN(ifa *netIf) {
	delete(e.ssnFlags, ifa.name)
}
