package server

import "net"

type fsmManager struct {
	fsms map[string][]*FSM
}

func newFSMManager() *fsmManager {
	return &fsmManager{
		fsms: make(map[string][]*FSM, 0),
	}
}

func (m *fsmManager) resolveCollision(addr net.IP) {

}

func (m *fsmManager) newFSMPassive() *FSM {
	return &FSM{}
}

func (m *fsmManager) newFSMActive() *FSM {
	return &FSM{}
}
