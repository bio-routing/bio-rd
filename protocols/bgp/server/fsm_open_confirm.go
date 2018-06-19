package server

import (
	"bytes"
	"fmt"
	"time"

	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
)

type openConfirmState struct {
	fsm *FSM2
}

func newOpenConfirmState(fsm *FSM2) *openConfirmState {
	return &openConfirmState{
		fsm: fsm,
	}
}

func (s *openConfirmState) run() (state, string) {
	for {
		select {
		case e := <-s.fsm.eventCh:
			if e == ManualStop {
				return s.manualStop()
			}
			continue
		case <-s.fsm.holdTimer.C:
			return s.holdTimerExpired()
		case <-s.fsm.keepaliveTimer.C:
			return s.keepaliveTimerExpired()
		case recvMsg := <-s.fsm.msgRecvCh:
			return s.msgReceived(recvMsg)
		}
	}
}

func (s *openConfirmState) manualStop() (state, string) {
	s.fsm.sendNotification(packet.Cease, 0)
	stopTimer(s.fsm.connectRetryTimer)
	s.fsm.con.Close()
	s.fsm.resetConnectRetryCounter()
	return newIdleState(s.fsm), "Manual stop event"
}

func (s *openConfirmState) automaticStop() (state, string) {
	s.fsm.sendNotification(packet.Cease, 0)
	stopTimer(s.fsm.connectRetryTimer)
	s.fsm.con.Close()
	s.fsm.connectRetryCounter++
	return newIdleState(s.fsm), "Automatic stop event"
}

func (s *openConfirmState) holdTimerExpired() (state, string) {
	s.fsm.sendNotification(packet.HoldTimeExpired, 0)
	stopTimer(s.fsm.connectRetryTimer)
	s.fsm.con.Close()
	s.fsm.connectRetryCounter++
	return newIdleState(s.fsm), "Holdtimer expired"
}

func (s *openConfirmState) keepaliveTimerExpired() (state, string) {
	err := s.fsm.sendKeepalive()
	if err != nil {
		stopTimer(s.fsm.connectRetryTimer)
		s.fsm.con.Close()
		s.fsm.connectRetryCounter++
		return newIdleState(s.fsm), fmt.Sprintf("Failed to send keepalive: %v", err)
	}
	s.fsm.keepaliveTimer.Reset(time.Second * s.fsm.keepaliveTime)
	return newOpenConfirmState(s.fsm), s.fsm.reason
}

func (s *openConfirmState) msgReceived(recvMsg msgRecvMsg) (state, string) {
	msg, err := packet.Decode(bytes.NewBuffer(recvMsg.msg))
	if err != nil {
		switch bgperr := err.(type) {
		case packet.BGPError:
			s.fsm.sendNotification(bgperr.ErrorCode, bgperr.ErrorSubCode)
		}
		stopTimer(s.fsm.connectRetryTimer)
		s.fsm.con.Close()
		s.fsm.connectRetryCounter++
		return newIdleState(s.fsm), fmt.Sprintf("Failed to decode BGP message: %v", err)
	}
	switch msg.Header.Type {
	case packet.NotificationMsg:
		return s.notification(msg)
	case packet.KeepaliveMsg:
		return s.keepaliveReceived()
	default:
		return s.unexpectedMessage()
	}
}

func (s *openConfirmState) notification(msg *packet.BGPMessage) (state, string) {
	stopTimer(s.fsm.connectRetryTimer)
	s.fsm.con.Close()
	nMsg := msg.Body.(*packet.BGPNotification)
	if nMsg.ErrorCode != packet.UnsupportedVersionNumber {
		s.fsm.connectRetryCounter++
	}

	return newIdleState(s.fsm), "Received NOTIFICATION"
}

func (s *openConfirmState) keepaliveReceived() (state, string) {
	s.fsm.holdTimer.Reset(time.Second * s.fsm.holdTime)
	return newEstablishedState(s.fsm), "Received KEEPALIVE"
}

func (s *openConfirmState) unexpectedMessage() (state, string) {
	s.fsm.sendNotification(packet.FiniteStateMachineError, 0)
	stopTimer(s.fsm.connectRetryTimer)
	s.fsm.con.Close()
	s.fsm.connectRetryCounter++
	return newIdleState(s.fsm), "FSM Error"
}
