package server

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/bio-routing/bio-rd/protocols/bgp/metrics"
	bmppkt "github.com/bio-routing/bio-rd/protocols/bmp/packet"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/tflow2/convert"
	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"
)

const (
	defaultBufferLen = 4096
)

// BMPServer represents a BMP server
type BMPServer struct {
	routers    map[string]*Router
	routersMu  sync.RWMutex
	ribClients map[string]map[afiClient]struct{}
	gloablMu   sync.RWMutex
	metrics    *bmpMetricsService
}

type afiClient struct {
	afi    uint8
	client routingtable.RouteTableClient
}

// NewServer creates a new BMP server
func NewServer() *BMPServer {
	b := &BMPServer{
		routers:    make(map[string]*Router),
		ribClients: make(map[string]map[afiClient]struct{}),
	}

	b.metrics = &bmpMetricsService{b}
	return b
}

// AddRouter adds a router to which we connect with BMP
func (b *BMPServer) AddRouter(addr net.IP, port uint16) {
	b.gloablMu.Lock()
	defer b.gloablMu.Unlock()

	r := newRouter(addr, port)
	b.addRouter(r)

	go func(r *Router) {
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

func (b *BMPServer) addRouter(r *Router) {
	b.routers[fmt.Sprintf("%s", r.address.String())] = r
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

func (b *BMPServer) getRouters() []*Router {
	b.routersMu.RLock()
	defer b.routersMu.RUnlock()

	ret := make([]*Router, 0, len(b.routers))
	for r := range b.routers {
		ret = append(ret, b.routers[r])
	}

	return ret
}

func recvBMPMsg(c net.Conn) (msg []byte, err error) {
	buffer := make([]byte, defaultBufferLen)
	_, err = io.ReadFull(c, buffer[0:bmppkt.MinLen])
	if err != nil {
		return nil, errors.Wrap(err, "Read failed")
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
		return nil, errors.Wrap(err, "Read failed")
	}

	return buffer[0:toRead], nil
}

func (b *BMPServer) GetRouters() []*Router {
	b.routersMu.RLock()
	defer b.routersMu.RUnlock()

	r := make([]*Router, 0, len(b.routers))
	for name := range b.routers {
		r = append(r, b.routers[name])
	}

	return r
}

func (b *BMPServer) GetRouter(name string) *Router {
	b.routersMu.RLock()
	defer b.routersMu.RUnlock()

	for x := range b.routers {
		if x != name {
			continue
		}

		return b.routers[x]
	}

	return nil
}

func (b *BMPServer) Metrics() (*metrics.BMPMetrics, error) {
	if b.metrics == nil {
		return nil, fmt.Errorf("Server not started yet")
	}

	return b.metrics.metrics(), nil
}
