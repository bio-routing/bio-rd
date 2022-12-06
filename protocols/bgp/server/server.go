package server

import (
	"fmt"
	"net"

	"github.com/bio-routing/bio-rd/routingtable/adjRIBOut"
	"github.com/bio-routing/bio-rd/routingtable/filter"

	"github.com/bio-routing/bio-rd/routingtable/adjRIBIn"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/metrics"
	"github.com/bio-routing/bio-rd/util/log"
	bnetutils "github.com/bio-routing/bio-rd/util/net"
)

const (
	BGPVersion = 4
)

type bgpServer struct {
	addListeners      chan listener
	acceptedListeners []listener
	acceptCh          chan net.Conn
	peers             *peerManager
	routerID          uint32
	metrics           *metricsService
	logger            log.LoggerInterface
}

type BGPServer interface {
	RouterID() uint32
	AddListener(l ...net.Listener) error
	AddListenerFromAddrString(addr ...string) error
	Start() error
	AddPeer(PeerConfig) error
	GetPeerConfig(*bnet.IP) *PeerConfig
	DisposePeer(*bnet.IP)
	GetPeers() []*bnet.IP
	GetPeerStatus(*bnet.IP) string
	Metrics() (*metrics.BGPMetrics, error)
	GetRIBIn(peerIP *bnet.IP, afi uint16, safi uint8) *adjRIBIn.AdjRIBIn
	GetRIBOut(peerIP *bnet.IP, afi uint16, safi uint8) *adjRIBOut.AdjRIBOut
	ConnectMockPeer(peer PeerConfig, con net.Conn)
	ReplaceImportFilterChain(peer *bnet.IP, c filter.Chain) error
	ReplaceExportFilterChain(peer *bnet.IP, c filter.Chain) error
}

// NewBGPServer creates a new instance of bgpServer
func NewBGPServer(routerID uint32) BGPServer {
	return newBGPServer(routerID)
}

func newBGPServer(routerID uint32) *bgpServer {
	server := &bgpServer{
		peers:             newPeerManager(),
		routerID:          routerID,
		addListeners:      make(chan listener, 256),
		acceptedListeners: make([]listener, 0),
		logger: log.GetLogger().WithFields(log.Fields{
			"router_id": bnet.IPv4(routerID).String(),
		}),
	}

	server.metrics = &metricsService{server}

	return server
}

func (b *bgpServer) RouterID() uint32 {
	return b.routerID
}

// GetPeers gets a list of all peers
func (b *bgpServer) GetPeers() []*bnet.IP {
	ret := make([]*bnet.IP, 0)

	for _, p := range b.peers.list() {
		ret = append(ret, p.addr)
	}

	return ret
}

type listener interface {
	net.Listener
	setTCPMD5(net.IP, string) error
}

type dummyListener struct {
	net.Listener
	logger log.LoggerInterface
}

func (d *dummyListener) setTCPMD5(net.IP, string) error {
	d.logger.Debug("setTCPMD5 called on dummyListener, ignoring...")

	return nil
}

func (b *bgpServer) AddListener(l ...net.Listener) error {
	for _, l := range l {
		if ll, ok := l.(listener); ok {
			b.addListeners <- ll
		} else {
			d := &dummyListener{
				Listener: l,
				logger:   b.logger}
			b.addListeners <- d
		}
	}

	return nil
}

func (b *bgpServer) AddListenerFromAddrString(addrs ...string) error {
	for _, addr := range addrs {
		l, err := NewTCPListener(addr)
		if err != nil {
			return fmt.Errorf("failed to start TCPListener for %s: %w", addr, err)
		}

		if err := b.AddListener(l); err != nil {
			return fmt.Errorf("failed to add listener: %w", err)
		}
	}

	return nil
}

func (b *bgpServer) accept(addr listener, acceptCh chan net.Conn) {
	for {
		conn, err := addr.Accept()

		if err != nil {
			b.logger.Errorf("failed to accept connection: %v", err)
			continue
		}

		acceptCh <- conn
	}
}

func (b *bgpServer) Start() error {
	b.acceptCh = make(chan net.Conn, 4096)

	go b.incomingConnectionWorker()

	go func() {
		for addr := range b.addListeners {
			go b.accept(addr, b.acceptCh)
			b.acceptedListeners = append(b.acceptedListeners, addr)
		}
	}()

	return nil
}

// ReplaceImportFilterChain replaces a peers import filter
func (b *bgpServer) ReplaceImportFilterChain(peerIP *bnet.IP, c filter.Chain) error {
	p := b.peers.get(peerIP)
	if p == nil {
		return fmt.Errorf("peer %q not found", peerIP.String())
	}

	p.replaceImportFilterChain(c)
	return nil
}

