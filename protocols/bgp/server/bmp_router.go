package server

import (
	"bytes"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/api"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	bmppkt "github.com/bio-routing/bio-rd/protocols/bmp/packet"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/adjRIBIn"
	"github.com/bio-routing/bio-rd/routingtable/adjRIBOut"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
	"github.com/bio-routing/tflow2/convert"
)

// Router represents a BMP enabled route in BMP context
type Router struct {
	name             string
	nameMu           sync.RWMutex
	address          net.IP
	port             uint16
	con              net.Conn
	established      uint32
	reconnectTimeMin int
	reconnectTimeMax int
	reconnectTime    int
	dialTimeout      time.Duration
	reconnectTimer   *time.Timer
	vrfRegistry      *vrf.VRFRegistry
	neighborManager  *neighborManager
	logger           *log.Logger
	runMu            sync.Mutex
	stop             chan struct{}

	ribClients   map[afiClient]struct{}
	ribClientsMu sync.Mutex

	counters routerCounters
}

type routerCounters struct {
	routeMonitoringMessages      uint64
	statisticsReportMessages     uint64
	peerDownNotificationMessages uint64
	peerUpNotificationMessages   uint64
	initiationMessages           uint64
	terminationMessages          uint64
	routeMirroringMessages       uint64
}

type neighbor struct {
	vrfID       uint64
	peerAddress [16]byte
	localAS     uint32
	peerAS      uint32
	routerID    uint32
	fsm         *FSM
	opt         *packet.DecodeOptions
}

func newRouter(addr net.IP, port uint16) *Router {
	return &Router{
		address:          addr,
		port:             port,
		reconnectTimeMin: 30,  // Suggested by RFC 7854
		reconnectTimeMax: 720, // Suggested by RFC 7854
		reconnectTimer:   time.NewTimer(time.Duration(0)),
		dialTimeout:      time.Second * 5,
		vrfRegistry:      vrf.NewVRFRegistry(),
		neighborManager:  newNeighborManager(),
		logger:           log.New(),
		stop:             make(chan struct{}),
		ribClients:       make(map[afiClient]struct{}),
	}
}

// GetVRF get's a VRF
func (r *Router) GetVRF(rd uint64) *vrf.VRF {
	return r.vrfRegistry.GetVRFByRD(rd)
}

// GetVRFs gets all VRFs
func (r *Router) GetVRFs() []*vrf.VRF {
	return r.vrfRegistry.List()
}

// Name gets a routers name
func (r *Router) Name() string {
	r.nameMu.RLock()
	defer r.nameMu.RUnlock()
	return r.name
}

// Address gets a routers address
func (r *Router) Address() net.IP {
	return r.address
}

func (r *Router) serve(con net.Conn) {
	r.con = con
	r.runMu.Lock()
	defer r.con.Close()
	defer r.runMu.Unlock()

	for {
		select {
		case <-r.stop:
			return
		default:
		}

		msg, err := recvBMPMsg(r.con)
		if err != nil {
			r.logger.Errorf("Unable to get message: %v", err)
			return
		}

		r.processMsg(msg)
	}
}

// GetNeighborSessions returns all neighbors as API session objects
func (r *Router) GetNeighborSessions() []*api.Session {
	sessions := make([]*api.Session, 0)

	for _, neigh := range r.neighborManager.list() {
		estSince := neigh.fsm.establishedTime.Unix()
		if estSince < 0 {
			estSince = 0
		}

		// for now get this from adjRibIn/adjRibOut, can be replaced when we
		// bmp gets its own bgpSrv or Router gets the bmpMetricsService
		var routesReceived, routesSent uint64
		for _, afi := range []uint16{packet.IPv4AFI, packet.IPv6AFI} {
			ribIn, err1 := r.GetNeighborRIBIn(neigh.fsm.peer.addr, afi, packet.UnicastSAFI)
			if err1 == nil {
				routesReceived += uint64(ribIn.RouteCount())
			}

			// adjRIBOut might not work properly with BMP, keeping it here for when it will
			ribOut, err2 := r.GetNeighborRIBOut(neigh.fsm.peer.addr, afi, packet.UnicastSAFI)
			if err2 == nil {
				routesSent += uint64(ribOut.RouteCount())
			}
		}
		session := &api.Session{
			LocalAddress:    neigh.fsm.peer.localAddr.ToProto(),
			NeighborAddress: neigh.fsm.peer.addr.ToProto(),
			LocalAsn:        neigh.localAS,
			PeerAsn:         neigh.peerAS,
			Status:          stateToProto(neigh.fsm.state),
			Stats: &api.SessionStats{
				RoutesReceived: routesReceived,
				RoutesExported: routesSent,
			},
			EstablishedSince: uint64(estSince),
		}
		sessions = append(sessions, session)
	}
	return sessions
}

