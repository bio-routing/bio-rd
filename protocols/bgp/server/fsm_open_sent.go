package server

import (
	"bytes"
	"fmt"
	"math"
	"time"

	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
)

type openSentState struct {
	fsm                   *FSM
	peerASNRcvd           uint32
	multiplePeerRolesRcvd bool
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
		case <-time.After(time.Second):
			return s.checkHoldtimer()
		case recvMsg := <-s.fsm.msgRecvCh:
			return s.msgReceived(recvMsg, opt)
		}
	}
}

func (s *openSentState) checkHoldtimer() (state, string) {
	if time.Since(s.fsm.lastUpdateOrKeepalive) > s.fsm.holdTime {
		return s.holdTimerExpired()
	}

	return newOpenSentState(s.fsm), s.fsm.reason
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
		return s.openMsgReceived(msg.Body.(*packet.BGPOpen))
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

func (s *openSentState) openMsgReceived(openMsg *packet.BGPOpen) (state, string) {
	s.peerASNRcvd = uint32(openMsg.ASN)

	s.fsm.neighborID = openMsg.BGPIdentifier

	if s.fsm.isBMP {
		return s.handleOpenMessage(openMsg)
	}

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
		s.fsm.updateLastUpdateOrKeepalive()
		s.fsm.keepaliveTime = s.fsm.holdTime / 3
		s.fsm.keepaliveTimer = time.NewTimer(s.fsm.keepaliveTime)
	}

	s.peerASNRcvd = uint32(openMsg.ASN)
	s.processOpenOptions(openMsg.OptParams)

	if s.peerASNRcvd != s.fsm.peer.peerASN {
		s.fsm.sendNotification(packet.OpenMessageError, packet.BadPeerAS)
		return newIdleState(s.fsm), fmt.Sprintf("Bad Peer AS %d, expected: %d", s.peerASNRcvd, s.fsm.peer.peerASN)
	}

	// Validate Peer Role relationship for eBGP peers
	if !s.fsm.isBMP && s.fsm.peer.localASN != s.fsm.peer.peerASN {
		err := s.validatePeerRole()
		if err != nil {
			s.fsm.sendNotification(packet.OpenMessageError, packet.RoleMismatchError)
			return newIdleState(s.fsm), err.Error()
		}
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
	case packet.PeerRoleCapabilityCode:
		s.processPeerRoleCapability(cap.Value.(packet.PeerRoleCapability))
	}
}

func (s *openSentState) processMultiProtocolCapability(cap packet.MultiProtocolCapability) {
	if cap.SAFI != packet.SAFIUnicast {
		return
	}

	if cap.AFI == packet.AFIIPv4 && !s.fsm.peer.ipv4MultiProtocolAdvertised {
		return
	}

	f := s.fsm.addressFamily(cap.AFI, cap.SAFI)
	if f != nil {
		f.multiProtocol = true
	}
}

func (s *openSentState) processAddPathCapability(addPathCap packet.AddPathCapability) {
	for _, addPathCapTuple := range addPathCap {
		if addPathCapTuple.SAFI != packet.SAFIUnicast {
			continue
		}

		f := s.fsm.addressFamily(addPathCapTuple.AFI, addPathCapTuple.SAFI)
		if f == nil {
			continue
		}

		peerAddressFamily := s.fsm.peer.addressFamily(addPathCapTuple.AFI, addPathCapTuple.SAFI)

		switch addPathCapTuple.SendReceive {
		case packet.AddPathReceive:
			if !peerAddressFamily.addPathSend.BestOnly {
				f.addPathTX = peerAddressFamily.addPathSend
			}
		case packet.AddPathSend:
			if peerAddressFamily.addPathReceive {
				f.addPathRX = true
			}
		case packet.AddPathSendReceive:
			if !peerAddressFamily.addPathSend.BestOnly {
				f.addPathTX = peerAddressFamily.addPathSend
			}

			if peerAddressFamily.addPathReceive {
				f.addPathRX = true
			}
		}
	}
}

func (s *openSentState) processASN4Capability(cap packet.ASN4Capability) {
	s.fsm.supports4OctetASN = true

	if s.peerASNRcvd == packet.ASTransASN {
		s.peerASNRcvd = cap.ASN4
	}
}

func (s *openSentState) processPeerRoleCapability(cap packet.PeerRoleCapability) {
	if !s.fsm.peer.peerRoleEnabled {
		return
	}

	if s.fsm.peer.peerRoleAdvByPeer && s.fsm.peer.peerRoleRemote != cap.PeerRole {
		s.multiplePeerRolesRcvd = true
	}

	s.fsm.peer.peerRoleAdvByPeer = true
	s.fsm.peer.peerRoleRemote = cap.PeerRole
}

// Validate BGP Peer Role rules as defined in RFC9234
func (s *openSentState) validatePeerRole() error {
	// Peer Role checking is deactivated, do nothing
	if !s.fsm.peer.peerRoleEnabled {
		return nil
	}

	// We're running in strict mode and require the peer to advertise a peer role to us, but it didn't
	if s.fsm.peer.peerRoleStrictMode && !s.fsm.peer.peerRoleAdvByPeer {
		return fmt.Errorf("role misatch error: Strict mode configured but peer didn't advertise a BGP role")
	}

	// We're running in allow mode and no peer role was advertised
	if !s.fsm.peer.peerRoleAdvByPeer {
		return nil
	}

	if s.multiplePeerRolesRcvd {
		return fmt.Errorf("role misatch error: Multiple different BGP roles received from peer")
	}

	// Allowed pairs of role capabilities according to section 4.2 of RFC9234
	if isPeerRelationshipProviderClient(s.fsm.peer.peerRoleLocal, s.fsm.peer.peerRoleRemote) ||
		isPeerRelationshipRSClientRS(s.fsm.peer.peerRoleLocal, s.fsm.peer.peerRoleRemote) ||
		isPeerRelationshipPeerPeer(s.fsm.peer.peerRoleLocal, s.fsm.peer.peerRoleRemote) {
		return nil
	}

	return fmt.Errorf("role misatch error: Local role %v incompatible to remote role %v",
		packet.PeerRoleName(s.fsm.peer.peerRoleLocal), packet.PeerRoleName(s.fsm.peer.peerRoleRemote))
}

func isPeerRelationshipProviderClient(localRole uint8, remoteRole uint8) bool {
	if (localRole == packet.PeerRoleRoleProvider && remoteRole == packet.PeerRoleRoleCustomer) ||
		(localRole == packet.PeerRoleRoleCustomer && remoteRole == packet.PeerRoleRoleProvider) {
		return true
	}

	return false
}

func isPeerRelationshipRSClientRS(localRole uint8, remoteRole uint8) bool {
	if (localRole == packet.PeerRoleRoleRSClient && remoteRole == packet.PeerRoleRoleRS) ||
		(localRole == packet.PeerRoleRoleRS && remoteRole == packet.PeerRoleRoleRSClient) {
		return true
	}

	return false
}

func isPeerRelationshipPeerPeer(localRole uint8, remoteRole uint8) bool {
	return localRole == packet.PeerRoleRolePeer && remoteRole == packet.PeerRoleRolePeer
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
