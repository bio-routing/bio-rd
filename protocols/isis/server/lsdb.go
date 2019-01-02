package server

import (
	"fmt"
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
	srmFlags map[*netIf]struct{}
	ssnFlags map[*netIf]struct{}
}

func newLSDBEntry(lspdu *packet.LSPDU) *lsdbEntry {
	return &lsdbEntry{
		lspdu:    lspdu,
		srmFlags: make(map[*netIf]struct{}),
		ssnFlags: make(map[*netIf]struct{}),
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
		SequenceNumber:    1,
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

	localLSPDU.SetChecksum()
	lsdb.lsps[localLSPID] = newLSDBEntry(localLSPDU)
	return lsdb
}

func (lsdb *lsdb) clearSRMSSN(ifa *netIf) {
	lsdb.lspsMu.Lock()

	for lspid, lsp := range lsdb.lsps {
		if _, ok := lsp.srmFlags[ifa]; ok {
			delete(lsdb.lsps[lspid].srmFlags, ifa)
		}
		if _, ok := lsp.ssnFlags[ifa]; ok {
			delete(lsdb.lsps[lspid].ssnFlags, ifa)
		}
	}

	lsdb.lspsMu.Unlock()
}

func (lsdb *lsdb) decrementRemainingLifetimes() {
	lsdb.lspsMu.Lock()

	for lspid, lspdbEntry := range lsdb.lsps {
		if lspdbEntry.lspdu.RemainingLifetime <= 1 {
			delete(lsdb.lsps, lspid)
			continue
		}

		lspdbEntry.lspdu.RemainingLifetime--
	}

	lsdb.lspsMu.Unlock()
}

func (lsdb *lsdb) scanSRMSSN(ifa *netIf) ([]*packet.LSPDU, []*packet.LSPEntry) {
	// We're breaking the rule of single resposibility here for performance reason.
	// Scans for SRM and SSN are synchronized anyways and this way we save an unnecessary
	// Lock() / Unlock() cycle and a map walk.

	lspdus := make([]*packet.LSPDU, 0)
	psnpEntries := make([]*packet.LSPEntry, 0)

	lsdb.lspsMu.Lock()
	for lspid, lsp := range lsdb.lsps {
		if _, ok := lsp.srmFlags[ifa]; ok {
			lspdus = append(lspdus, lsp.lspdu)
			delete(lsdb.lsps[lspid].srmFlags, ifa)
		}
		if _, ok := lsp.ssnFlags[ifa]; ok {
			psnpEntries = append(psnpEntries, &packet.LSPEntry{
				SequenceNumber:    lsp.lspdu.SequenceNumber,
				RemainingLifetime: lsp.lspdu.RemainingLifetime,
				LSPChecksum:       lsp.lspdu.Checksum,
				LSPID:             lsp.lspdu.LSPID,
			})
			delete(lsdb.lsps[lspid].ssnFlags, ifa)
		}
	}

	lsdb.lspsMu.Unlock()
	return lspdus, psnpEntries
}

func lspEntryToLSPDU(lspEntry *packet.LSPEntry) *packet.LSPDU {
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
			return
		}

		// Same LSP
		lsdb.lsps[lspdu.LSPID].clearSRM(ifa)
		lsdb.lsps[lspdu.LSPID].setSSN(ifa)
		return
	}

	// New LSP
	lsdb.lsps[lspdu.LSPID] = newLSDBEntry(lspdu)
	lsdb.setSRMAnyIf(lspdu.LSPID)
	lsdb.lsps[lspdu.LSPID].clearSRM(ifa)
	lsdb.lsps[lspdu.LSPID].setSSN(ifa)
}

func (lsdb *lsdb) processPSNP(ifa *netIf, psnp *packet.PSNP) {
	lsdb.lspsMu.Lock()

	for _, lspEntry := range psnp.LSPEntries {
		lsdb.lsps[lspEntry.LSPID].clearSRM(ifa)
	}

	lsdb.lspsMu.Unlock()
}

func (lsdb *lsdb) processCSNP(ifa *netIf, csnp *packet.CSNP) {
	lsdb.lspsMu.Lock()

	seenLSPIDs := make(map[packet.LSPID]struct{})

	for _, lspEntry := range csnp.GetLSPEntries() {
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
		if lspEntry.RemainingLifetime == 0 {
			continue
		}
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

	lsdb.lspsMu.Unlock()
}

func (lsdb *lsdb) getCSNP() *packet.CSNP {
	lsdb.lspsMu.RLock()

	p := &packet.CSNP{
		PDULength:  packet.CSNPMinLen,
		SourceID:   types.NewSourceID(lsdb.server.systemID(), 0),
		StartLSPID: packet.LSPID{},
		EndLSPID: packet.LSPID{
			SystemID:     types.SystemID{255, 255, 255, 255, 255, 255},
			PseudonodeID: 65535,
		},
		TLVs: make([]packet.TLV, 0, 1),
	}

	lspEntries := make([]*packet.LSPEntry, 0)
	i := 0
	for _, lspEntry := range lsdb.lsps {
		lspEntry := &packet.LSPEntry{
			SequenceNumber:    lspEntry.lspdu.SequenceNumber,
			RemainingLifetime: lspEntry.lspdu.RemainingLifetime,
			LSPChecksum:       lspEntry.lspdu.Checksum,
			LSPID:             lspEntry.lspdu.LSPID,
		}

		lspEntries = append(lspEntries, lspEntry)
		i++
	}

	p.TLVs = append(p.TLVs, packet.NewLSPEntriesTLV(lspEntries))
	lsdb.lspsMu.RUnlock()
	return p
}

func (lsdb *lsdb) setSRMAny(ifa *netIf) {
	lsdb.lspsMu.Lock()

	for _, lsdbEntry := range lsdb.lsps {
		if lsdbEntry.lspdu.SequenceNumber == 0 {
			continue
		}
		lsdbEntry.setSRM(ifa)
	}

	lsdb.lspsMu.Unlock()
}

func (lsdb *lsdb) setSRMAnyIf(lspid packet.LSPID) {
	for _, ifa := range lsdb.server.interfaces {
		lsdb.lsps[lspid].setSRM(ifa)
	}
}

func (e *lsdbEntry) setSRM(ifa *netIf) {
	fmt.Printf("####### SETTING SRM flag for %v on %v\n", e.lspdu.LSPID.String(), ifa.name)
	e.srmFlags[ifa] = struct{}{}
}

func (e *lsdbEntry) clearSRM(ifa *netIf) {
	delete(e.srmFlags, ifa)
}

func (e *lsdbEntry) setSSN(ifa *netIf) {
	e.ssnFlags[ifa] = struct{}{}
}

func (e *lsdbEntry) clearSSN(ifa *netIf) {
	delete(e.ssnFlags, ifa)
}
