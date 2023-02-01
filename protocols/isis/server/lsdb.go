package server

import (
	"sync"

	bbclock "github.com/benbjohnson/clock"
	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/bio-routing/bio-rd/util/log"
)

const lspRefreshThresholdSeconds = 300

type lsdb struct {
	srv       *Server
	lsps      map[packet.LSPID]*lsdbEntry
	lspsMu    sync.RWMutex
	done      chan struct{}
	wg        sync.WaitGroup
	refreshCh chan struct{}
}

func newLSDB(s *Server) *lsdb {
	return &lsdb{
		srv:       s,
		lsps:      make(map[packet.LSPID]*lsdbEntry),
		done:      make(chan struct{}),
		refreshCh: make(chan struct{}, 1),
	}
}

func (l *lsdb) level() int {
	if l.srv.lsdbL1 == l {
		return 1
	}

	return 2
}

func (l *lsdb) fields() log.Fields {
	return log.Fields{
		"level": l.level(),
	}
}

func (l *lsdb) dispose() {
	l.stop()
	l.srv = nil
}

func (l *lsdb) start(decrementTicker *bbclock.Ticker, minLSPTransTicker *bbclock.Ticker, psnpTransTicker *bbclock.Ticker, csnpTransTicker *bbclock.Ticker) {
	l.wg.Add(1)
	go l.decrementRemainingLifetimesRoutine(decrementTicker)

	l.wg.Add(1)
	go l.sendLSPDUsRoutine(minLSPTransTicker)

	l.wg.Add(1)
	go l.sendPSNPsRoutine(psnpTransTicker)

	l.wg.Add(1)
	go l.sendCSNPsRoutine(csnpTransTicker)

	l.wg.Add(1)
	go l.l2LSPUpdater()
}

func (l *lsdb) stop() {
	close(l.done)
	l.wg.Wait()
}

