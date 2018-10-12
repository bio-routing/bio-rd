package server

import (
	"bytes"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/bio-routing/bio-rd/protocols/bmp/packet"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/adjRIBIn"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	"github.com/bio-routing/tflow2/convert"
	log "github.com/sirupsen/logrus"

	bgppkt "github.com/bio-routing/bio-rd/protocols/bgp/packet"
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
	ipv4     *adjRIBIn.AdjRIBIn
	ipv6     *adjRIBIn.AdjRIBIn
}

func (r *router) serve() {
	for {
		msg, err := recvMsg(r.con)
		if err != nil {
			log.Errorf("Unable to get message: %v", err)
			return
		}

		bmpMsg, err := packet.Decode(msg)
		if err != nil {
			log.Errorf("Unable to decode BMP message: %v", err)
			return
		}

		fmt.Printf("%v\n", bmpMsg)

		switch bmpMsg.MsgType() {
		case packet.PeerUpNotificationType:
			r.processPeerUpNotification(bmpMsg.(*packet.PeerUpNotification))
		case packet.PeerDownNotificationType:
			r.processPeerDownNotification(bmpMsg.(*packet.PeerDownNotification))
		case packet.InitiationMessageType:
			r.processInitiationMsg(bmpMsg.(*packet.InitiationMessage))
		case packet.TerminationMessageType:
			r.processTerminationMsg(bmpMsg.(*packet.TerminationMessage))
			return

		}

	}
}

func (r *router) processRouteMonitoringMsg(msg *packet.RouteMonitoringMsg) {
	r.neighborsMu.Lock()
	defer r.neighborsMu.Unlock()

	if _, ok := r.neighbors[msg.PerPeerHeader.PeerAddress]; !ok {
		log.Errorf("Received route monitoring message for non-existent neighbor %v on %s", msg.PerPeerHeader.PeerAddress, r.address.String())
		return
	}

	bgpUpdate, err := bgppkt.DecodeUpdateMsg(bytes.NewBuffer(msg.BGPUpdate), uint16(len(msg.BGPUpdate)), nil)
	if err != nil {
		log.Errorf("Unable to decode BGP update message from %v on %s: %v", msg.PerPeerHeader.PeerAddress, r.address.String(), err)
	}

	afi, safi := bgpUpdate.AddressFamily()
	if safi != bgppkt.UnicastSAFI {
		// only unicast support, so other SAFIs are ignored
		return
	}

	switch afi {
	case bgppkt.IPv4AFI:

	case bgppkt.IPv6AFI:

	}
}

func (r *router) processInitiationMsg(msg *packet.InitiationMessage) {
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

func (r *router) processTerminationMsg(msg *packet.TerminationMessage) {
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

func (r *router) processPeerDownNotification(msg *packet.PeerDownNotification) {
	r.neighborsMu.Lock()
	defer r.neighborsMu.Unlock()

	if _, ok := r.neighbors[msg.PerPeerHeader.PeerAddress]; !ok {
		log.Warningf("Received peer down notification for %v: Peer doesn't exist.", msg.PerPeerHeader.PeerAddress)
		return
	}

	delete(r.neighbors, msg.PerPeerHeader.PeerAddress)
}

func (r *router) processPeerUpNotification(msg *packet.PeerUpNotification) {
	r.neighborsMu.Lock()
	defer r.neighborsMu.Unlock()

	if _, ok := r.neighbors[msg.PerPeerHeader.PeerAddress]; ok {
		log.Warningf("Received peer up notification for %v: Peer exists already.", msg.PerPeerHeader.PeerAddress)
		return
	}

	sentOpen, err := bgppkt.DecodeOpenMsg(bytes.NewBuffer(msg.SentOpenMsg))
	if err != nil {
		log.Warningf("Unable to decode sent open message sent from %v to %v: %v", r.address.String(), msg.PerPeerHeader.PeerAddress, err)
		return
	}

	recvOpen, err := bgppkt.DecodeOpenMsg(bytes.NewBuffer(msg.ReceivedOpenMsg))
	if err != nil {
		log.Warningf("Unable to decode received open message sent from %v to %v: %v", msg.PerPeerHeader.PeerAddress, r.address.String(), err)
		return
	}

	localAS := uint32(sentOpen.ASN)
	// TODO: Get 32bit ASN from OPEN message

	r.neighbors[msg.PerPeerHeader.PeerAddress] = &neighbor{
		localAS:  localAS,
		peerAS:   msg.PerPeerHeader.PeerAS,
		address:  msg.PerPeerHeader.PeerAddress,
		routerID: recvOpen.BGPIdentifier,
		ipv4:     adjRIBIn.New(nil, routingtable.NewContributingASNs(), recvOpen.BGPIdentifier, 0),
		ipv6:     adjRIBIn.New(nil, routingtable.NewContributingASNs(), recvOpen.BGPIdentifier, 0),
	}
}
