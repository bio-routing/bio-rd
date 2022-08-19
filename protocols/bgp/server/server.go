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
	listenAddrs []string
	listeners   []*TCPListener
	acceptCh    chan net.Conn
	peers       *peerManager
	routerID    uint32
	metrics     *metricsService
}

type BGPServer interface {
	RouterID() uint32
	Start() error
	AddPeer(PeerConfig) error
	GetPeerConfig(*bnet.IP) *PeerConfig
	DisposePeer(*bnet.IP)
	GetPeers() []*bnet.IP
	Metrics() (*metrics.BGPMetrics, error)
	GetRIBIn(peerIP *bnet.IP, afi uint16, safi uint8) *adjRIBIn.AdjRIBIn
	GetRIBOut(peerIP *bnet.IP, afi uint16, safi uint8) *adjRIBOut.AdjRIBOut
	ConnectMockPeer(peer PeerConfig, con net.Conn)
	ReplaceImportFilterChain(peer *bnet.IP, c filter.Chain) error
	ReplaceExportFilterChain(peer *bnet.IP, c filter.Chain) error
}

// NewBGPServer creates a new instance of bgpServer
func NewBGPServer(routerID uint32, addrs []string) BGPServer {
	return newBGPServer(routerID, addrs)
}

func newBGPServer(routerID uint32, addrs []string) *bgpServer {
	server := &bgpServer{
		peers:       newPeerManager(),
		routerID:    routerID,
		listenAddrs: addrs,
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

func (b *bgpServer) Start() error {
	if len(b.listenAddrs) > 0 {
		acceptCh := make(chan net.Conn, 4096)
		for _, addr := range b.listenAddrs {
			l, err := NewTCPListener(addr, acceptCh)
			if err != nil {
				return fmt.Errorf("failed to start TCPListener for %s: %w", addr, err)
			}
			b.listeners = append(b.listeners, l)
		}
		b.acceptCh = acceptCh

		go b.incomingConnectionWorker()
	}

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
			log.WithFields(log.Fields{
				"source": c.RemoteAddr(),
			}).Info("TCP connection from unknown source")
			continue
		}

		log.WithFields(log.Fields{
			"source": c.RemoteAddr(),
		}).Info("Incoming TCP connection")

		log.WithFields(log.Fields{
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
		for _, l := range b.listeners {
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

	log.WithFields(log.Fields{
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

func (b *bgpServer) DisposePeer(addr *bnet.IP) {
	p := b.peers.get(addr)
	if p == nil {
		return
	}

	log.Infof("disposing BGP session with %s", addr.String())
	p.stop()
	b.peers.remove(addr)
}

func (b *bgpServer) Metrics() (*metrics.BGPMetrics, error) {
	if b.metrics == nil {
		return nil, fmt.Errorf("server not started yet")
	}

	return b.metrics.metrics(), nil
}
