package server

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	bmppkt "github.com/bio-routing/bio-rd/protocols/bmp/packet"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	log "github.com/sirupsen/logrus"
	"github.com/taktv6/tflow2/convert"
)

const (
	defaultBufferLen = 4096
)

// BMPServer represents a BMP server
type BMPServer struct {
	routers    map[string]*router
	ribClients map[string]map[afiClient]struct{}
	gloablMu   sync.RWMutex
}

type afiClient struct {
	afi    uint8
	client routingtable.RouteTableClient
}

// NewServer creates a new BMP server
func NewServer() *BMPServer {
	return &BMPServer{
		routers:    make(map[string]*router),
		ribClients: make(map[string]map[afiClient]struct{}),
	}
}

// SubscribeRIBs subscribes c for all RIB updates of router rtr
func (b *BMPServer) SubscribeRIBs(client routingtable.RouteTableClient, rtr net.IP, afi uint8) {
	b.gloablMu.Lock()
	defer b.gloablMu.Unlock()

	rtrStr := rtr.String()
	if _, ok := b.ribClients[rtrStr]; !ok {
		b.ribClients[rtrStr] = make(map[afiClient]struct{})
	}

	ac := afiClient{
		afi:    afi,
		client: client,
	}
	if _, ok := b.ribClients[rtrStr][ac]; ok {
		return
	}

	b.ribClients[rtrStr][ac] = struct{}{}

	if _, ok := b.routers[rtrStr]; !ok {
		return
	}

	b.routers[rtrStr].subscribeRIBs(client, afi)
}

// UnsubscribeRIBs unsubscribes client from RIBs of address family afi
func (b *BMPServer) UnsubscribeRIBs(client routingtable.RouteTableClient, rtr net.IP, afi uint8) {
	b.gloablMu.Lock()
	defer b.gloablMu.Unlock()

	rtrStr := rtr.String()
	if _, ok := b.ribClients[rtrStr]; !ok {
		return
	}

	ac := afiClient{
		afi:    afi,
		client: client,
	}
	if _, ok := b.ribClients[rtrStr][ac]; !ok {
		return
	}

	delete(b.ribClients[rtrStr], ac)
	b.routers[rtrStr].unsubscribeRIBs(client, afi)
}

// AddRouter adds a router to which we connect with BMP
func (b *BMPServer) AddRouter(addr net.IP, port uint16, rib4 *locRIB.LocRIB, rib6 *locRIB.LocRIB) {
	b.gloablMu.Lock()
	defer b.gloablMu.Unlock()

	r := newRouter(addr, port, rib4, rib6)
	b.routers[fmt.Sprintf("%s", r.address.String())] = r

	go func(r *router) {
		for {
			<-r.reconnectTimer.C
			c, err := net.Dial("tcp", fmt.Sprintf("%s:%d", r.address.String(), r.port))
			if err != nil {
				log.Infof("Unable to connect to BMP router: %v", err)
				if r.reconnectTime == 0 {
					r.reconnectTime = r.reconnectTimeMin
				} else if r.reconnectTime < r.reconnectTimeMax {
					r.reconnectTime *= 2
				}
				r.reconnectTimer = time.NewTimer(time.Second * time.Duration(r.reconnectTime))
				continue
			}

			r.reconnectTime = 0
			log.Infof("Connected to %s", r.address.String())
			r.serve(c)
		}
	}(r)
}

// RemoveRouter removes a BMP monitored router
func (b *BMPServer) RemoveRouter(addr net.IP, port uint16) {
	b.gloablMu.Lock()
	defer b.gloablMu.Unlock()

	id := addr.String()
	r := b.routers[id]
	r.stop <- struct{}{}
	delete(b.routers, id)
}

func recvBMPMsg(c net.Conn) (msg []byte, err error) {
	buffer := make([]byte, defaultBufferLen)
	_, err = io.ReadFull(c, buffer[0:bmppkt.MinLen])
	if err != nil {
		return nil, fmt.Errorf("Read failed: %v", err)
	}

	l := convert.Uint32b(buffer[1:5])
	if l > defaultBufferLen {
		tmp := buffer
		buffer = make([]byte, l)
		copy(buffer, tmp)
	}

	toRead := l
	_, err = io.ReadFull(c, buffer[bmppkt.MinLen:toRead])
	if err != nil {
		return nil, fmt.Errorf("Read failed: %v", err)
	}

	return buffer[0:toRead], nil
}
