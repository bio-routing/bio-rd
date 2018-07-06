package server

import (
	"bytes"
	"fmt"
	"net"
	"time"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/adjRIBIn"
	"github.com/bio-routing/bio-rd/routingtable/adjRIBOut"
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

func (s *establishedState) init() error {
	contributingASNs := s.fsm.rib.GetContributingASNs()

	s.fsm.adjRIBIn = adjRIBIn.New(s.fsm.peer.importFilter, contributingASNs, s.fsm.peer.routerID, s.fsm.peer.clusterID)
	contributingASNs.Add(s.fsm.peer.localASN)
	s.fsm.adjRIBIn.Register(s.fsm.rib)

	host, _, err := net.SplitHostPort(s.fsm.con.LocalAddr().String())
	if err != nil {
		return fmt.Errorf("Unable to get local address: %v", err)
	}
	hostIP := net.ParseIP(host)
	if hostIP == nil {
		return fmt.Errorf("Unable to parse address")
	}
	localAddr, err := bnet.IPFromBytes(hostIP)
	if err != nil {
		return fmt.Errorf("Unable to parse address: %v", err)
	}

	n := &routingtable.Neighbor{
		Type:                 route.BGPPathType,
		Address:              s.fsm.peer.addr,
		IBGP:                 s.fsm.peer.localASN == s.fsm.peer.peerASN,
		LocalASN:             s.fsm.peer.localASN,
		RouteServerClient:    s.fsm.peer.routeServerClient,
		LocalAddress:         localAddr,
		CapAddPathRX:         s.fsm.options.AddPathRX,
		RouteReflectorClient: s.fsm.peer.routeReflectorClient,
		ClusterID:            s.fsm.peer.clusterID,
	}

	s.fsm.adjRIBOut = adjRIBOut.New(n, s.fsm.peer.exportFilter)
	clientOptions := routingtable.ClientOptions{
		BestOnly: true,
	}
	if s.fsm.options.AddPathRX {
		clientOptions = s.fsm.peer.addPathSend
	}

	s.fsm.updateSender = newUpdateSender(s.fsm)
	s.fsm.updateSender.Start(time.Millisecond * 5)

	s.fsm.adjRIBOut.Register(s.fsm.updateSender)
	s.fsm.rib.RegisterWithOptions(s.fsm.adjRIBOut, clientOptions)

	s.fsm.ribsInitialized = true
	return nil
}

func (s *establishedState) uninit() {
	s.fsm.rib.GetContributingASNs().Remove(s.fsm.peer.localASN)
	s.fsm.adjRIBIn.Unregister(s.fsm.rib)
	s.fsm.rib.Unregister(s.fsm.adjRIBOut)
	s.fsm.adjRIBOut.Unregister(s.fsm.updateSender)
	s.fsm.updateSender.Destroy()

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
	msg, err := packet.Decode(bytes.NewBuffer(data), s.fsm.options)
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
		fmt.Println(data)
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
	s.multiProtocolUpdates(u)

	return newEstablishedState(s.fsm), s.fsm.reason
}

func (s *establishedState) withdraws(u *packet.BGPUpdate) {
	for r := u.WithdrawnRoutes; r != nil; r = r.Next {
		pfx := bnet.NewPfx(bnet.IPv4(r.IP), r.Pfxlen)
		s.fsm.adjRIBIn.RemovePath(pfx, nil)
	}
}

func (s *establishedState) updates(u *packet.BGPUpdate) {
	for r := u.NLRI; r != nil; r = r.Next {
		pfx := bnet.NewPfx(bnet.IPv4(r.IP), r.Pfxlen)

		path := s.newRoutePath()
		s.processAttributes(u.PathAttributes, path)

		s.fsm.adjRIBIn.AddPath(pfx, path)
	}
}

func (s *establishedState) multiProtocolUpdates(u *packet.BGPUpdate) {
	if !s.fsm.options.SupportsMultiProtocol {
		return
	}

	path := s.newRoutePath()
	s.processAttributes(u.PathAttributes, path)

	for pa := u.PathAttributes; pa != nil; pa = pa.Next {
		switch pa.TypeCode {
		case packet.MultiProtocolReachNLRICode:
			s.multiProtocolUpdate(path, pa.Value.(packet.MultiProtocolReachNLRI))
		case packet.MultiProtocolUnreachNLRICode:
			s.multiProtocolWithdraw(path, pa.Value.(packet.MultiProtocolUnreachNLRI))
		}
	}
}

func (s *establishedState) newRoutePath() *route.Path {
	return &route.Path{
		Type: route.BGPPathType,
		BGPPath: &route.BGPPath{
			Source: s.fsm.peer.addr,
			EBGP:   s.fsm.peer.localASN != s.fsm.peer.peerASN,
		},
	}
}

func (s *establishedState) multiProtocolUpdate(path *route.Path, nlri packet.MultiProtocolReachNLRI) {
	path.BGPPath.NextHop = nlri.NextHop

	for _, pfx := range nlri.Prefixes {
		s.fsm.adjRIBIn.AddPath(pfx, path)
	}
}

func (s *establishedState) multiProtocolWithdraw(path *route.Path, nlri packet.MultiProtocolUnreachNLRI) {
	for _, pfx := range nlri.Prefixes {
		s.fsm.adjRIBIn.RemovePath(pfx, path)
	}
}

func (s *establishedState) processAttributes(attrs *packet.PathAttribute, path *route.Path) {
	for pa := attrs; pa != nil; pa = pa.Next {
		switch pa.TypeCode {
		case packet.OriginAttr:
			path.BGPPath.Origin = pa.Value.(uint8)
		case packet.LocalPrefAttr:
			path.BGPPath.LocalPref = pa.Value.(uint32)
		case packet.MEDAttr:
			path.BGPPath.MED = pa.Value.(uint32)
		case packet.NextHopAttr:
			path.BGPPath.NextHop = pa.Value.(bnet.IP)
		case packet.ASPathAttr:
			path.BGPPath.ASPath = pa.Value.(types.ASPath)
			path.BGPPath.ASPathLen = path.BGPPath.ASPath.Length()
		case packet.AggregatorAttr:
			aggr := pa.Value.(types.Aggregator)
			path.BGPPath.Aggregator = &aggr
		case packet.AtomicAggrAttr:
			path.BGPPath.AtomicAggregate = true
		case packet.CommunitiesAttr:
			path.BGPPath.Communities = pa.Value.([]uint32)
		case packet.LargeCommunitiesAttr:
			path.BGPPath.LargeCommunities = pa.Value.([]types.LargeCommunity)
		case packet.OriginatorIDAttr:
			path.BGPPath.OriginatorID = pa.Value.(uint32)
		case packet.ClusterListAttr:
			path.BGPPath.ClusterList = pa.Value.([]uint32)
		default:
			unknownAttr := s.processUnknownAttribute(pa)
			if unknownAttr != nil {
				path.BGPPath.UnknownAttributes = append(path.BGPPath.UnknownAttributes, *unknownAttr)
			}
		}
	}
}

func (s *establishedState) processUnknownAttribute(attr *packet.PathAttribute) *types.UnknownPathAttribute {
	if !attr.Transitive {
		return nil
	}

	u := &types.UnknownPathAttribute{
		Transitive: true,
		Optional:   attr.Optional,
		Partial:    attr.Partial,
		TypeCode:   attr.TypeCode,
		Value:      attr.Value.([]byte),
	}

	return u
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
