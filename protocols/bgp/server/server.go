package server

import (
	"fmt"
	"net"

	"github.com/bio-routing/bio-rd/routingtable/adjRIBOut"

	"github.com/bio-routing/bio-rd/routingtable/adjRIBIn"

	"github.com/bio-routing/bio-rd/config"
	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/metrics"
	bnetutils "github.com/bio-routing/bio-rd/util/net"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	BGPVersion = 4
)

type bgpServer struct {
	listeners []*TCPListener
	acceptCh  chan net.Conn
	peers     *peerManager
	routerID  uint32
	localASN  uint32
	metrics   *metricsService
}

type BGPServer interface {
	RouterID() uint32
	Start(*config.Global) error
	AddPeer(config.Peer) error
	Metrics() (*metrics.BGPMetrics, error)
	GetRIBIn(peerIP bnet.IP, afi uint16, safi uint8) *adjRIBIn.AdjRIBIn
	GetRIBOut(peerIP bnet.IP, afi uint16, safi uint8) *adjRIBOut.AdjRIBOut
	ConnectMockPeer(peer config.Peer, con net.Conn)
}

// NewBgpServer creates a new instance of bgpServer
func NewBgpServer() BGPServer {
	return newBgpServer()
}

func newBgpServer() *bgpServer {
	server := &bgpServer{
		peers: newPeerManager(),
	}

	server.metrics = &metricsService{server}
	return server
}

func (b *bgpServer) RouterID() uint32 {
	return b.routerID
}

func (b *bgpServer) Start(c *config.Global) error {
	if err := c.SetDefaultGlobalConfigValues(); err != nil {
		return errors.Wrap(err, "Failed to load defaults")
	}

	log.Infof("ROUTER ID: %d\n", c.RouterID)
	b.routerID = c.RouterID
	b.localASN = c.LocalASN

	if c.Listen {
		acceptCh := make(chan net.Conn, 4096)
		for _, addr := range c.LocalAddressList {
			l, err := NewTCPListener(addr, c.Port, acceptCh)
			if err != nil {
				return errors.Wrapf(err, "Failed to start TCPListener for %s", addr.String())
			}
			b.listeners = append(b.listeners, l)
		}
		b.acceptCh = acceptCh

		go b.incomingConnectionWorker()
	}

	return nil
}

func (b *bgpServer) GetRIBIn(peerIP bnet.IP, afi uint16, safi uint8) *adjRIBIn.AdjRIBIn {
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

func (b *bgpServer) GetRIBOut(peerIP bnet.IP, afi uint16, safi uint8) *adjRIBOut.AdjRIBOut {
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
		peer := b.peers.get(peerAddr)
		if peer == nil {
			c.Close()
			log.WithFields(log.Fields{
				"source": c.RemoteAddr(),
			}).Warning("TCP connection from unknown source")
			continue
		}

		log.WithFields(log.Fields{
			"source": c.RemoteAddr(),
		}).Info("Incoming TCP connection")

		log.WithField("Peer", peerAddr).Debug("Sending incoming TCP connection to fsm for peer")
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

func (b *bgpServer) ConnectMockPeer(peer config.Peer, con net.Conn) {
	acceptCh := make(chan net.Conn, 4096)
	b.acceptCh = acceptCh
	go b.incomingConnectionWorker()

	b.acceptCh <- con
}

func (b *bgpServer) AddPeer(c config.Peer) error {
	peer, err := newPeer(c, b)
	if err != nil {
		return err
	}

	peer.routerID = c.RouterID
	b.peers.add(peer)
	if !c.Passive {
		peer.Start()
	}

	return nil
}

func (b *bgpServer) Metrics() (*metrics.BGPMetrics, error) {
	if b.metrics == nil {
		return nil, fmt.Errorf("Server not started yet")
	}

	return b.metrics.metrics(), nil
}
