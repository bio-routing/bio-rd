package server

import (
	"net"
	"time"

	"github.com/bio-routing/bio-rd/protocols/bmp/packet"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
	log "github.com/sirupsen/logrus"
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
	vrfs             []*vrf.VRF
	msgCounter       uint64
}

func (r *router) serve() {
	for {
		msg, err := recvMsg(r.con)
		if err != nil {
			log.Errorf("Unable to get message: %v", err)
			return
		}
		r.msgCounter++

		bmpMsg, err := packet.Decode(msg)
		if err != nil {
			log.Errorf("Unable to decode BMP message: %v", err)
			return
		}

		if r.msgCounter == 1 {
			if bmpMsg.MsgType() != packet.InitiationMessageType {
				log.Errorf("Invalid message type of first message (%d) for neighbor %s", bmpMsg.MsgType(), r.address.String())
				return
			}
		}

		switch bmpMsg.MsgType() {
		case packet.TerminationMessageType:
			log.Infof("BMP neighbor %s has sent a termination message", r.address.String())
			return
		case packet.RouteMonitoringType:
			r.processRouteMonitoringMsg(bmpMsg.(*packet.RouteMonitoringMsg))
			return
		}
	}
}

func (r *router) processRouteMonitoringMsg(m *packet.RouteMonitoringMsg) {

}