func (l *lsdb) decrementRemainingLifetimesRoutine(t *bbclock.Ticker) {
	defer l.wg.Done()

	for {
		select {
		case <-t.C:
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
		if lspid.SystemID == l.srv.systemID() && lspdbEntry.lspdu.RemainingLifetime < lspRefreshThresholdSeconds {
			l.requestL2LSPUpdate()
		}

		if lspdbEntry.lspdu.RemainingLifetime <= 1 {
			delete(l.lsps, lspid)
			continue
		}

		lspdbEntry.lspdu.RemainingLifetime--
	}
}

func (l *lsdb) sendLSPDUsRoutine(t *bbclock.Ticker) {
	defer l.wg.Done()

	for {
		select {
		case <-t.C:
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
			if ifa.cfg.Passive {
				continue
			}

			ifa.sendLSPDU(entry.lspdu, l.level())
		}
	}
}

func (l *lsdb) processCSNP(from *netIfa, csnp *packet.CSNP) {
	l.lspsMu.Lock()
	defer l.lspsMu.Unlock()

	for _, lspEntry := range csnp.GetLSPEntries() {
		l.processCSNPLSPEntry(lspEntry, from)
	}

	for lspID, lsdbEntry := range l.lsps {
		// we need to check if we have LSPs the neighbor did not describe.
		// For any that we have but our neighbor doesn't we set SRM flag so
		// the entry gets propagated.

		if lsdbEntry.lspdu.RemainingLifetime == 0 || lsdbEntry.lspdu.SequenceNumber == 0 {
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

func (l *lsdb) processPSNP(from *netIfa, psnp *packet.PSNP) {
	l.lspsMu.Lock()
	defer l.lspsMu.Unlock()

	for _, lspEntry := range psnp.GetLSPEntries() {
		if _, exists := l.lsps[lspEntry.LSPID]; !exists {
			continue
		}

		l.lsps[lspEntry.LSPID].clearSRMFlag(from)
	}
}

func (l *lsdb) sendCSNPsRoutine(t *bbclock.Ticker) {
	defer l.wg.Done()

	for {
		select {
		case <-t.C:
			l.sendCSNPss()
		case <-l.done:
			return
		}
	}
}

func (l *lsdb) sendCSNPss() {
	for _, ifa := range l.srv.netIfaManager.getAllInterfaces() {
		if len(ifa.neighborManagerL2.getNeighborsUp()) < 1 {
			continue
		}

		l.sendCSNPs(ifa)
	}
}

func (l *lsdb) sendPSNPsRoutine(t *bbclock.Ticker) {
	defer l.wg.Done()

	for {
		select {
		case <-t.C:
			l.sendPSNPss()
		case <-l.done:
			return
		}
	}
}

func (l *lsdb) sendPSNPss() {
	l.lspsMu.Lock()
	defer l.lspsMu.Unlock()

	srcID := types.SourceID{
		SystemID: l.srv.nets[0].SystemID,
	}

	for _, ifa := range l.srv.netIfaManager.getAllInterfaces() {
		if ifa.cfg.Passive {
			continue
		}

		lspdus := l._getLSPWithSSNSet(ifa)
		for _, psnp := range packet.NewPSNPs(srcID, lspdus, ifa.ethernetInterface.GetMTU()) {
			ifa.sendPSNP(&psnp, l.level())
		}
	}

	l._clearAllSSNFlags()
}

func (l *lsdb) _getLSPWithSSNSet(ifa *netIfa) []*packet.LSPEntry {
	ret := make([]*packet.LSPEntry, 0)
	for _, lsp := range l.lsps {
		if !lsp.getSSN(ifa) {
			continue
		}

		ret = append(ret, lsp.lspdu.ToLSPEntry())
	}

	return ret
}

func (l *lsdb) _clearAllSSNFlags() {
	for _, e := range l.lsps {
		e.clearAllSSNFlags()
	}
}

func (l *lsdb) getLSPEntries() []*packet.LSPEntry {
	l.lspsMu.Lock()
	defer l.lspsMu.Unlock()

	ret := make([]*packet.LSPEntry, 0, len(l.lsps))
	for _, e := range l.lsps {
		ret = append(ret, e.lspdu.ToLSPEntry())
	}

	return ret
}

func (l *lsdb) getCSNPs(ifa *netIfa) []packet.CSNP {
	srcID := types.SourceID{
		SystemID: l.srv.nets[0].SystemID,
	}

	return packet.NewCSNPs(srcID, l.getLSPEntries(), ifa.ethernetInterface.GetMTU())
}

func (l *lsdb) sendCSNPs(ifa *netIfa) {
	for _, c := range l.getCSNPs(ifa) {
		ifa.sendCSNP(&c, l.level())
	}
}

func (l *lsdb) processLSP(ifa *netIfa, lspdu *packet.LSPDU) {
	log.Debug("Processing received LSP")
	l.lspsMu.Lock()
	defer l.lspsMu.Unlock()

	existingLSDBEntry, exists := l.lsps[lspdu.LSPID]
	if !exists || lspdu.SequenceNumber > existingLSDBEntry.lspdu.SequenceNumber {
		log.Debugf("ISIS: Received newer LSPDU %v sequence number %d", lspdu.LSPID, lspdu.SequenceNumber)
		l.processNewerLSPDU(ifa, lspdu)
		return
	}

	if lspdu.SequenceNumber == existingLSDBEntry.lspdu.SequenceNumber {
		log.Debugf("ISIS: Received same sequence LSPDU %v sequence number %d", lspdu.LSPID, lspdu.SequenceNumber)
		existingLSDBEntry.processSameLSPDU(ifa)
		return
	}

	log.Debugf("ISIS: Received older LSPDU %v sequence number %d / %d", lspdu.LSPID, existingLSDBEntry.lspdu.SequenceNumber, lspdu.SequenceNumber)
	existingLSDBEntry.newerLocalLSPDU(ifa)
}

func (l *lsdb) processNewerLSPDU(ifa *netIfa, lspdu *packet.LSPDU) {
	lsdbEntry := newLSDBEntry(lspdu)

	for _, i := range l.srv.netIfaManager.getAllInterfacesExcept(ifa) {
		lsdbEntry.setSRM(i)
	}

	lsdbEntry.clearSRMFlag(ifa)
	lsdbEntry.setSSN(ifa)

	l.lsps[lspdu.LSPID] = lsdbEntry
	return
}

// requestL2LSPUpdate queues an update request if none is pending
func (l *lsdb) requestL2LSPUpdate() {
	select {
	case l.refreshCh <- struct{}{}:
		return
	default:
		return
	}
}

func (l *lsdb) l2LSPUpdater() {
	defer l.wg.Done()

	for {
		select {
		case <-l.done:
			return
		case <-l.refreshCh:
			l.updateL2LSP()
		}
	}
}

func (l *lsdb) updateL2LSP() {
	lspdu := l.srv.generateLocalLSP()
	lsdbEntry := newLSDBEntry(lspdu)

	for _, ifa := range l.srv.netIfaManager.getAllInterfaces() {
		lsdbEntry.setSRM(ifa)
	}

	l.lspsMu.Lock()
	defer l.lspsMu.Unlock()

	l.lsps[lspdu.LSPID] = lsdbEntry
}