func (r *Router) processMsg(msg []byte) {
	bmpMsg, err := bmppkt.Decode(msg)
	if err != nil {
		r.logger.Errorf("Unable to decode BMP message: %v", err)
		return
	}

	switch bmpMsg.MsgType() {
	case bmppkt.PeerUpNotificationType:
		err = r.processPeerUpNotification(bmpMsg.(*bmppkt.PeerUpNotification))
		if err != nil {
			r.logger.Errorf("Unable to process peer up notification: %v", err)
		}
	case bmppkt.PeerDownNotificationType:
		r.processPeerDownNotification(bmpMsg.(*bmppkt.PeerDownNotification))
	case bmppkt.InitiationMessageType:
		r.processInitiationMsg(bmpMsg.(*bmppkt.InitiationMessage))
	case bmppkt.TerminationMessageType:
		r.processTerminationMsg(bmpMsg.(*bmppkt.TerminationMessage))
		return
	case bmppkt.RouteMonitoringType:
		r.processRouteMonitoringMsg(bmpMsg.(*bmppkt.RouteMonitoringMsg))
	case bmppkt.RouteMirroringMessageType:
		atomic.AddUint64(&r.counters.routeMirroringMessages, 1)
	}
}

func (r *Router) processRouteMonitoringMsg(msg *bmppkt.RouteMonitoringMsg) {
	atomic.AddUint64(&r.counters.routeMonitoringMessages, 1)

	n := r.neighborManager.getNeighbor(msg.PerPeerHeader.PeerDistinguisher, msg.PerPeerHeader.PeerAddress)
	if n == nil {
		r.logger.Errorf("Received route monitoring message for non-existent neighbor %d/%v on %s", msg.PerPeerHeader.PeerDistinguisher, msg.PerPeerHeader.PeerAddress, r.address.String())
		return
	}

	s := n.fsm.state.(*establishedState)
	opt := s.fsm.decodeOptions()
	opt.Use32BitASN = !msg.PerPeerHeader.GetAFlag()
	s.msgReceived(msg.BGPUpdate, opt)
}

func (r *Router) processInitiationMsg(msg *bmppkt.InitiationMessage) {
	atomic.AddUint64(&r.counters.initiationMessages, 1)

	r.nameMu.Lock()
	defer r.nameMu.Unlock()

	const (
		stringType   = 0
		sysDescrType = 1
		sysNameType  = 2
	)

	logMsg := fmt.Sprintf("Received initiation message from %s:", r.address.String())

	for _, tlv := range msg.TLVs {
		switch tlv.InformationType {
		case stringType:
			logMsg += fmt.Sprintf(" Message: %q", string(tlv.Information))
		case sysDescrType:
			logMsg += fmt.Sprintf(" sysDescr.: %s", string(tlv.Information))
		case sysNameType:
			r.name = string(tlv.Information)
			logMsg += fmt.Sprintf(" sysName.: %s", string(tlv.Information))
		}
	}

	r.logger.Info(logMsg)
}

func (r *Router) getNeighborAddressFamily(addr *bnet.IP, afi uint16, safi uint8) (*fsmAddressFamily, error) {
	if safi != packet.UnicastSAFI {
		return nil, fmt.Errorf("Unsupported safi, only unicast is supported")
	}

	for _, neigh := range r.neighborManager.list() {
		if *neigh.fsm.peer.addr == *addr {
			af := neigh.fsm.addressFamily(afi, safi)
			if af == nil {
				return nil, fmt.Errorf("Address family not available")
			}
			return af, nil
		}
	}

	return nil, fmt.Errorf("Could not find neighbor with ip %s", addr.String())
}

// GetNeighborRIBIn returns the AdjRIBIn of a BMP neighbor
func (r *Router) GetNeighborRIBIn(addr *bnet.IP, afi uint16, safi uint8) (*adjRIBIn.AdjRIBIn, error) {
	neighAF, err := r.getNeighborAddressFamily(addr, afi, safi)
	if err != nil {
		return nil, errors.Wrap(err, "Could not get RIBIn")
	}
	if neighAF.adjRIBIn == nil {
		return nil, fmt.Errorf("RIBIn not available")
	}
	return neighAF.adjRIBIn.(*adjRIBIn.AdjRIBIn), nil
}

