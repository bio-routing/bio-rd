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
	"github.com/bio-routing/bio-rd/util/log"
	"github.com/bio-routing/tflow2/convert"
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
	listener        net.Listener
	adjRIBInFactory adjRIBInFactoryI
	acceptAny       bool
	ignorePeerASNs  []uint32
}

type BMPServerConfig struct {
	KeepalivePeriod time.Duration
	AcceptAny       bool
	IgnorePeerASNs  []uint32
}

type afiClient struct {
	afi    uint8
	client routingtable.RouteTableClient
}

// NewServer creates a new BMP server
func NewServer(cfg BMPServerConfig) *BMPServer {
	b := &BMPServer{
		routers:         make(map[string]*Router),
		ribClients:      make(map[string]map[afiClient]struct{}),
		keepalivePeriod: cfg.KeepalivePeriod,
		adjRIBInFactory: adjRIBInFactory{},
		acceptAny:       cfg.AcceptAny,
		ignorePeerASNs:  cfg.IgnorePeerASNs,
	}

	b.metrics = &bmpMetricsService{b}
	return b
}

func NewServerWithAdjRIBInFactory(cfg BMPServerConfig, adjRIBInFactory adjRIBInFactoryI) *BMPServer {
	b := NewServer(cfg)
	b.adjRIBInFactory = adjRIBInFactory

	return b
}

// Listen starts a listener for routers to start the BMP connection
// the listener needs to be closed by calling Close() on the BMPServer
func (b *BMPServer) Listen(addr string) error {
	tcp, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"component": "bmp_server",
			"address":   addr,
		}).Info("Unable to resolve address")
		return err
	}
	l, err := net.ListenTCP("tcp", tcp)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"component": "bmp_server",
			"address":   addr,
		}).Infof("Unable to listen on %s", tcp.String())
		return err
	}
	b.listener = l
	return nil
}

// Serve accepts all incoming connections for the BMP server until Close() is called.
func (b *BMPServer) Serve() error {
	l := b.listener
	for {
		c, err := l.Accept()
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"component": "bmp_server",
				"address":   b.LocalAddr(),
			}).Infof("Unable to accept on %s", b.LocalAddr())
			return err
		}

		// Do we know this router from configuration or have seen it before?
		r := b.getRouter(c.RemoteAddr().(*net.TCPAddr).IP.String())
		dynamic := false
		if r == nil {
			// If we don't know this router and don't accept connection from anyone, we drop it
			if !b.acceptAny {
				continue
			}

			dynamic = true
			tcpRemoteAddr := c.RemoteAddr().(*net.TCPAddr)
			localPort := uint16(b.LocalAddr().(*net.TCPAddr).Port)
			err := b.AddRouter(tcpRemoteAddr.IP, localPort, true, dynamic)
			if err != nil {
				log.Errorf("failed to add router: %v", err)
				continue
			}

			r = b.getRouter(tcpRemoteAddr.IP.String())
			// Potentially we could have already removed this router if it were connecting twice and did not offer an init message
			if r == nil {
				continue
			}
		}

		if b.validateConnection(r, c) {
			go b.handleConnection(c, r, true, dynamic)
		}
	}
}

// LocalAddr returns the local address the server is listening to.
func (b *BMPServer) LocalAddr() net.Addr {
	if b.listener != nil {
		return b.listener.Addr()
	}
	return nil
}

func (b *BMPServer) validateConnection(r *Router, c net.Conn) bool {
	if !r.passive {
		log.WithFields(log.Fields{
			"component":      "bmp_server",
			"remote_address": c.RemoteAddr(),
		}).Error("dropping unconfigured connection")
		c.Close()
		return false
	}

	if atomic.LoadUint32(&r.established) == 1 {
		log.WithFields(log.Fields{
			"component":      "bmp_server",
			"remote_address": c.RemoteAddr(),
			"router":         r.Name(),
		}).Error("router already connected, dropping connection")
		c.Close()
		return false
	}

	return true
}

func (b *BMPServer) handleConnection(c net.Conn, r *Router, passive bool, dynamic bool) error {
	if b.keepalivePeriod != 0 {
		if err := c.(*net.TCPConn).SetKeepAlive(true); err != nil {
			log.WithError(err).Error("Unable to enable keepalive")
			return err
		}
		if err := c.(*net.TCPConn).SetKeepAlivePeriod(b.keepalivePeriod); err != nil {
			log.WithError(err).Error("Unable to set keepalive period")
			return err
		}
	}

	log.WithFields(log.Fields{
		"component": "bmp_server",
		"address":   c.RemoteAddr().String(),
		"passive":   passive,
	}).Info("Connected")

	atomic.StoreUint32(&r.established, 1)
	err := r.serve(c)
	atomic.StoreUint32(&r.established, 0)

	if err != nil {
		if dynamic && r.counters.initiationMessages == 0 {
			defer b.RemoveRouter(r.address)

			log.WithFields(log.Fields{
				"component": "bmp_server",
				"address":   c.RemoteAddr().String(),
				"router":    r.Name(),
				"passive":   passive,
			}).WithError(err).Info("r.serve() failed for dynamic peer")
			return err

		} else {
			log.WithFields(log.Fields{
				"component": "bmp_server",
				"address":   c.RemoteAddr().String(),
				"router":    r.Name(),
				"passive":   passive,
			}).WithError(err).Error("r.serve() failed")
		}
		return err
	}

	return nil
}

func conString(host string, port uint16) string {
	return fmt.Sprintf("%s:%d", host, port)
}

// AddRouter adds a router to which we connect with BMP
func (b *BMPServer) AddRouter(addr net.IP, port uint16, passive bool, dynamic bool) error {
	b.routersMu.Lock()
	defer b.routersMu.Unlock()

	return b._addRouter(addr, port, passive, dynamic)
}

func (b *BMPServer) _addRouter(addr net.IP, port uint16, passive bool, dynamic bool) error {
	r := newRouter(addr, port, passive, b.adjRIBInFactory, b.ignorePeerASNs)
	if _, exists := b.routers[r.address.String()]; exists {
		return fmt.Errorf("router %s already configured,", r.address.String())
	}

	b.routers[r.address.String()] = r

	if r.passive {
		// router initializes the connection
		return nil
	}

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

			r.reconnectTime = r.reconnectTimeMin
			r.reconnectTimer = time.NewTimer(time.Second * time.Duration(r.reconnectTime))

			b.handleConnection(c, r, false, dynamic)
		}
	}(r)

	return nil
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

func (b *BMPServer) getRouter(name string) *Router {
	b.routersMu.RLock()
	defer b.routersMu.RUnlock()

	if _, ok := b.routers[name]; ok {
		return b.routers[name]
	}

	return nil
}

func recvBMPMsg(c net.Conn) (msg []byte, err error) {
	buffer := make([]byte, defaultBufferLen)
	_, err = io.ReadFull(c, buffer[0:bmppkt.MinLen])
	if err != nil {
		return nil, fmt.Errorf("Read failed: %w", err)
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
		return nil, fmt.Errorf("Read failed: %w", err)
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
	if r := b.getRouter(name); r != nil {
		return r
	}
	return nil
}

// Metrics gets BMP server metrics
func (b *BMPServer) Metrics() (*metrics.BMPMetrics, error) {
	if b.metrics == nil {
		return nil, fmt.Errorf("server not started yet")
	}

	return b.metrics.metrics(), nil
}

// Close tears down all open connections
func (b *BMPServer) Close() error {
	return b.listener.Close()
}
