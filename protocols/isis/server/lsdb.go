package server

import (
	"sync"

	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	btime "github.com/bio-routing/bio-rd/util/time"

	log "github.com/sirupsen/logrus"
)

type lsdb struct {
	srv    *Server
	lsps   map[packet.LSPID]*lsdbEntry
	lspsMu sync.RWMutex
	done   chan struct{}
	wg     sync.WaitGroup
}

func newLSDB(s *Server) *lsdb {
	return &lsdb{
		srv:  s,
		done: make(chan struct{}),
	}
}

func (l *lsdb) fields() log.Fields {
	level := 0
	if l.srv.lsdbL1 == l {
		level = 1
	}
	if l.srv.lsdbL2 == l {
		level = 2
	}

	return log.Fields{
		"level": level,
	}
}

func (l *lsdb) dispose() {
	l.stop()
}

func (l *lsdb) start(decrementTicker btime.Ticker, minLSPTransmissionTicker btime.Ticker) {
	l.wg.Add(1)
	go l.decrementRemainingLifetimesRoutine(decrementTicker)

	l.wg.Add(1)
	go l.sendLSPDUsRoutine(minLSPTransmissionTicker)
}

func (l *lsdb) stop() {
	close(l.done)
	l.wg.Wait()
}

func (l *lsdb) decrementRemainingLifetimesRoutine(t btime.Ticker) {
	defer l.wg.Done()

	for {
		select {
		case <-t.C():
			l.decrementRemainingLifetimes()
		case <-l.done:
			return
		}
	}
}

func (l *lsdb) decrementRemainingLifetimes() {
	l.lspsMu.Lock()
	defer l.lspsMu.Unlock()

	for lspid, lspdbEntry := range l.lsps {
		if lspdbEntry.lspdu.RemainingLifetime <= 1 {
			delete(l.lsps, lspid)
			continue
		}

		lspdbEntry.lspdu.RemainingLifetime--
	}
}

func (l *lsdb) setSRMAllLSPs(ifa *netIfa) {
	log.WithFields(l.fields()).Debugf("Setting SRM flags for interface %s", ifa.name)

	for _, lsp := range l.lsps {
		lsp.setSRM(ifa)
	}
}

func (l *lsdb) sendLSPDUsRoutine(t btime.Ticker) {
	defer l.wg.Done()

	for {
		select {
		case <-t.C():
			l.sendLSPDUs()
		case <-l.done:
			return
		}
	}
}

func (l *lsdb) sendLSPDUs() {
	l.lspsMu.RLock()
	defer l.lspsMu.RUnlock()

	for _, entry := range l.lsps {
		for _, ifa := range entry.getInterfacesSRMSet() {
			ifa.sendLSPDU(entry.lspdu)
		}
	}
}

func (l *lsdb) processCSNP(csnp *packet.CSNP, from *netIfa) {
	l.lspsMu.Lock()
	defer l.lspsMu.Unlock()

	for _, lspEntry := range csnp.GetLSPEntries() {
		l.processCSNPLSPEntry(lspEntry, from)
	}

	for lspID, lsdbEntry := range l.lsps {
		// we need to check if we have LSPs the neighbor did not describe.
		// For any that we have but our neighbor doesn't we set SRM flag so
		// the entry gets propagated.

		if lsdbEntry.lspdu.RemainingLifetime <= 0 || lsdbEntry.lspdu.SequenceNumber <= 0 {
			continue
		}

		if !csnp.RangeContainsLSPID(lspID) {
			continue
		}

		if !csnp.ContainsLSPEntry(lspID) {
			lsdbEntry.setSRM(from)
		}
	}
}

func (l *lsdb) processCSNPLSPEntry(lspEntry *packet.LSPEntry, from *netIfa) {
	e := l._getLSPDU(lspEntry.LSPID)
	if e == nil {
		l.processCSNPLSPEntryUnknown(lspEntry, from)
		return
	}

	if e.sameAsInLSPEntry(lspEntry) {
		e.clearSRMFlag(from)
		return
	}

	if e.newerInDatabase(lspEntry) {
		e.clearSSNFlag(from)
		e.setSRM(from)
		return
	}

	if e.olderInDatabase(lspEntry) {
		e.clearSRMFlag(from)
		e.setSSN(from)
		return
	}
}

func (l *lsdb) processCSNPLSPEntryUnknown(lspEntry *packet.LSPEntry, from *netIfa) {
	l.lsps[lspEntry.LSPID] = newEmptyLSDBEntry(lspEntry)
	l.lsps[lspEntry.LSPID].setSSN(from)
}

func (l *lsdb) _getLSPDU(needle packet.LSPID) *lsdbEntry {
	return l.lsps[needle]
}

func (l *lsdb) _exists(pkt *packet.LSPDU) bool {
	_, exists := l.lsps[pkt.LSPID]
	return exists
}

func (l *lsdb) _isNewer(pkt *packet.LSPDU) bool {
	return pkt.SequenceNumber > l.lsps[pkt.LSPID].lspdu.SequenceNumber
}

func (l *lsdb) processPSNP(psnp *packet.CSNP, from *netIfa) {
	l.lspsMu.Lock()
	defer l.lspsMu.Unlock()

	for _, lspEntry := range psnp.GetLSPEntries() {
		if _, exists := l.lsps[lspEntry.LSPID]; !exists {
			continue
		}

		l.lsps[lspEntry.LSPID].clearSRMFlag(from)
	}
}
