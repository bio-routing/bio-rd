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
	log.WithFields(l.fields()).Debug("Setting SRM flags for interface %s", ifa.name)

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
	for _, lspEntry := range csnp.GetLSPEntries() {
		e := l._getLSPDU(lspEntry.LSPID)
		if e == nil {
			// LSP was unknown until now. We need to create an entry with Sequence number = 0
		}

		if e.sameAsInLSPEntry(lspEntry) {
			e.clearSRMFlag(from)
			continue
		}

		if e.newerInDatabase(lspEntry) {
			e.clearSSNFlag(from)
			e.setSRM(from)
			continue
		}

		if e.olderInDatabase(lspEntry) {
			e.clearSRMFlag(from)
			e.setSSN(from)
			continue
		}
	}

	// TODO: Check LSDB for entries in the range of the CSNP but not listed in it's TLVs.
	// Set SRM flag so we send the LSP our neighbor doesn't know about.
}

func (l *lsdb) _getLSPDU(needle packet.LSPID) *lsdbEntry {
	for _, e := range l.lsps {
		if e.lspdu.LSPID == needle {
			return e
		}
	}

	return nil
}
