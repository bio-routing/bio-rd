package server

import (
	"fmt"
	"sync"

	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	log "github.com/sirupsen/logrus"
)

type state interface {
	run() (state, string)
	getState() uint8
}

// FSM is the per neighbor finite state machine
type FSM struct {
	neighbor *neighbor
	state    state
	stateMu  sync.Mutex
	pktCh    chan *packet.ISISPacket
}

func newFSM(n *neighbor) *FSM {
	fsm := &FSM{
		neighbor: n,
		pktCh:    make(chan *packet.ISISPacket),
	}

	fsm.state = newInitializingState(fsm)
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
	log.Warningf("Received PDU type %d on %s", pkt.Header.PDUType, fsm.neighbor.ifa.name)
}

func stateName(s state) string {
	switch s.(type) {
	case *initializingState:
		return "initializing"
	case *downState:
		return "down"
	default:
		panic(fmt.Sprintf("Unknown state: %v", s))
	}
}
