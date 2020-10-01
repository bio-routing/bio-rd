package server

import (
	"fmt"
	"net"
)

type connectState struct {
	fsm *FSM
}

func newConnectState(fsm *FSM) *connectState {
	return &connectState{
		fsm: fsm,
	}
}

func (s connectState) run() (state, string) {
	for {
		select {
		case e := <-s.fsm.eventCh:
			switch e {
			case ManualStop:
				return s.manualStop()
			case Cease:
				return newCeaseState(), "Cease"
			default:
				continue
			}
		case <-s.fsm.connectRetryTimer.C:
			s.connectRetryTimerExpired()
			continue
		case c := <-s.fsm.conCh:
			return s.connectionSuccess(c)
		}
	}
}

func (s *connectState) connectionSuccess(c net.Conn) (state, string) {
	err := s.fsm.sockSettings(c)
	if err != nil {
		return newIdleState(s.fsm), fmt.Sprintf("Unable to set socket options: %v", err)
	}

	s.fsm.con = c
	stopTimer(s.fsm.connectRetryTimer)
	err = s.fsm.sendOpen()
	if err != nil {
		return newIdleState(s.fsm), fmt.Sprintf("Unable to send open: %v", err)
	}
	return newOpenSentState(s.fsm), "TCP connection succeeded"
}

func (s *connectState) connectRetryTimerExpired() {
	s.fsm.resetConnectRetryTimer()
	s.fsm.tcpConnect()
}

func (s *connectState) manualStop() (state, string) {
	s.fsm.resetConnectRetryCounter()
	stopTimer(s.fsm.connectRetryTimer)
	return newIdleState(s.fsm), "Manual stop event"
}
