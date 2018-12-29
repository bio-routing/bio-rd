package server

import (
	"fmt"
	"sync"
	"time"

	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	log "github.com/sirupsen/logrus"
)

type state interface {
	run() (state, string)
	getState() uint8
}

// FSM is the per neighbor finite state machine
type FSM struct {
	isisServer *ISISServer
	neighbor   *neighbor
	state      state
	stateMu    sync.Mutex
	pktCh      chan *packet.ISISPacket
	holdTimer  *time.Timer
	stopCh     chan struct{}
}

func newFSM(srv *ISISServer, n *neighbor) *FSM {
	fsm := &FSM{
		isisServer: srv,
		neighbor:   n,
		pktCh:      make(chan *packet.ISISPacket),
		stopCh:     make(chan struct{}),
	}

	fsm.state = newDownState(fsm)
	return fsm
}

func (fsm *FSM) start() {
	go fsm.run()
	return
}

func (fsm *FSM) run() {
	next, reason := fsm.state.run()
	for {
		newState := stateName(next)
		oldState := stateName(fsm.state)

		if oldState != newState {
			log.WithFields(log.Fields{
				"peer":       fsm.neighbor.systemID.String(),
				"last_state": oldState,
				"new_state":  newState,
				"reason":     reason,
			}).Info("ISIS FSM: Neighbor state change")
		}

		if newState == "down" {
			return
		}

		fsm.stateMu.Lock()
		fsm.state = next
		fsm.stateMu.Unlock()

		next, reason = fsm.state.run()
	}
}

func (fsm *FSM) receive(pkt *packet.ISISPacket) {
	fsm.pktCh <- pkt
	log.Debugf("Received PDU type %d on %s", pkt.Header.PDUType, fsm.neighbor.ifa.name)
}

func stateName(s state) string {
	switch s.(type) {
	case *initializingState:
		return "initializing"
	case *downState:
		return "down"
	case *upState:
		return "up"
	default:
		panic(fmt.Sprintf("Unknown state: %v", s))
	}
}
