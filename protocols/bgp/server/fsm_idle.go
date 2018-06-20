package server

import (
	"time"
)

type idleState struct {
	fsm            *FSM
	newStateReason string
}

func newIdleState(fsm *FSM) *idleState {
	return &idleState{
		fsm: fsm,
	}
}

func (s idleState) run() (state, string) {
	if s.fsm.peer.reconnectInterval != 0 {
		time.Sleep(s.fsm.peer.reconnectInterval)
		go s.fsm.activate()
	}
	for {
		event := <-s.fsm.eventCh
		switch event {
		case ManualStart:
			return s.manualStart()
		case AutomaticStart:
			return s.automaticStart()
		case Cease:
			return newCeaseState(), "Cease"
		default:
			continue
		}
	}
}

func (s *idleState) manualStart() (state, string) {
	s.newStateReason = "Received ManualStart event"
	return s.start()
}

func (s *idleState) automaticStart() (state, string) {
	s.newStateReason = "Received AutomaticStart event"
	return s.start()
}

func (s *idleState) start() (state, string) {
	s.fsm.resetConnectRetryCounter()
	s.fsm.startConnectRetryTimer()
	go s.fsm.tcpConnect()

	return newConnectState(s.fsm), s.newStateReason
}
