package server

import (
	"bytes"
	"fmt"

	tnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/adjRIBIn"
	"github.com/bio-routing/bio-rd/routingtable/adjRIBOut"
	"github.com/bio-routing/bio-rd/routingtable/adjRIBOutAddPath"
)

type establishedState struct {
	fsm *FSM
}

func newEstablishedState(fsm *FSM) *establishedState {
	return &establishedState{
		fsm: fsm,
	}
}

func (s establishedState) run() (state, string) {
	if !s.fsm.ribsInitialized {
		s.init()
	}

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
		case <-s.fsm.keepaliveTimer.C:
			return s.keepaliveTimerExpired()
		case recvMsg := <-s.fsm.msgRecvCh:
			return s.msgReceived(recvMsg)
		}
	}
}

func (s *establishedState) init() {
	s.fsm.adjRIBIn = adjRIBIn.New()

	s.fsm.peer.importFilter.Register(s.fsm.rib)
	s.fsm.adjRIBIn.Register(s.fsm.peer.importFilter)

	n := &routingtable.Neighbor{
		Type:    route.BGPPathType,
		Address: tnet.IPv4ToUint32(s.fsm.peer.addr),
	}

	clientOptions := routingtable.ClientOptions{
		BestOnly: true,
	}
	if s.fsm.capAddPathSend {
		s.fsm.updateSender = newUpdateSenderAddPath(s.fsm)
		s.fsm.adjRIBOut = adjRIBOutAddPath.New(n)
		clientOptions = s.fsm.peer.addPathSend
	} else {
		s.fsm.updateSender = newUpdateSender(s.fsm)
		s.fsm.adjRIBOut = adjRIBOut.New(n)
	}

	s.fsm.adjRIBOut.Register(s.fsm.updateSender)
	s.fsm.peer.exportFilter.Register(s.fsm.adjRIBOut)
	s.fsm.rib.RegisterWithOptions(s.fsm.peer.exportFilter, clientOptions)

	s.fsm.ribsInitialized = true
}

func (s *establishedState) uninit() {
	s.fsm.adjRIBIn.Unregister(s.fsm.peer.importFilter)
	s.fsm.peer.importFilter.Unregister(s.fsm.rib)

	s.fsm.rib.Unregister(s.fsm.peer.exportFilter)
	s.fsm.peer.exportFilter.Unregister(s.fsm.adjRIBOut)
	s.fsm.updateSender.Unregister(s.fsm.adjRIBOut)

	s.fsm.adjRIBIn = nil
	s.fsm.adjRIBOut = nil

	s.fsm.ribsInitialized = false
}

func (s *establishedState) manualStop() (state, string) {
	s.fsm.sendNotification(packet.Cease, 0)
	s.uninit()
	stopTimer(s.fsm.connectRetryTimer)
	s.fsm.con.Close()
	s.fsm.connectRetryCounter = 0
	return newIdleState(s.fsm), "Manual stop event"
}

func (s *establishedState) automaticStop() (state, string) {
	s.fsm.sendNotification(packet.Cease, 0)
	s.uninit()
	stopTimer(s.fsm.connectRetryTimer)
	s.fsm.con.Close()
	s.fsm.connectRetryCounter++
	return newIdleState(s.fsm), "Automatic stop event"
}

func (s *establishedState) cease() (state, string) {
	s.fsm.sendNotification(packet.Cease, 0)
	s.uninit()
	s.fsm.con.Close()
	return newCeaseState(), "Cease"
}

func (s *establishedState) holdTimerExpired() (state, string) {
	s.fsm.sendNotification(packet.HoldTimeExpired, 0)
	s.uninit()
	stopTimer(s.fsm.connectRetryTimer)
	s.fsm.con.Close()
	s.fsm.connectRetryCounter++
	return newIdleState(s.fsm), "Holdtimer expired"
}