// GetNeighborRIBOut returns the AdjRIBOut of a BMP neighbor
func (r *Router) GetNeighborRIBOut(addr *bnet.IP, afi uint16, safi uint8) (*adjRIBOut.AdjRIBOut, error) {
	neighAF, err := r.getNeighborAddressFamily(addr, afi, safi)
	if err != nil {
		return nil, errors.Wrap(err, "Could not get RIBIn")
	}
	if neighAF.adjRIBOut == nil {
		return nil, fmt.Errorf("RIBOut not available")
	}
	return neighAF.adjRIBOut.(*adjRIBOut.AdjRIBOut), nil
}

func (r *Router) processTerminationMsg(msg *bmppkt.TerminationMessage) {
	const (
		stringType = 0
		reasonType = 1

		adminDown     = 0
		unspecReason  = 1
		outOfRes      = 2
		redundantCon  = 3
		permAdminDown = 4
	)

	atomic.AddUint64(&r.counters.terminationMessages, 1)
	logMsg := fmt.Sprintf("Received termination message from %s: ", r.address.String())
	for _, tlv := range msg.TLVs {
		switch tlv.InformationType {
		case stringType:
			logMsg += fmt.Sprintf("Message: %q", string(tlv.Information))
		case reasonType:
			reason := convert.Uint16b(tlv.Information[:1])
			switch reason {
			case adminDown:
				logMsg += "Session administratively down"
			case unspecReason:
				logMsg += "Unspecified reason"
			case outOfRes:
				logMsg += "Out of resources"
			case redundantCon:
				logMsg += "Redundant connection"
			case permAdminDown:
				logMsg += "Session permanently administratively closed"
			}
		}
	}

	r.logger.Warning(logMsg)

	r.con.Close()
	r.neighborManager.disposeAll()
}

func (r *Router) processPeerDownNotification(msg *bmppkt.PeerDownNotification) {
	r.logger.WithFields(log.Fields{
		"address":            r.address.String(),
		"router":             r.name,
		"peer_distinguisher": msg.PerPeerHeader.PeerDistinguisher,
		"peer_address":       addrToNetIP(msg.PerPeerHeader.PeerAddress).String(),
	}).Infof("peer down notification received")
	atomic.AddUint64(&r.counters.peerDownNotificationMessages, 1)

	err := r.neighborManager.neighborDown(msg.PerPeerHeader.PeerDistinguisher, msg.PerPeerHeader.PeerAddress)
	if err != nil {
		r.logger.Errorf("Failed to process peer down notification: %v", err)
	}
}

