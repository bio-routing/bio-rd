package server

import (
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"
)

type state interface {
	run() (state, string)
}

// FSM is the per neighbor finite state machine
type FSM struct {
	neighbor *neighbor
	state    state
	stateMu  sync.Mutex
}

func newFSM(n *neighbor) *FSM {
	fsm := &FSM{
		neighbor: n,
	}

	fsm.state = newNewState(fsm)
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
			}).Info("FSM: Neighbor state change")
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

func stateName(s state) string {
	switch s.(type) {
	case *newState:
		return "new"
	/*case *oneWayState:
		return "oneway"
	case *initializingState:
		return "initializing"
	case *upState:
		return "up"
	case *downState:
		return "down"
	case *rejectState:
		return "reject"*/
	default:
		panic(fmt.Sprintf("Unknown state: %v", s))
	}
}