func (s *establishedState) keepaliveTimerExpired() (state, string) {
	err := s.fsm.sendKeepalive()
	if err != nil {
		stopTimer(s.fsm.connectRetryTimer)
		s.fsm.con.Close()
		s.fsm.connectRetryCounter++
		return newIdleState(s.fsm), fmt.Sprintf("Failed to send keepalive: %v", err)
	}

	s.fsm.keepaliveTimer.Reset(s.fsm.keepaliveTime)
	return newEstablishedState(s.fsm), s.fsm.reason
}

func (s *establishedState) msgReceived(data []byte) (state, string) {
	msg, err := packet.Decode(bytes.NewBuffer(data))
	if err != nil {
		switch bgperr := err.(type) {
		case packet.BGPError:
			s.fsm.sendNotification(bgperr.ErrorCode, bgperr.ErrorSubCode)
		}
		stopTimer(s.fsm.connectRetryTimer)
		s.fsm.con.Close()
		s.fsm.connectRetryCounter++
		return newIdleState(s.fsm), "Failed to decode BGP message"
	}
	switch msg.Header.Type {
	case packet.NotificationMsg:
		return s.notification()
	case packet.UpdateMsg:
		return s.update(msg)
	case packet.KeepaliveMsg:
		return s.keepaliveReceived()
	default:
		return s.unexpectedMessage()
	}
}

func (s *establishedState) notification() (state, string) {
	stopTimer(s.fsm.connectRetryTimer)
	s.uninit()
	s.fsm.con.Close()
	s.fsm.connectRetryCounter++
	return newIdleState(s.fsm), "Received NOTIFICATION"
}

func (s *establishedState) update(msg *packet.BGPMessage) (state, string) {
	if s.fsm.holdTime != 0 {
		s.fsm.holdTimer.Reset(s.fsm.holdTime)
	}

	u := msg.Body.(*packet.BGPUpdate)
	s.withdraws(u)
	s.updates(u)

	return newEstablishedState(s.fsm), s.fsm.reason
}

func (s *establishedState) withdraws(u *packet.BGPUpdate) {
	for r := u.WithdrawnRoutes; r != nil; r = r.Next {
		pfx := tnet.NewPfx(r.IP, r.Pfxlen)
		s.fsm.adjRIBIn.RemovePath(pfx, nil)
	}
}

func (s *establishedState) updates(u *packet.BGPUpdate) {
	for r := u.NLRI; r != nil; r = r.Next {
		pfx := tnet.NewPfx(r.IP, r.Pfxlen)

		path := &route.Path{
			Type: route.BGPPathType,
			BGPPath: &route.BGPPath{
				Source: tnet.IPv4ToUint32(s.fsm.peer.addr),
			},
		}

		for pa := u.PathAttributes; pa != nil; pa = pa.Next {
			switch pa.TypeCode {
			case packet.OriginAttr:
				path.BGPPath.Origin = pa.Value.(uint8)
			case packet.LocalPrefAttr:
				path.BGPPath.LocalPref = pa.Value.(uint32)
			case packet.MEDAttr:
				path.BGPPath.MED = pa.Value.(uint32)
			case packet.NextHopAttr:
				path.BGPPath.NextHop = pa.Value.(uint32)
			case packet.ASPathAttr:
				path.BGPPath.ASPath = pa.ASPathString()
				path.BGPPath.ASPathLen = pa.ASPathLen()
			case packet.CommunitiesAttr:
				path.BGPPath.Communities = pa.CommunityString()
			case packet.LargeCommunitiesAttr:
				path.BGPPath.LargeCommunities = pa.LargeCommunityString()
			}
		}
		s.fsm.adjRIBIn.AddPath(pfx, path)
	}
}

func (s *establishedState) keepaliveReceived() (state, string) {
	if s.fsm.holdTime != 0 {
		s.fsm.holdTimer.Reset(s.fsm.holdTime)
	}
	return newEstablishedState(s.fsm), s.fsm.reason
}

func (s *establishedState) unexpectedMessage() (state, string) {
	s.fsm.sendNotification(packet.FiniteStateMachineError, 0)
	s.uninit()
	stopTimer(s.fsm.connectRetryTimer)
	s.fsm.con.Close()
	s.fsm.connectRetryCounter++
	return newIdleState(s.fsm), "FSM Error"
}
