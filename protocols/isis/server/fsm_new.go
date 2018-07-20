package server

type newState struct {
	fsm *FSM
}

func newNewState(fsm *FSM) *newState {
	return &newState{
		fsm: fsm,
	}
}

func (s newState) run() (state, string) {
	for {
		select {}
	}
}
