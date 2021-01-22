package server

import (
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
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

type BMPServerInterface interface {
	GetRouter(rtr string) RouterInterface
	GetRouters() []RouterInterface
}

// BMPServer represents a BMP server
type BMPServer struct {
	routers         map[string]*Router
	routersMu       sync.RWMutex
	ribClients      map[string]map[afiClient]struct{}
	keepalivePeriod time.Duration
	metrics         *bmpMetricsService
}

type afiClient struct {
	afi    uint8
	client routingtable.RouteTableClient
}

// NewServer creates a new BMP server
func NewServer(keepalivePeriod time.Duration) *BMPServer {
	b := &BMPServer{
		routers:         make(map[string]*Router),
		ribClients:      make(map[string]map[afiClient]struct{}),
		keepalivePeriod: keepalivePeriod,
	}

	b.metrics = &bmpMetricsService{b}
	return b
}

func conString(host string, port uint16) string {
	return fmt.Sprintf("%s:%d", host, port)
}

// AddRouter adds a router to which we connect with BMP
func (b *BMPServer) AddRouter(addr net.IP, port uint16) {
	r := newRouter(addr, port)
	b.addRouter(r)

	go func(r *Router) {
		for {
			select {
			case <-r.stop:
				log.WithFields(log.Fields{
					"component": "bmp_server",
					"address":   conString(r.address.String(), r.port),
				}).Info("Stop event: Stopping reconnect routine")
				return
			case <-r.reconnectTimer.C:
				log.WithFields(log.Fields{
					"component": "bmp_server",
					"address":   conString(r.address.String(), r.port),
				}).Info("Reconnect timer expired: Establishing connection")
			}

			c, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", r.address.String(), r.port), r.dialTimeout)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"component": "bmp_server",
					"address":   conString(r.address.String(), r.port),
				}).Info("Unable to connect to BMP router")
				if r.reconnectTime == 0 {
					r.reconnectTime = r.reconnectTimeMin
				} else if r.reconnectTime < r.reconnectTimeMax {
					r.reconnectTime *= 2
				}
				r.reconnectTimer = time.NewTimer(time.Second * time.Duration(r.reconnectTime))
				continue
			}

			if b.keepalivePeriod != 0 {
				err = c.(*net.TCPConn).SetKeepAlivePeriod(b.keepalivePeriod)
				if err != nil {
					log.WithError(err).Error("Unable to set keepalive period")
					return
				}
			}

			atomic.StoreUint32(&r.established, 1)
			r.reconnectTime = r.reconnectTimeMin
			r.reconnectTimer = time.NewTimer(time.Second * time.Duration(r.reconnectTime))
			log.WithFields(log.Fields{
				"component": "bmp_server",
				"address":   conString(r.address.String(), r.port),
			}).Info("Connected")

			err = r.serve(c)
			atomic.StoreUint32(&r.established, 0)
			if err != nil {
				r.logger.WithFields(log.Fields{
					"component": "bmp_server",
					"address":   conString(r.address.String(), r.port),
				}).WithError(err).Error("r.serve() failed")
			} else {
				r.logger.WithFields(log.Fields{
					"component": "bmp_server",
					"address":   conString(r.address.String(), r.port),
				}).Info("r.Serve returned without error. Stopping reconnect routine")
				return
			}
		}
	}(r)
}

func (b *BMPServer) addRouter(r *Router) {
	b.routersMu.Lock()
	defer b.routersMu.Unlock()

	b.routers[fmt.Sprintf("%s", r.address.String())] = r
}

func (b *BMPServer) deleteRouter(addr net.IP) {
	b.routersMu.Lock()
	defer b.routersMu.Unlock()

	delete(b.routers, addr.String())
}

// RemoveRouter removes a BMP monitored router
func (b *BMPServer) RemoveRouter(addr net.IP) {
	id := addr.String()
	r := b.routers[id]
	close(r.stop)

	b.deleteRouter(addr)
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

// GetRouters gets all routers
func (b *BMPServer) GetRouters() []RouterInterface {
	b.routersMu.RLock()
	defer b.routersMu.RUnlock()

	r := make([]RouterInterface, 0, len(b.routers))
	for name := range b.routers {
		r = append(r, b.routers[name])
	}

	return r
}

// GetRouter gets a router
func (b *BMPServer) GetRouter(name string) RouterInterface {
	b.routersMu.RLock()
	defer b.routersMu.RUnlock()

	if _, ok := b.routers[name]; ok {
		return b.routers[name]
	}

	return nil
}

// Metrics gets BMP server metrics
func (b *BMPServer) Metrics() (*metrics.BMPMetrics, error) {
	if b.metrics == nil {
		return nil, fmt.Errorf("Server not started yet")
	}

	return b.metrics.metrics(), nil
}
