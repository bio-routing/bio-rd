package server

import (
	"fmt"
	"net"
	"time"
)

type activeState struct {
	fsm *FSM
}

func newActiveState(fsm *FSM) *activeState {
	return &activeState{
		fsm: fsm,
	}
}

func (s activeState) run() (state, string) {
	for {
		select {
		case e := <-s.fsm.eventCh:
			switch e {
			case ManualStop:
				return s.manualStop()
			case Cease:
				return s.cease()
			default:
				continue
			}
		case <-s.fsm.connectRetryTimer.C:
			return s.connectRetryTimerExpired()
		case c := <-s.fsm.conCh:
			return s.connectionSuccess(c)
		}
	}
}

func (s *activeState) manualStop() (state, string) {
	s.fsm.con.Close()
	s.fsm.resetConnectRetryCounter()
	stopTimer(s.fsm.connectRetryTimer)
	return newIdleState(s.fsm), "Manual stop event"
}

func (s *activeState) cease() (state, string) {
	s.fsm.con.Close()
	return newCeaseState(), "Cease"
}

func (s *activeState) connectRetryTimerExpired() (state, string) {
	s.fsm.resetConnectRetryTimer()
	s.fsm.tcpConnect()
	return newConnectState(s.fsm), "Connect retry timer expired"
}

func (s *activeState) connectionSuccess(con net.Conn) (state, string) {
	s.fsm.con = con
	stopTimer(s.fsm.connectRetryTimer)
	err := s.fsm.sendOpen()
	if err != nil {
		s.fsm.resetConnectRetryTimer()
		s.fsm.connectRetryCounter++
		return newIdleState(s.fsm), fmt.Sprintf("Sending OPEN message failed: %v", err)
	}
	s.fsm.holdTimer = time.NewTimer(time.Minute * 4)
	return newOpenSentState(s.fsm), "Sent OPEN message"
}
