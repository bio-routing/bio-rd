package server

import (
	"bytes"
	"fmt"
	"net"
	"sync"
	"time"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	bmppkt "github.com/bio-routing/bio-rd/protocols/bmp/packet"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	log "github.com/sirupsen/logrus"
	"github.com/taktv6/tflow2/convert"
)

type router struct {
	name             string
	address          net.IP
	port             uint16
	con              net.Conn
	reconnectTimeMin int
	reconnectTimeMax int
	reconnectTime    int
	reconnectTimer   *time.Timer
	rib4             *locRIB.LocRIB
	rib6             *locRIB.LocRIB
	neighbors        map[[16]byte]*neighbor
	neighborsMu      sync.Mutex
	logger           *log.Logger
	runMu            sync.Mutex
	stop             chan struct{}

	ribClients   map[afiClient]struct{}
	ribClientsMu sync.Mutex
}

type neighbor struct {
	localAS     uint32
	peerAS      uint32
	peerAddress [16]byte
	routerID    uint32
	fsm         *FSM
	opt         *packet.DecodeOptions
}

func newRouter(addr net.IP, port uint16, rib4 *locRIB.LocRIB, rib6 *locRIB.LocRIB) *router {
	return &router{
		address:          addr,
		port:             port,
		reconnectTimeMin: 30,  // Suggested by RFC 7854
		reconnectTimeMax: 720, // Suggested by RFC 7854
		reconnectTimer:   time.NewTimer(time.Duration(0)),
		rib4:             rib4,
		rib6:             rib6,
		neighbors:        make(map[[16]byte]*neighbor),
		logger:           log.New(),
		stop:             make(chan struct{}),
		ribClients:       make(map[afiClient]struct{}),
	}
}

func (r *router) subscribeRIBs(client routingtable.RouteTableClient, afi uint8) {
	ac := afiClient{
		afi:    afi,
		client: client,
	}

	r.ribClientsMu.Lock()
	defer r.ribClientsMu.Unlock()
	if _, ok := r.ribClients[ac]; ok {
		return
	}
	r.ribClients[ac] = struct{}{}

	r.neighborsMu.Lock()
	defer r.neighborsMu.Unlock()
	for _, n := range r.neighbors {
		if afi == packet.IPv4AFI {
			n.fsm.ipv4Unicast.adjRIBIn.Register(client)
		}
		if afi == packet.IPv6AFI {
			n.fsm.ipv6Unicast.adjRIBIn.Register(client)
		}
	}
}

func (r *router) unsubscribeRIBs(client routingtable.RouteTableClient, afi uint8) {
	ac := afiClient{
		afi:    afi,
		client: client,
	}

	r.ribClientsMu.Lock()
	defer r.ribClientsMu.Unlock()
	if _, ok := r.ribClients[ac]; !ok {
		return
	}
	delete(r.ribClients, ac)

	r.neighborsMu.Lock()
	defer r.neighborsMu.Unlock()
	for _, n := range r.neighbors {
		if !n.fsm.ribsInitialized {
			continue
		}
		if afi == packet.IPv4AFI {
			n.fsm.ipv4Unicast.adjRIBIn.Unregister(client)
		}
		if afi == packet.IPv6AFI {
			n.fsm.ipv6Unicast.adjRIBIn.Unregister(client)
		}
	}
}

func (r *router) serve(con net.Conn) {
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
		}
	}
}

func (r *router) processRouteMonitoringMsg(msg *bmppkt.RouteMonitoringMsg) {
	r.neighborsMu.Lock()
	defer r.neighborsMu.Unlock()

	if _, ok := r.neighbors[msg.PerPeerHeader.PeerAddress]; !ok {
		r.logger.Errorf("Received route monitoring message for non-existent neighbor %v on %s", msg.PerPeerHeader.PeerAddress, r.address.String())
		return
	}

	n := r.neighbors[msg.PerPeerHeader.PeerAddress]
	s := n.fsm.state.(*establishedState)
	s.msgReceived(msg.BGPUpdate, s.fsm.decodeOptions())
}

