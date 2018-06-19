package server

type idleState struct {
	fsm            *FSM2
	newStateReason string
}

func newIdleState(fsm *FSM2) *idleState {
	return &idleState{
		fsm: fsm,
	}
}

func (s *idleState) run() (state, string) {
	for {
		switch <-s.fsm.eventCh {
		case ManualStart:
			s.manualStart()
		case AutomaticStart:
			s.automaticStart()
		default:
			continue
		}

		return newConnectState(s.fsm), s.newStateReason
	}
}

func (s *idleState) manualStart() {
	s.newStateReason = "Received ManualStart event"
	s.start()
}

func (s *idleState) automaticStart() {
	s.newStateReason = "Received AutomaticStart event"
	s.start()
}

func (s *idleState) start() {
	s.fsm.resetConnectRetryCounter()
	s.fsm.startConnectRetryTimer()
	if s.fsm.active {
		s.fsm.tcpConnect()
	}
}
