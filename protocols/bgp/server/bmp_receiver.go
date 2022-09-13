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

type BMPReceiverInterface interface {
	GetRouter(rtr string) RouterInterface
	GetRouters() []RouterInterface
}

// BMPReceiver represents a BMP receiver
type BMPReceiver struct {
	routers          map[string]*Router
	routersMu        sync.RWMutex
	ribClients       map[string]map[afiClient]struct{}
	keepalivePeriod  time.Duration
	metrics          *bmpMetricsService
	listener         net.Listener
	adjRIBInFactory  adjRIBInFactoryI
	acceptAny        bool
	ignorePeerASNs   []uint32
	ignorePrePolicy  bool
	ignorePostPolicy bool
}

type BMPReceiverConfig struct {
	KeepalivePeriod  time.Duration
	AcceptAny        bool
	IgnorePeerASNs   []uint32
	IgnorePrePolicy  bool
	IgnorePostPolicy bool
}

type afiClient struct {
	afi    uint8
	client routingtable.RouteTableClient
}

// NewBMPReceiver creates a new BMP receiver
func NewBMPReceiver(cfg BMPReceiverConfig) *BMPReceiver {
	b := &BMPReceiver{
		routers:          make(map[string]*Router),
		ribClients:       make(map[string]map[afiClient]struct{}),
		keepalivePeriod:  cfg.KeepalivePeriod,
		adjRIBInFactory:  adjRIBInFactory{},
		acceptAny:        cfg.AcceptAny,
		ignorePeerASNs:   cfg.IgnorePeerASNs,
		ignorePrePolicy:  cfg.IgnorePrePolicy,
		ignorePostPolicy: cfg.IgnorePostPolicy,
	}

	b.metrics = &bmpMetricsService{b}
	return b
}

func NewBMPReceiverWithAdjRIBInFactory(cfg BMPReceiverConfig, adjRIBInFactory adjRIBInFactoryI) *BMPReceiver {
	b := NewBMPReceiver(cfg)
	b.adjRIBInFactory = adjRIBInFactory

	return b
}

// Listen starts a listener for routers to start the BMP connection
// the listener needs to be closed by calling Close() on the BMPReceiver
func (b *BMPReceiver) Listen(addr string) error {
	tcp, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"component": "bmp_receiver",
			"address":   addr,
		}).Info("Unable to resolve address")
		return err
	}
	l, err := net.ListenTCP("tcp", tcp)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"component": "bmp_receiver",
			"address":   addr,
		}).Infof("Unable to listen on %s", tcp.String())
		return err
	}
	b.listener = l
	return nil
}

// Serve accepts all incoming connections for the BMP receiver until Close() is called.
func (b *BMPReceiver) Serve() error {
	l := b.listener
	for {
		c, err := l.Accept()
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"component": "bmp_receiver",
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

// LocalAddr returns the local address the receiver is listening to.
func (b *BMPReceiver) LocalAddr() net.Addr {
	if b.listener != nil {
		return b.listener.Addr()
	}
	return nil
}

func (b *BMPReceiver) validateConnection(r *Router, c net.Conn) bool {
	if !r.passive {
		log.WithFields(log.Fields{
			"component":      "bmp_receiver",
			"remote_address": c.RemoteAddr(),
		}).Error("dropping unconfigured connection")
		c.Close()
		return false
	}

	if atomic.LoadUint32(&r.established) == 1 {
		log.WithFields(log.Fields{
			"component":      "bmp_receiver",
			"remote_address": c.RemoteAddr(),
			"router":         r.Name(),
		}).Error("router already connected, dropping connection")
		c.Close()
		return false
	}

	return true
}

func (b *BMPReceiver) handleConnection(c net.Conn, r *Router, passive bool, dynamic bool) error {
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
		"component": "bmp_receiver",
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
				"component": "bmp_receiver",
				"address":   c.RemoteAddr().String(),
				"router":    r.Name(),
				"passive":   passive,
			}).WithError(err).Info("r.serve() failed for dynamic peer")
			return err

		} else {
			log.WithFields(log.Fields{
				"component": "bmp_receiver",
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
func (b *BMPReceiver) AddRouter(addr net.IP, port uint16, passive bool, dynamic bool) error {
	b.routersMu.Lock()
	defer b.routersMu.Unlock()

	return b._addRouter(addr, port, passive, dynamic)
}

func (b *BMPReceiver) _addRouter(addr net.IP, port uint16, passive bool, dynamic bool) error {
	rCfg := RouterConfig{
		Passive:          passive,
		IgnorePeerASNs:   b.ignorePeerASNs,
		IgnorePrePolicy:  b.ignorePrePolicy,
		IgnorePostPolicy: b.ignorePostPolicy,
	}
	r := newRouter(addr, port, b.adjRIBInFactory, rCfg)
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
					"component": "bmp_receiver",
					"address":   conString(r.address.String(), r.port),
				}).Info("Stop event: Stopping reconnect routine")
				return
			case <-r.reconnectTimer.C:
				log.WithFields(log.Fields{
					"component": "bmp_receiver",
					"address":   conString(r.address.String(), r.port),
				}).Info("Reconnect timer expired: Establishing connection")
			}

			c, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", r.address.String(), r.port), r.dialTimeout)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"component": "bmp_receiver",
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

func (b *BMPReceiver) deleteRouter(addr net.IP) {
	b.routersMu.Lock()
	defer b.routersMu.Unlock()

	delete(b.routers, addr.String())
}

// RemoveRouter removes a BMP monitored router
func (b *BMPReceiver) RemoveRouter(addr net.IP) {
	id := addr.String()
	r := b.routers[id]
	close(r.stop)

	b.deleteRouter(addr)
}

func (b *BMPReceiver) getRouters() []*Router {
	b.routersMu.RLock()
	defer b.routersMu.RUnlock()

	ret := make([]*Router, 0, len(b.routers))
	for r := range b.routers {
		ret = append(ret, b.routers[r])
	}

	return ret
}

func (b *BMPReceiver) getRouter(name string) *Router {
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
func (b *BMPReceiver) GetRouters() []RouterInterface {
	b.routersMu.RLock()
	defer b.routersMu.RUnlock()

	r := make([]RouterInterface, 0, len(b.routers))
	for name := range b.routers {
		r = append(r, b.routers[name])
	}

	return r
}

// GetRouter gets a router
func (b *BMPReceiver) GetRouter(name string) RouterInterface {
	if r := b.getRouter(name); r != nil {
		return r
	}
	return nil
}

// Metrics gets BMP receiver metrics
func (b *BMPReceiver) Metrics() (*metrics.BMPMetrics, error) {
	if b.metrics == nil {
		return nil, fmt.Errorf("BMP receiver not started yet")
	}

	return b.metrics.metrics(), nil
}

// Close tears down all open connections
func (b *BMPReceiver) Close() error {
	return b.listener.Close()
}
