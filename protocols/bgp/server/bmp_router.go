package server

import (
	"bytes"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	bmppkt "github.com/bio-routing/bio-rd/protocols/bmp/packet"
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

		fmt.Printf("%v\n", bmpMsg)

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

	bgpUpdate, err := packet.DecodeUpdateMsg(bytes.NewBuffer(msg.BGPUpdate[19:]), uint16(len(msg.BGPUpdate[19:])), nil)
	if err != nil {
		log.Errorf("Unable to decode BGP update message from %v on %s: %v", msg.PerPeerHeader.PeerAddress, r.address.String(), err)
	}

	n := r.neighbors[msg.PerPeerHeader.PeerAddress]
	s := n.fsm.state.(*establishedState)
	s.update(bgpUpdate)
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

	fmt.Printf("msg.SentOpenMsg[19:]: %v\n", msg.SentOpenMsg[19:])
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

	fsm := &FSM{}
	fsm.state = newEstablishedState(fsm)
	n := &neighbor{
		localAS:  localAS,
		peerAS:   msg.PerPeerHeader.PeerAS,
		address:  msg.PerPeerHeader.PeerAddress,
		routerID: recvOpen.BGPIdentifier,
		fsm:      fsm,
	}

	r.neighbors[msg.PerPeerHeader.PeerAddress] = n
}
