package server

import (
	"sync"

	"github.com/bio-routing/bio-rd/protocols/isis/packet"
)

type lsdbEntry struct {
	lspdu    *packet.LSPDU
	srmFlags map[*netIfa]struct{}
	ssnFlags map[*netIfa]struct{}
	mutex    sync.RWMutex
}

type LSDBEntry struct {
	lspdu    *packet.LSPDU
	srmFlags []string
	ssnFlags []string
}

func (l *LSDBEntry) GetLSPDU() *packet.LSPDU {
	return l.lspdu
}

func newLSDBEntry(lspdu *packet.LSPDU) *lsdbEntry {
	return &lsdbEntry{
		lspdu:    lspdu,
		srmFlags: make(map[*netIfa]struct{}),
		ssnFlags: make(map[*netIfa]struct{}),
	}
}

func newEmptyLSDBEntry(lspEntry *packet.LSPEntry) *lsdbEntry {
	return &lsdbEntry{
		lspdu: &packet.LSPDU{
			RemainingLifetime: lspEntry.RemainingLifetime,
			LSPID:             lspEntry.LSPID,
			SequenceNumber:    0,
			Checksum:          lspEntry.LSPChecksum,
		},
		srmFlags: make(map[*netIfa]struct{}),
		ssnFlags: make(map[*netIfa]struct{}),
	}
}

func (l *lsdbEntry) Export() *LSDBEntry {
	c := &LSDBEntry{
		lspdu:    l.lspdu.Copy(),
		srmFlags: make([]string, 0),
		ssnFlags: make([]string, 0),
	}

	for ifa := range l.srmFlags {
		c.srmFlags = append(c.srmFlags, ifa.getName())
	}

	for ifa := range l.ssnFlags {
		c.ssnFlags = append(c.ssnFlags, ifa.getName())
	}

	return c
}

func (l *lsdbEntry) dropInterface(ifa *netIfa) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	delete(l.srmFlags, ifa)
	delete(l.ssnFlags, ifa)
}

func (l *lsdbEntry) setSRM(ifa *netIfa) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.lspdu.SequenceNumber == 0 {
		return
	}

	l.srmFlags[ifa] = struct{}{}
}

func (l *lsdbEntry) setSSN(ifa *netIfa) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.ssnFlags[ifa] = struct{}{}
}

func (l *lsdbEntry) getSSN(ifa *netIfa) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	_, exists := l.ssnFlags[ifa]
	return exists
}

func (l *lsdbEntry) getInterfacesSRMSet() []*netIfa {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if len(l.srmFlags) == 0 {
		return nil
	}

	ret := make([]*netIfa, len(l.srmFlags))

	i := 0
	for ifa := range l.srmFlags {
		ret[i] = ifa
		i++
	}

	return ret
}

func (l *lsdbEntry) clearSRMFlag(ifa *netIfa) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	delete(l.srmFlags, ifa)
}

func (l *lsdbEntry) clearSSNFlag(ifa *netIfa) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	delete(l.ssnFlags, ifa)
}

func (l *lsdbEntry) clearAllSSNFlags() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.ssnFlags = make(map[*netIfa]struct{})
}

func (l *lsdbEntry) sameAsInLSPEntry(needle *packet.LSPEntry) bool {
	return l.lspdu.LSPID == needle.LSPID && l.lspdu.SequenceNumber == needle.SequenceNumber
}

func (l *lsdbEntry) newerInDatabase(x *packet.LSPEntry) bool {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	return l.lspdu.SequenceNumber > x.SequenceNumber
}

func (l *lsdbEntry) olderInDatabase(x *packet.LSPEntry) bool {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	return l.lspdu.SequenceNumber < x.SequenceNumber
}

func (l *lsdbEntry) processSameLSPDU(ifa *netIfa) {
	l.clearSRMFlag(ifa)
	l.setSSN(ifa)
}

func (l *lsdbEntry) newerLocalLSPDU(ifa *netIfa) {
	l.setSRM(ifa)
	l.clearSSNFlag(ifa)
}