func (r *router) processInitiationMsg(msg *bmppkt.InitiationMessage) {
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

func (r *router) processTerminationMsg(msg *bmppkt.TerminationMessage) {
	const (
		stringType = 0
		reasonType = 1

		adminDown     = 0
		unspecReason  = 1
		outOfRes      = 2
		redundantCon  = 3
		permAdminDown = 4
	)

	logMsg := fmt.Sprintf("Received termination message from %s: ", r.address.String())
	for _, tlv := range msg.TLVs {
		switch tlv.InformationType {
		case stringType:
			logMsg += fmt.Sprintf("Message: %q", string(tlv.Information))
		case reasonType:
			reason := convert.Uint16b(tlv.Information[:2])
			switch reason {
			case adminDown:
				logMsg += "Session administratively down"
			case unspecReason:
				logMsg += "Unespcified reason"
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
	for n := range r.neighbors {
		r.peerDown(n)
	}
}

func (r *router) processPeerDownNotification(msg *bmppkt.PeerDownNotification) {
	r.neighborsMu.Lock()
	defer r.neighborsMu.Unlock()

	if _, ok := r.neighbors[msg.PerPeerHeader.PeerAddress]; !ok {
		r.logger.Warningf("Received peer down notification for %v: Peer doesn't exist.", msg.PerPeerHeader.PeerAddress)
		return
	}

	r.peerDown(msg.PerPeerHeader.PeerAddress)
}

func (r *router) peerDown(addr [16]byte) {
	if r.neighbors[addr].fsm != nil {
		if r.neighbors[addr].fsm.ipv4Unicast != nil {
			r.neighbors[addr].fsm.ipv4Unicast.bmpDispose()
		}

		if r.neighbors[addr].fsm.ipv6Unicast != nil {
			r.neighbors[addr].fsm.ipv6Unicast.bmpDispose()
		}
	}

	delete(r.neighbors, addr)
}

func (r *router) processPeerUpNotification(msg *bmppkt.PeerUpNotification) error {
	r.neighborsMu.Lock()
	defer r.neighborsMu.Unlock()

	if _, ok := r.neighbors[msg.PerPeerHeader.PeerAddress]; ok {
		return fmt.Errorf("Received peer up notification for %v: Peer exists already", msg.PerPeerHeader.PeerAddress)
	}

	if len(msg.SentOpenMsg) < packet.MinOpenLen {
		return fmt.Errorf("Received peer up notification for %v: Invalid sent open message: %v", msg.PerPeerHeader.PeerAddress, msg.SentOpenMsg)
	}

	sentOpen, err := packet.DecodeOpenMsg(bytes.NewBuffer(msg.SentOpenMsg[packet.HeaderLen:]))
	if err != nil {
		return fmt.Errorf("Unable to decode sent open message sent from %v to %v: %v", r.address.String(), msg.PerPeerHeader.PeerAddress, err)
	}

	if len(msg.ReceivedOpenMsg) < packet.MinOpenLen {
		return fmt.Errorf("Received peer up notification for %v: Invalid received open message: %v", msg.PerPeerHeader.PeerAddress, msg.ReceivedOpenMsg)
	}

	recvOpen, err := packet.DecodeOpenMsg(bytes.NewBuffer(msg.ReceivedOpenMsg[packet.HeaderLen:]))
	if err != nil {
		return fmt.Errorf("Unable to decode received open message sent from %v to %v: %v", msg.PerPeerHeader.PeerAddress, r.address.String(), err)
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
			addr:      peerAddress,
			localAddr: localAddress,
			peerASN:   msg.PerPeerHeader.PeerAS,
			localASN:  uint32(sentOpen.ASN),
			ipv4:      &peerAddressFamily{},
			ipv6:      &peerAddressFamily{},
		},
	}

	fsm.peer.configureBySentOpen(sentOpen)

	fsm.ipv4Unicast = newFSMAddressFamily(packet.IPv4AFI, packet.UnicastSAFI, &peerAddressFamily{
		rib:          r.rib4,
		importFilter: filter.NewAcceptAllFilter(),
	}, fsm)
	fsm.ipv4Unicast.bmpInit()

	fsm.ipv6Unicast = newFSMAddressFamily(packet.IPv6AFI, packet.UnicastSAFI, &peerAddressFamily{
		rib:          r.rib6,
		importFilter: filter.NewAcceptAllFilter(),
	}, fsm)
	fsm.ipv6Unicast.bmpInit()

	fsm.state = newOpenSentState(fsm)
	openSent := fsm.state.(*openSentState)
	openSent.openMsgReceived(recvOpen)

	fsm.state = newEstablishedState(fsm)
	n := &neighbor{
		localAS:     fsm.peer.localASN,
		peerAS:      msg.PerPeerHeader.PeerAS,
		peerAddress: msg.PerPeerHeader.PeerAddress,
		routerID:    recvOpen.BGPIdentifier,
		fsm:         fsm,
		opt:         fsm.decodeOptions(),
	}

	r.neighbors[msg.PerPeerHeader.PeerAddress] = n

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
	caps := getCaps(msg.OptParams)
	for _, cap := range caps {
		switch cap.Code {
		case packet.AddPathCapabilityCode:
			addPathCap := cap.Value.(packet.AddPathCapability)
			peerFamily := p.addressFamily(addPathCap.AFI, addPathCap.SAFI)
			if peerFamily == nil {
				continue
			}
			switch addPathCap.SendReceive {
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
		case packet.ASN4CapabilityCode:
			asn4Cap := cap.Value.(packet.ASN4Capability)
			p.localASN = asn4Cap.ASN4
		}
	}
}

func getCaps(optParams []packet.OptParam) packet.Capabilities {
	for _, optParam := range optParams {
		if optParam.Type != packet.CapabilitiesParamType {
			continue
		}

		return optParam.Value.(packet.Capabilities)
	}
	return nil
}
