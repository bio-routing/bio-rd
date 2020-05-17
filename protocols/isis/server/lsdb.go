package server

import (
	"sync"

	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	btime "github.com/bio-routing/bio-rd/util/time"
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

func (l *lsdb) dispose() {
	l.stop()
}

func (l *lsdb) start(t btime.Ticker) {
	l.wg.Add(1)
	go l.decrementRemainingLifetimesRoutine(t)
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
	for _, lsp := range l.lsps {
		lsp.setSRM(ifa)
	}
}
