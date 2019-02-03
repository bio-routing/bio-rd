package server

import (
	"bytes"
	"fmt"
	"time"

	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
)

type openConfirmState struct {
	fsm *FSM
}

func newOpenConfirmState(fsm *FSM) *openConfirmState {
	return &openConfirmState{
		fsm: fsm,
	}
}

func (s openConfirmState) run() (state, string) {
	opt := s.fsm.decodeOptions()

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
		case <-time.After(time.Second):
			return s.checkHoldtimer()
		case <-s.fsm.keepaliveTimer.C:
			return s.keepaliveTimerExpired()
		case recvMsg := <-s.fsm.msgRecvCh:
			return s.msgReceived(recvMsg, opt)
		}
	}
}

func (s *openConfirmState) checkHoldtimer() (state, string) {
	if time.Since(s.fsm.lastUpdateOrKeepalive) > s.fsm.holdTime {
		return s.holdTimerExpired()
	}

	return newOpenConfirmState(s.fsm), s.fsm.reason
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

func (s *openConfirmState) cease() (state, string) {
	s.fsm.sendNotification(packet.Cease, 0)
	s.fsm.con.Close()
	return newCeaseState(), "Cease"
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
	s.fsm.keepaliveTimer.Reset(s.fsm.keepaliveTime)
	return newOpenConfirmState(s.fsm), s.fsm.reason
}

func (s *openConfirmState) msgReceived(data []byte, opt *packet.DecodeOptions) (state, string) {
	msg, err := packet.Decode(bytes.NewBuffer(data), opt)
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
	s.fsm.updateLastUpdateOrKeepalive()
	return newEstablishedState(s.fsm), "Received KEEPALIVE"
}

func (s *openConfirmState) unexpectedMessage() (state, string) {
	s.fsm.sendNotification(packet.FiniteStateMachineError, 0)
	stopTimer(s.fsm.connectRetryTimer)
	s.fsm.con.Close()
	s.fsm.connectRetryCounter++
	return newIdleState(s.fsm), "FSM Error"
}
