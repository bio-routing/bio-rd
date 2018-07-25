package server

import (
	"bytes"
	"fmt"
	"math"
	"time"

	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
)

type openSentState struct {
	fsm         *FSM
	peerASNRcvd uint32
}

func newOpenSentState(fsm *FSM) *openSentState {
	return &openSentState{
		fsm: fsm,
	}
}

func (s openSentState) run() (state, string) {
	go s.fsm.msgReceiver()

	opt := s.fsm.decodeOptions()

	for {
		select {
		case e := <-s.fsm.eventCh:
			switch e {
			case ManualStop:
				return s.manualStop()
			case AutomaticStop:
				return s.automaticStop()
			case Cease:
				return s.cease()
			default:
				continue
			}
		case <-s.fsm.holdTimer.C:
			return s.holdTimerExpired()
		case recvMsg := <-s.fsm.msgRecvCh:
			return s.msgReceived(recvMsg, opt)
		}
	}
}

func (s *openSentState) manualStop() (state, string) {
	s.fsm.sendNotification(packet.Cease, 0)
	s.fsm.resetConnectRetryTimer()
	s.fsm.con.Close()
	s.fsm.resetConnectRetryCounter()
	return newIdleState(s.fsm), "Manual stop event"
}

func (s *openSentState) automaticStop() (state, string) {
	s.fsm.sendNotification(packet.Cease, 0)
	s.fsm.resetConnectRetryTimer()
	s.fsm.con.Close()
	s.fsm.connectRetryCounter++
	return newIdleState(s.fsm), "Automatic stop event"
}

func (s *openSentState) cease() (state, string) {
	s.fsm.sendNotification(packet.Cease, 0)
	s.fsm.con.Close()
	return newCeaseState(), "Cease"
}

func (s *openSentState) holdTimerExpired() (state, string) {
	s.fsm.sendNotification(packet.HoldTimeExpired, 0)
	stopTimer(s.fsm.connectRetryTimer)
	s.fsm.con.Close()
	s.fsm.connectRetryCounter++
	return newIdleState(s.fsm), "Holdtimer expired"
}

func (s *openSentState) msgReceived(data []byte, opt *packet.DecodeOptions) (state, string) {
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
	case packet.OpenMsg:
		return s.openMsgReceived(msg)
	default:
		return s.unexpectedMessage()
	}
}

func (s *openSentState) unexpectedMessage() (state, string) {
	s.fsm.sendNotification(packet.FiniteStateMachineError, 0)
	stopTimer(s.fsm.connectRetryTimer)
	s.fsm.con.Close()
	s.fsm.connectRetryCounter++
	return newIdleState(s.fsm), "FSM Error"
}

func (s *openSentState) openMsgReceived(msg *packet.BGPMessage) (state, string) {
	openMsg := msg.Body.(*packet.BGPOpen)
	s.peerASNRcvd = uint32(openMsg.ASN)

	s.fsm.neighborID = openMsg.BGPIdentifier
	stopTimer(s.fsm.connectRetryTimer)
	if s.fsm.peer.collisionHandling(s.fsm) {
		return s.cease()
	}
	err := s.fsm.sendKeepalive()
	if err != nil {
		return s.tcpFailure()
	}

	return s.handleOpenMessage(openMsg)
}

func (s *openSentState) handleOpenMessage(openMsg *packet.BGPOpen) (state, string) {
	s.fsm.holdTime = time.Duration(math.Min(float64(s.fsm.peer.holdTime), float64(time.Duration(openMsg.HoldTime)*time.Second)))
	if s.fsm.holdTime != 0 {
		if !s.fsm.holdTimer.Reset(s.fsm.holdTime) {
			<-s.fsm.holdTimer.C
		}
		s.fsm.keepaliveTime = s.fsm.holdTime / 3
		s.fsm.keepaliveTimer = time.NewTimer(s.fsm.keepaliveTime)
	}

	s.peerASNRcvd = uint32(openMsg.ASN)
	s.processOpenOptions(openMsg.OptParams)

	if s.peerASNRcvd != s.fsm.peer.peerASN {
		s.fsm.sendNotification(packet.OpenMessageError, packet.BadPeerAS)
		return newCeaseState(), fmt.Sprintf("Bad Peer AS %d, expected: %d", s.peerASNRcvd, s.fsm.peer.peerASN)
	}

	return newOpenConfirmState(s.fsm), "Received OPEN message"
}

func (s *openSentState) tcpFailure() (state, string) {
	s.fsm.con.Close()
	s.fsm.resetConnectRetryTimer()
	return newActiveState(s.fsm), "TCP connection failure"
}

func (s *openSentState) processOpenOptions(optParams []packet.OptParam) {
	for _, optParam := range optParams {
		if optParam.Type != packet.CapabilitiesParamType {
			continue
		}

		s.processCapabilities(optParam.Value.(packet.Capabilities))
	}
}

func (s *openSentState) processCapabilities(caps packet.Capabilities) {
	for _, cap := range caps {
		s.processCapability(cap)
	}
}

func (s *openSentState) processCapability(cap packet.Capability) {
	switch cap.Code {
	case packet.AddPathCapabilityCode:
		s.processAddPathCapability(cap.Value.(packet.AddPathCapability))
	case packet.ASN4CapabilityCode:
		s.processASN4Capability(cap.Value.(packet.ASN4Capability))
	case packet.MultiProtocolCapabilityCode:
		s.processMultiProtocolCapability(cap.Value.(packet.MultiProtocolCapability))
	}
}

func (s *openSentState) processMultiProtocolCapability(cap packet.MultiProtocolCapability) {
	if cap.SAFI != packet.UnicastSAFI {
		return
	}

	f := s.fsm.addressFamily(cap.AFI, cap.SAFI)
	if f != nil {
		f.multiProtocol = true
	}
}

func (s *openSentState) processAddPathCapability(addPathCap packet.AddPathCapability) {
	if addPathCap.SAFI != packet.UnicastSAFI {
		return
	}

	f := s.fsm.addressFamily(addPathCap.AFI, addPathCap.SAFI)
	if f == nil {
		return
	}

	switch addPathCap.SendReceive {
	case packet.AddPathReceive:
		if !f.addPathSend.BestOnly {
			f.addPathTX = true
		}
	case packet.AddPathSend:
		if f.addPathTXConfigured {
			f.addPathTX = true
		}
	case packet.AddPathSendReceive:
		if !f.addPathSend.BestOnly {
			f.addPathTX = true
		}
		if f.addPathTXConfigured {
			f.addPathTX = true
		}
	}
}

func (s *openSentState) processASN4Capability(cap packet.ASN4Capability) {
	s.fsm.supports4OctetASN = true

	if s.peerASNRcvd == packet.ASTransASN {
		s.peerASNRcvd = cap.ASN4
	}
}

func (s *openSentState) notification(msg *packet.BGPMessage) (state, string) {
	stopTimer(s.fsm.connectRetryTimer)
	s.fsm.con.Close()
	nMsg := msg.Body.(*packet.BGPNotification)
	if nMsg.ErrorCode != packet.UnsupportedVersionNumber {
		s.fsm.connectRetryCounter++
	}

	return newIdleState(s.fsm), "Received NOTIFICATION"
}
