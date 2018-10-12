package server

import (
	"fmt"
	"net"
	"time"

	"github.com/bio-routing/bio-rd/protocols/bmp/packet"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
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

		// TODO: Finish implementation
		fmt.Printf("%v\n", bmpMsg)
	}

}
