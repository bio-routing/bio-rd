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
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	log "github.com/sirupsen/logrus"
	"github.com/taktv6/tflow2/convert"
)

type router struct {
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
}

type neighbor struct {
	localAS  uint32
	peerAS   uint32
	address  [16]byte
	routerID uint32
	fsm      *FSM
	opt      *packet.DecodeOptions
}

func (r *router) serve() {
	for {
		msg, err := recvBMPMsg(r.con)
		if err != nil {
			log.Errorf("Unable to get message: %v", err)
			return
		}

		bmpMsg, err := bmppkt.Decode(msg)
		if err != nil {
			log.Errorf("Unable to decode BMP message: %v", err)
			fmt.Printf("msg: %v\n", msg)
			return
		}

		switch bmpMsg.MsgType() {
		case bmppkt.PeerUpNotificationType:
			r.processPeerUpNotification(bmpMsg.(*bmppkt.PeerUpNotification))
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
		log.Errorf("Received route monitoring message for non-existent neighbor %v on %s", msg.PerPeerHeader.PeerAddress, r.address.String())
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
			logMsg += fmt.Sprintf(" sysName.: %s", string(tlv.Information))
		}
	}

	log.Info(logMsg)
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

	log.Warning(logMsg)

	r.con.Close()
	for n := range r.neighbors {
		delete(r.neighbors, n)
	}
}

func (r *router) processPeerDownNotification(msg *bmppkt.PeerDownNotification) {
	r.neighborsMu.Lock()
	defer r.neighborsMu.Unlock()

	if _, ok := r.neighbors[msg.PerPeerHeader.PeerAddress]; !ok {
		log.Warningf("Received peer down notification for %v: Peer doesn't exist.", msg.PerPeerHeader.PeerAddress)
		return
	}

	delete(r.neighbors, msg.PerPeerHeader.PeerAddress)
}

func (r *router) processPeerUpNotification(msg *bmppkt.PeerUpNotification) {
	r.neighborsMu.Lock()
	defer r.neighborsMu.Unlock()

	if _, ok := r.neighbors[msg.PerPeerHeader.PeerAddress]; ok {
		log.Warningf("Received peer up notification for %v: Peer exists already.", msg.PerPeerHeader.PeerAddress)
		return
	}

	sentOpen, err := packet.DecodeOpenMsg(bytes.NewBuffer(msg.SentOpenMsg[19:]))
	if err != nil {
		log.Warningf("Unable to decode sent open message sent from %v to %v: %v", r.address.String(), msg.PerPeerHeader.PeerAddress, err)
		return
	}

	recvOpen, err := packet.DecodeOpenMsg(bytes.NewBuffer(msg.ReceivedOpenMsg[19:]))
	if err != nil {
		log.Warningf("Unable to decode received open message sent from %v to %v: %v", msg.PerPeerHeader.PeerAddress, r.address.String(), err)
		return
	}

	localAS := uint32(sentOpen.ASN)

	// TODO: Get 32bit ASN from OPEN message

	addrLen := 4
	for i := 0; i < 12; i++ {
		if msg.PerPeerHeader.PeerAddress[i] == 0 {
			continue
		}
		addrLen = 16
		break
	}

	peerAddress, err := bnet.IPFromBytes(msg.PerPeerHeader.PeerAddress[16-addrLen:])
	if err != nil {
		log.Warningf("Unable to convert peer address %v: %v", msg.PerPeerHeader.PeerAddress, err)
		return
	}
	localAddress, err := bnet.IPFromBytes(msg.LocalAddress[16-addrLen:])
	if err != nil {
		log.Warningf("Unable to convert peer address %v: %v", msg.PerPeerHeader.PeerAddress, err)
		return
	}

	fsm := &FSM{
		peer: &peer{
			peerASN:  msg.PerPeerHeader.PeerAS,
			localASN: localAS,
			ipv4:     &peerAddressFamily{},
			ipv6:     &peerAddressFamily{},
		},
		supports4OctetASN: true,
		supportsAddPathRX: true,
	}

	caps := getCaps(sentOpen.OptParams)
	for _, cap := range caps {
		switch cap.Code {
		case packet.AddPathCapabilityCode:
			addPathCap := cap.Value.(packet.AddPathCapability)
			f := fsm.addressFamily(addPathCap.AFI, addPathCap.SAFI)
			switch addPathCap.SendReceive {
			case packet.AddPathReceive:
				f.addPathRXConfigured = true
			case packet.AddPathSend:
				f.addPathTXConfigured = true
			case packet.AddPathSendReceive:
				f.addPathRXConfigured = true
				f.addPathTXConfigured = true
			}
		case packet.ASN4CapabilityCode:
			asn4Cap := cap.Value.(packet.ASN4Capability)
			localAS = asn4Cap.ASN4
			// TODO: Make 4Byte ASN configurable
		case packet.MultiProtocolCapabilityCode:
			mpCap := cap.Value.(packet.MultiProtocolCapability)
			f := fsm.addressFamily(mpCap.AFI, mpCap.SAFI)
			f.multiProtocol = true
		}
	}

	rtNeighbor := &routingtable.Neighbor{
		Address:      peerAddress,
		LocalAddress: localAddress,
		Type:         route.BGPPathType,
		IBGP:         msg.PerPeerHeader.PeerAS == localAS,
	}

	fsm.ipv4Unicast = newFSMAddressFamily(packet.IPv4AFI, packet.UnicastSAFI, &peerAddressFamily{
		rib:          r.rib4,
		importFilter: filter.NewAcceptAllFilter(),
		exportFilter: filter.NewDrainFilter(),
	}, fsm)
	fsm.ipv4Unicast.init(rtNeighbor)

	fsm.ipv6Unicast = newFSMAddressFamily(packet.IPv6AFI, packet.UnicastSAFI, &peerAddressFamily{
		rib:          r.rib6,
		importFilter: filter.NewAcceptAllFilter(),
		exportFilter: filter.NewDrainFilter(),
	}, fsm)
	fsm.ipv6Unicast.init(rtNeighbor)

	fsm.state = newOpenSentState(fsm)
	openSent := fsm.state.(*openSentState)
	openSent.openMsgReceived(recvOpen)

	fsm.state = newEstablishedState(fsm)
	n := &neighbor{
		localAS:  localAS,
		peerAS:   msg.PerPeerHeader.PeerAS,
		address:  msg.PerPeerHeader.PeerAddress,
		routerID: recvOpen.BGPIdentifier,
		fsm:      fsm,
		opt:      fsm.decodeOptions(),
	}

	r.neighbors[msg.PerPeerHeader.PeerAddress] = n
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
