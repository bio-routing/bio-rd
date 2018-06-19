package server

type activeState struct {
	fsm *FSM2
}

func newActiveState(fsm *FSM2) *activeState {
	return &activeState{
		fsm: fsm,
	}
}

func (s *activeState) run() (state, string) {
	for {
		select {
		case e := <-s.fsm.eventCh:
			if e == ManualStop {

			}
			continue
		case <-s.fsm.connectRetryTimer.C:

		case c := <-s.fsm.conCh:

		}
	}
}