// ReplaceExportFilterChain replaces a peers import filter
func (b *bgpServer) ReplaceExportFilterChain(peerIP *bnet.IP, c filter.Chain) error {
	p := b.peers.get(peerIP)
	if p == nil {
		return fmt.Errorf("peer %q not found", peerIP.String())
	}

	p.replaceExportFilterChain(c)
	return nil
}

func (b *bgpServer) GetRIBIn(peerIP *bnet.IP, afi uint16, safi uint8) *adjRIBIn.AdjRIBIn {
	p := b.peers.get(peerIP)
	if p == nil {
		return nil
	}

	if len(p.fsms) != 1 {
		return nil
	}

	fsm := p.fsms[0]
	f := fsm.addressFamily(afi, safi)
	if f == nil {
		return nil
	}

	return f.adjRIBIn.(*adjRIBIn.AdjRIBIn)
}

func (b *bgpServer) GetRIBOut(peerIP *bnet.IP, afi uint16, safi uint8) *adjRIBOut.AdjRIBOut {
	p := b.peers.get(peerIP)
	if p == nil {
		return nil
	}

	if len(p.fsms) != 1 {
		return nil
	}

	fsm := p.fsms[0]
	f := fsm.addressFamily(afi, safi)
	if f == nil {
		return nil
	}

	return f.adjRIBOut.(*adjRIBOut.AdjRIBOut)
}

func (b *bgpServer) incomingConnectionWorker() {
	for {
		c := <-b.acceptCh

		peerAddr, _ := bnetutils.BIONetIPFromAddr(c.RemoteAddr().String())
		peer := b.peers.get(peerAddr.Dedup())
		if peer == nil {
			c.Close()
			b.logger.WithFields(log.Fields{
				"source": c.RemoteAddr(),
			}).Info("TCP connection from unknown source")
			continue
		}

		b.logger.WithFields(log.Fields{
			"source": c.RemoteAddr(),
		}).Info("Incoming TCP connection")

		b.logger.WithFields(log.Fields{
			"peer": peerAddr,
		}).Debug("Sending incoming TCP connection to fsm for peer")
		fsm := NewActiveFSM(peer)
		fsm.state = newActiveState(fsm)
		fsm.startConnectRetryTimer()

		peer.fsmsMu.Lock()
		peer.fsms = append(peer.fsms, fsm)
		peer.fsmsMu.Unlock()

		go fsm.run()
		fsm.conCh <- c
	}
}

func (b *bgpServer) ConnectMockPeer(peer PeerConfig, con net.Conn) {
	acceptCh := make(chan net.Conn, 4096)
	b.acceptCh = acceptCh
	go b.incomingConnectionWorker()

	b.acceptCh <- con
}

func (b *bgpServer) AddPeer(c PeerConfig) error {
	c.LocalAddress = c.LocalAddress.Dedup()
	c.PeerAddress = c.PeerAddress.Dedup()

	peer, err := newPeer(c, b)
	if err != nil {
		return err
	}

	if c.AuthenticationKey != "" {
		for _, l := range b.acceptedListeners {
			err = l.setTCPMD5(c.PeerAddress.ToNetIP(), c.AuthenticationKey)
			if err != nil {
				return fmt.Errorf("unable to set TCP MD5 secret: %w", err)
			}
		}
	}

	peer.routerID = c.RouterID
	b.peers.add(peer)
	if !c.Passive {
		peer.Start()
	}

	b.logger.WithFields(log.Fields{
		"peer_address":  c.PeerAddress,
		"local_address": c.LocalAddress,
		"peer_as":       c.PeerAS,
		"local_as":      c.LocalAS,
	}).Infof("Added BGP peer")

	return nil
}

// GetPeerConfig gets a BGP peer by its address
func (b *bgpServer) GetPeerConfig(addr *bnet.IP) *PeerConfig {
	p := b.peers.get(addr)
	if p != nil {
		return p.config
	}

	return nil
}

// GetPeerStatus gets the status of a BGP peer by its address
func (b *bgpServer) GetPeerStatus(addr *bnet.IP) string {
	p := b.peers.get(addr)
	if p != nil {
		p.fsms[0].stateMu.RLock()
		defer p.fsms[0].stateMu.RUnlock()

		return stateName(p.fsms[0].state)
	}

	return ""
}

func (b *bgpServer) DisposePeer(addr *bnet.IP) {
	p := b.peers.get(addr)
	if p == nil {
		return
	}

	b.logger.Infof("disposing BGP session with %s", addr.String())
	p.stop()
	b.peers.remove(addr)
}

func (b *bgpServer) Metrics() (*metrics.BGPMetrics, error) {
	if b.metrics == nil {
		return nil, fmt.Errorf("server not started yet")
	}

	return b.metrics.metrics(), nil
}
