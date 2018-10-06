package server

import "github.com/bio-routing/bio-rd/protocols/isis/packet"

type downState struct {
	fsm *FSM
}

func newDownState(fsm *FSM) *downState {
	return &downState{
		fsm: fsm,
	}
}

func (s downState) getState() uint8 {
	return packet.DOWN_STATE
}

func (s downState) run() (state, string) {
	return s, "This shall never be executed"
}