func (r *Router) processPeerUpNotification(msg *bmppkt.PeerUpNotification) error {
	atomic.AddUint64(&r.counters.peerUpNotificationMessages, 1)
	r.logger.WithFields(log.Fields{
		"address":            r.address.String(),
		"router":             r.name,
		"peer_distinguisher": msg.PerPeerHeader.PeerDistinguisher,
		"peer_address":       addrToNetIP(msg.PerPeerHeader.PeerAddress).String(),
	}).Infof("peer up notification received")

	if len(msg.SentOpenMsg) < packet.MinOpenLen {
		return fmt.Errorf("Received peer up notification for %v: Invalid sent open message: %v", msg.PerPeerHeader.PeerAddress, msg.SentOpenMsg)
	}

	sentOpen, err := packet.DecodeOpenMsg(bytes.NewBuffer(msg.SentOpenMsg[packet.HeaderLen:]))
	if err != nil {
		return errors.Wrapf(err, "Unable to decode sent open message sent from %v to %v", r.address.String(), msg.PerPeerHeader.PeerAddress)
	}

	if len(msg.ReceivedOpenMsg) < packet.MinOpenLen {
		return fmt.Errorf("Received peer up notification for %v: Invalid received open message: %v", msg.PerPeerHeader.PeerAddress, msg.ReceivedOpenMsg)
	}

	recvOpen, err := packet.DecodeOpenMsg(bytes.NewBuffer(msg.ReceivedOpenMsg[packet.HeaderLen:]))
	if err != nil {
		return errors.Wrapf(err, "Unable to decode received open message sent from %v to %v", msg.PerPeerHeader.PeerAddress, r.address.String())
	}

	addrLen := net.IPv4len
	if msg.PerPeerHeader.GetIPVersion() == 6 {
		addrLen = net.IPv6len
	}

	// bnet.IPFromBytes can only fail if length of argument is not 4 or 16. However, length is ensured here.
	peerAddress, _ := bnet.IPFromBytes(msg.PerPeerHeader.PeerAddress[16-addrLen:])
	localAddress, _ := bnet.IPFromBytes(msg.LocalAddress[16-addrLen:])

	fsm := &FSM{
		isBMP: true,
		peer: &peer{
			routerID:  sentOpen.BGPIdentifier,
			addr:      peerAddress.Dedup(),
			localAddr: localAddress.Dedup(),
			peerASN:   msg.PerPeerHeader.PeerAS,
			localASN:  uint32(sentOpen.ASN),
			ipv4:      &peerAddressFamily{},
			ipv6:      &peerAddressFamily{},
			vrf:       r.vrfRegistry.CreateVRFIfNotExists(fmt.Sprintf("%d", msg.PerPeerHeader.PeerDistinguisher), msg.PerPeerHeader.PeerDistinguisher),
		},
	}

	fsm.peer.fsms = []*FSM{
		fsm,
	}

	fsm.peer.configureBySentOpen(sentOpen)

	rib4, found := fsm.peer.vrf.RIBByName("inet.0")
	if !found {
		return fmt.Errorf("Unable to get inet RIB")
	}
	fsm.ipv4Unicast = newFSMAddressFamily(packet.IPv4AFI, packet.UnicastSAFI, &peerAddressFamily{
		rib:               rib4,
		importFilterChain: filter.NewAcceptAllFilterChain(),
	}, fsm)
	fsm.ipv4Unicast.bmpInit()

	rib6, found := fsm.peer.vrf.RIBByName("inet6.0")
	if !found {
		return fmt.Errorf("Unable to get inet6 RIB")
	}

	fsm.ipv6Unicast = newFSMAddressFamily(packet.IPv6AFI, packet.UnicastSAFI, &peerAddressFamily{
		rib:               rib6,
		importFilterChain: filter.NewAcceptAllFilterChain(),
	}, fsm)
	fsm.ipv6Unicast.bmpInit()

	fsm.state = newOpenSentState(fsm)
	openSent := fsm.state.(*openSentState)
	openSent.openMsgReceived(recvOpen)

	fsm.state = newEstablishedState(fsm)
	n := &neighbor{
		vrfID:       msg.PerPeerHeader.PeerDistinguisher,
		localAS:     fsm.peer.localASN,
		peerAS:      msg.PerPeerHeader.PeerAS,
		peerAddress: msg.PerPeerHeader.PeerAddress,
		routerID:    recvOpen.BGPIdentifier,
		fsm:         fsm,
		opt:         fsm.decodeOptions(),
	}

	err = r.neighborManager.addNeighbor(n)
	if err != nil {
		return errors.Wrap(err, "Unable to add neighbor")
	}

	r.ribClientsMu.Lock()
	defer r.ribClientsMu.Unlock()
	n.registerClients(r.ribClients)

	return nil
}

func (n *neighbor) registerClients(clients map[afiClient]struct{}) {
	for ac := range clients {
		if ac.afi == packet.IPv4AFI {
			n.fsm.ipv4Unicast.adjRIBIn.Register(ac.client)
		}
		if ac.afi == packet.IPv6AFI {
			n.fsm.ipv6Unicast.adjRIBIn.Register(ac.client)
		}
	}
}

func (p *peer) configureBySentOpen(msg *packet.BGPOpen) {
	capsList := getCaps(msg.OptParams)
	for _, caps := range capsList {
		for _, cap := range caps {
			switch cap.Code {
			case packet.AddPathCapabilityCode:
				addPathCap := cap.Value.(packet.AddPathCapability)
				for _, addPathCapTuple := range addPathCap {
					peerFamily := p.addressFamily(addPathCapTuple.AFI, addPathCapTuple.SAFI)
					if peerFamily == nil {
						continue
					}
					switch addPathCapTuple.SendReceive {
					case packet.AddPathSend:
						peerFamily.addPathSend = routingtable.ClientOptions{
							MaxPaths: 10,
						}
					case packet.AddPathReceive:
						peerFamily.addPathReceive = true
					case packet.AddPathSendReceive:
						peerFamily.addPathReceive = true
						peerFamily.addPathSend = routingtable.ClientOptions{
							MaxPaths: 10,
						}
					}
				}
			}
		}
	}
}

func getCaps(optParams []packet.OptParam) []packet.Capabilities {
	res := make([]packet.Capabilities, 0)
	for _, optParam := range optParams {
		if optParam.Type != packet.CapabilitiesParamType {
			continue
		}

		res = append(res, optParam.Value.(packet.Capabilities))
	}

	return res
}

func addrToNetIP(a [16]byte) net.IP {
	for i := 0; i < 12; i++ {
		if a[i] != 0 {
			return net.IP(a[:])
		}
	}

	return net.IP(a[12:])
}
