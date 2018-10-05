package server

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	log "github.com/sirupsen/logrus"
)

const (
	defaultBufferLen = 4096
)

type BMPServer struct {
	routers   map[string]*router
	routersMu sync.RWMutex
}

func NewServer() *BMPServer {
	return &BMPServer{
		routers: make(map[string]*router),
	}
}

func (b *BMPServer) AddRouter(addr net.IP, port uint16, rib4 *locRIB.LocRIB, rib6 *locRIB.LocRIB) {
	r := &router{
		address:        addr,
		port:           port,
		reconnectTime:  0,
		reconnectTimer: time.NewTimer(time.Duration(0)),
		rib4:           rib4,
		rib6:           rib6,
	}

	b.routersMu.Lock()
	b.routers[fmt.Sprintf("%s:%d", r.address.String(), r.port)] = r
	b.routersMu.Unlock()

	go func(r *router) {
		for {
			<-r.reconnectTimer.C
			c, err := net.Dial("tcp", fmt.Sprintf("%s:%d", r.address.String(), r.port))
			if err != nil {
				log.Infof("Unable to connect to BMP router: %v", err)
				if r.reconnectTime == 0 {
					r.reconnectTime = 30
				} else if r.reconnectTime < 720 {
					r.reconnectTime *= 2
				}
				r.reconnectTimer = time.NewTimer(time.Second * time.Duration(r.reconnectTime))
				continue
			}

			r.reconnectTime = 0
			r.con = c
			r.serve()
		}
	}(r)
}

func recvMsg(c net.Conn) (msg []byte, err error) {
	buffer := make([]byte, defaultBufferLen)
	_, err = io.ReadFull(c, buffer[0:packet.MinLen])
	if err != nil {
		return nil, fmt.Errorf("Read failed: %v", err)
	}

	l := int(buffer[1])*256*256*256 + int(buffer[2])*256*256 + int(buffer[3])*256 + int(buffer[4])
	if l > defaultBufferLen {
		tmp := buffer
		buffer = make([]byte, l)
		copy(buffer, tmp)
	}

	toRead := l
	_, err = io.ReadFull(c, buffer[packet.MinLen:toRead])
	if err != nil {
		return nil, fmt.Errorf("Read failed: %v", err)
	}

	return buffer, nil
}