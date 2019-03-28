package server

import (
	"bytes"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
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
		err := s.init()
		if err != nil {
			return newCeaseState(), fmt.Sprintf("Init failed: %v", err)
		}
	}

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
		case <-s.fsm.keepaliveTimer.C:
			return s.keepaliveTimerExpired()
		case <-time.After(time.Second):
			return s.checkHoldtimer()
		case recvMsg := <-s.fsm.msgRecvCh:
			return s.msgReceived(recvMsg, opt)
		}
	}
}

func (s *establishedState) checkHoldtimer() (state, string) {
	if time.Since(s.fsm.lastUpdateOrKeepalive) > s.fsm.holdTime {
		return s.holdTimerExpired()
	}

	return newEstablishedState(s.fsm), s.fsm.reason
}

func (s *establishedState) init() error {
	host, _, err := net.SplitHostPort(s.fsm.con.LocalAddr().String())
	if err != nil {
		return errors.Wrap(err, "Unable to get local address")
	}
	localAddr, err := bnet.IPFromString(host)
	if err != nil {
		return errors.Wrap(err, "Unable to parse address")
	}

	n := &routingtable.Neighbor{
		Type:                 route.BGPPathType,
		Address:              s.fsm.peer.addr,
		IBGP:                 s.fsm.peer.localASN == s.fsm.peer.peerASN,
		LocalASN:             s.fsm.peer.localASN,
		RouteServerClient:    s.fsm.peer.routeServerClient,
		LocalAddress:         localAddr,
		RouteReflectorClient: s.fsm.peer.routeReflectorClient,
		ClusterID:            s.fsm.peer.clusterID,
	}

	if s.fsm.ipv4Unicast != nil {
		s.fsm.ipv4Unicast.init(n)
	}

	if s.fsm.ipv6Unicast != nil {
		s.fsm.ipv6Unicast.init(n)
	}

	s.fsm.ribsInitialized = true
	return nil
}

func (s *establishedState) uninit() {
	if s.fsm.ipv4Unicast != nil {
		s.fsm.ipv4Unicast.dispose()
	}

	if s.fsm.ipv6Unicast != nil {
		s.fsm.ipv6Unicast.dispose()
	}

	s.fsm.counters.reset()
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
		s.uninit()
		stopTimer(s.fsm.connectRetryTimer)
		s.fsm.con.Close()
		s.fsm.connectRetryCounter++
		return newIdleState(s.fsm), fmt.Sprintf("Failed to send keepalive: %v", err)
	}

	s.fsm.keepaliveTimer.Reset(s.fsm.keepaliveTime)
	return newEstablishedState(s.fsm), s.fsm.reason
}

func (s *establishedState) msgReceived(data []byte, opt *packet.DecodeOptions) (state, string) {
	msg, err := packet.Decode(bytes.NewBuffer(data), opt)
	if err != nil {
		switch bgperr := err.(type) {
		case packet.BGPError:
			s.fsm.sendNotification(bgperr.ErrorCode, bgperr.ErrorSubCode)
		}
		stopTimer(s.fsm.connectRetryTimer)
		if s.fsm.con != nil {
			s.fsm.con.Close()
		}
		s.fsm.connectRetryCounter++
		return newIdleState(s.fsm), fmt.Sprintf("Failed to decode BGP message: %v", err)
	}

	switch msg.Header.Type {
	case packet.NotificationMsg:
		fmt.Println(data)
		return s.notification()
	case packet.UpdateMsg:
		return s.update(msg.Body.(*packet.BGPUpdate))
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

func (s *establishedState) update(u *packet.BGPUpdate) (state, string) {
	atomic.AddUint64(&s.fsm.counters.updatesReceived, 1)

	if s.fsm.holdTime != 0 {
		s.fsm.updateLastUpdateOrKeepalive()
	}

	if s.fsm.ipv4Unicast != nil {
		s.fsm.ipv4Unicast.processUpdate(u)
	}

	if s.fsm.ipv6Unicast != nil {
		s.fsm.ipv6Unicast.processUpdate(u)
	}

	afi, safi := s.updateAddressFamily(u)

	if safi != packet.UnicastSAFI {
		// only unicast support, so other SAFIs are ignored
		return newEstablishedState(s.fsm), s.fsm.reason
	}

	switch afi {
	case packet.IPv4AFI:
		if s.fsm.ipv4Unicast == nil {
			log.Warnf("Received update for family IPv4 unicast, but this family is not configured.")
		}

	case packet.IPv6AFI:
		if s.fsm.ipv6Unicast == nil {
			log.Warnf("Received update for family IPv6 unicast, but this family is not configured.")
		}

	}

	return newEstablishedState(s.fsm), s.fsm.reason
}

func (s *establishedState) updateAddressFamily(u *packet.BGPUpdate) (afi uint16, safi uint8) {
	if u.WithdrawnRoutes != nil || u.NLRI != nil {
		return packet.IPv4AFI, packet.UnicastSAFI
	}

	for cur := u.PathAttributes; cur != nil; cur = cur.Next {
		if cur.TypeCode == packet.MultiProtocolReachNLRICode {
			a := cur.Value.(packet.MultiProtocolReachNLRI)
			return a.AFI, a.SAFI
		}

		if cur.TypeCode == packet.MultiProtocolUnreachNLRICode {
			a := cur.Value.(packet.MultiProtocolUnreachNLRI)
			return a.AFI, a.SAFI
		}
	}

	return
}

func (s *establishedState) keepaliveReceived() (state, string) {
	if s.fsm.holdTime != 0 {
		s.fsm.updateLastUpdateOrKeepalive()
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
