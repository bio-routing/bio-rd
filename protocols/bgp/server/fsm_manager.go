package server

import "net"

type fsmManager struct {
	fsms map[string][]*FSM2
}

func newFSMManager() *fsmManager {
	return &fsmManager{
		fsms: make(map[string][]*FSM2, 0),
	}
}

func (m *fsmManager) resolveCollision(addr net.IP) {

}

func (m *fsmManager) newFSMPassive() *FSM2 {
	return &FSM2{}
}

func (m *fsmManager) newFSMActive() *FSM2 {
	return &FSM2{}
}
