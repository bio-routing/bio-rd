package server

import (
	"sync"
	"time"

	"github.com/bio-routing/bio-rd/protocols/isis/packet"
)

type state interface {
	run() (state, string)
	getState() uint8
}

type fsm struct {
	neighbor  *neighbor
	state     state
	stateMu   sync.Mutex
	pktCh     chan *packet.ISISPacket
	holdTimer *time.Timer
	done      chan struct{}
	wg        sync.WaitGroup
}

func newFSM(srv *Server, n *neighbor) *fsm {
	fsm := &fsm{
		neighbor: n,
		pktCh:    make(chan *packet.ISISPacket),
		done:     make(chan struct{}),
	}

	fsm.state = newFSMDownState(fsm)
	return fsm
}

func (f *fsm) start() {
	f.neighbor.dev.srv.log.Infof("Starting FSM for %s on %s", f.neighbor.macAddress.String(), f.neighbor.dev.name)
	f.wg.Add(1)
	go f.state.run()
}

func (f *fsm) dispose() {
	close(f.done)
	f.wg.Wait()
}

func (f *fsm) receivePacket(p *packet.ISISPacket) {
	f.pktCh <- p
}
