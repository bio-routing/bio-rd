package server

import (
	"sync"

	"github.com/bio-routing/bio-rd/protocols/isis/packet"
)

type lsdbEntry struct {
	lspdu      *packet.LSPDU
	srmFlags   map[*netIfa]struct{}
	srmFlagsMu sync.RWMutex
	ssnFlags   map[*netIfa]struct{}
	ssnFlagsMu sync.RWMutex
}

func newLSDBEntry(lspdu *packet.LSPDU) *lsdbEntry {
	return &lsdbEntry{
		srmFlags: make(map[*netIfa]struct{}),
		ssnFlags: make(map[*netIfa]struct{}),
	}
}

func (l *lsdbEntry) dropInterface(ifa *netIfa) {
	l.srmFlagsMu.Lock()
	defer l.srmFlagsMu.Unlock()

	l.ssnFlagsMu.Lock()
	defer l.ssnFlagsMu.Unlock()

	delete(l.srmFlags, ifa)
	delete(l.ssnFlags, ifa)
}

func (l *lsdbEntry) setSRM(ifa *netIfa) {
	l.srmFlagsMu.Lock()
	defer l.srmFlagsMu.Unlock()

	l.srmFlags[ifa] = struct{}{}
}

func (l *lsdbEntry) setSSN(ifa *netIfa) {
	l.ssnFlagsMu.Lock()
	defer l.ssnFlagsMu.Unlock()

	l.ssnFlags[ifa] = struct{}{}
}

func (l *lsdbEntry) getInterfacesSRMSet() []*netIfa {
	l.srmFlagsMu.Lock()
	defer l.srmFlagsMu.Unlock()

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
