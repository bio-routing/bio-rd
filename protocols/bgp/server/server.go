package server

import (
	"fmt"

	"github.com/bio-routing/bio-rd/net/tcp"
	"github.com/bio-routing/bio-rd/routingtable/adjRIBOut"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/routingtable/vrf"

	"github.com/bio-routing/bio-rd/routingtable/adjRIBIn"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/metrics"
	"github.com/bio-routing/bio-rd/util/log"
	bnetutils "github.com/bio-routing/bio-rd/util/net"
)

const (
	BGPVersion            = 4
	DefaultLocaPreference = 100
)

type BGPServerConfig struct {
	// Mandatory attributes
	RouterID         uint32
	ListenAddrsByVRF map[string][]string
	DefaultVRF       *vrf.VRF

	// Optional attributes
	DefaultLocalPreference *uint32
}

type bgpServer struct {
	config          BGPServerConfig
	listenerManager tcp.ListenerManagerI
	peers           *peerManager
	metrics         *metricsService
}

type BGPServer interface {
	RouterID() uint32
	Start()
	AddPeer(PeerConfig) error
	GetPeerConfig(*vrf.VRF, *bnet.IP) *PeerConfig
	DisposePeer(*vrf.VRF, *bnet.IP)
	GetPeers() []PeerKey
	Metrics() (*metrics.BGPMetrics, error)
	GetRIBIn(vrf *vrf.VRF, peerIP *bnet.IP, afi uint16, safi uint8) *adjRIBIn.AdjRIBIn
	GetRIBOut(vrf *vrf.VRF, peerIP *bnet.IP, afi uint16, safi uint8) *adjRIBOut.AdjRIBOut
	ReplaceImportFilterChain(vrf *vrf.VRF, peer *bnet.IP, c filter.Chain) error
	ReplaceExportFilterChain(vrf *vrf.VRF, peer *bnet.IP, c filter.Chain) error
	GetDefaultVRF() *vrf.VRF
	SetListenerManager(lm tcp.ListenerManagerI)
}

// NewBGPServer creates a new instance of BGPServer with the given BGPServerConfig
func NewBGPServer(config BGPServerConfig) BGPServer {
	if config.DefaultLocalPreference == nil {
		hundret := uint32(100)
		config.DefaultLocalPreference = &hundret
	}

	return newBGPServer(config)
}

func newBGPServer(config BGPServerConfig) *bgpServer {
	server := &bgpServer{
		config:          config,
		peers:           newPeerManager(),
		listenerManager: tcp.NewListenerManager(config.ListenAddrsByVRF),
	}

	server.metrics = &metricsService{server}
	return server
}

func (b *bgpServer) SetListenerManager(lm tcp.ListenerManagerI) {
	b.listenerManager = lm
}

func (b *bgpServer) GetDefaultVRF() *vrf.VRF {
	return b.config.DefaultVRF
}

func (b *bgpServer) RouterID() uint32 {
	return b.config.RouterID
}

// GetPeers gets a list of all peers
func (b *bgpServer) GetPeers() []PeerKey {
	ret := make([]PeerKey, 0)

	for _, p := range b.peers.list() {
		ret = append(ret, p.peerKey())
	}

	return ret
}

// ReplaceImportFilterChain replaces a peers import filter
func (b *bgpServer) ReplaceImportFilterChain(vrf *vrf.VRF, peerIP *bnet.IP, c filter.Chain) error {
	p := b.peers.get(vrf, peerIP)
	if p == nil {
		return fmt.Errorf("peer %q not found in VRF %q", peerIP.String(), vrf.Name())
	}

	p.replaceImportFilterChain(c)
	return nil
}

// ReplaceExportFilterChain replaces a peers import filter
func (b *bgpServer) ReplaceExportFilterChain(vrf *vrf.VRF, peerIP *bnet.IP, c filter.Chain) error {
	p := b.peers.get(vrf, peerIP)
	if p == nil {
		return fmt.Errorf("peer %q not found in VRF %q", peerIP.String(), vrf.Name())
	}

	p.replaceExportFilterChain(c)
	return nil
}

func (b *bgpServer) GetRIBIn(vrf *vrf.VRF, peerIP *bnet.IP, afi uint16, safi uint8) *adjRIBIn.AdjRIBIn {
	p := b.peers.get(vrf, peerIP)
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

func (b *bgpServer) GetRIBOut(vrf *vrf.VRF, peerIP *bnet.IP, afi uint16, safi uint8) *adjRIBOut.AdjRIBOut {
	p := b.peers.get(vrf, peerIP)
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
		c := <-b.listenerManager.AcceptCh()

		peerAddr, _ := bnetutils.BIONetIPFromAddr(c.Conn.RemoteAddr().String())
		peer := b.peers.get(c.VRF, peerAddr.Dedup())
		if peer == nil {
			c.Conn.Close()
			log.WithFields(log.Fields{
				"source": c.Conn.RemoteAddr(),
			}).Info("TCP connection from unknown source")
			continue
		}

		log.WithFields(log.Fields{
			"source": c.Conn.RemoteAddr(),
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
		fsm.conCh <- c.Conn
	}
}

func (b *bgpServer) Start() {
	go b.incomingConnectionWorker()
}

func (b *bgpServer) AddPeer(c PeerConfig) error {
	c.LocalAddress = c.LocalAddress.Dedup()
	c.PeerAddress = c.PeerAddress.Dedup()

	peer, err := newPeer(c, b)
	if err != nil {
		return err
	}

	err = b.listenerManager.CreateListenersIfNotExists(c.VRF)
	if err != nil {
		return err
	}

	if c.AuthenticationKey != "" {
		for _, l := range b.listenerManager.GetListeners(c.VRF) {
			err = l.SetTCPMD5(c.PeerAddress.ToNetIP(), c.AuthenticationKey)
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
func (b *bgpServer) GetPeerConfig(vrf *vrf.VRF, addr *bnet.IP) *PeerConfig {
	p := b.peers.get(vrf, addr)
	if p != nil {
		return p.config
	}

	return nil
}

func (b *bgpServer) DisposePeer(vrf *vrf.VRF, addr *bnet.IP) {
	p := b.peers.get(vrf, addr)
	if p == nil {
		return
	}

	log.Infof("disposing BGP session with %s", addr.String())
	p.stop()
	b.peers.remove(PeerKey{
		vrf:        vrf,
		neighborIP: addr,
	})
}

func (b *bgpServer) Metrics() (*metrics.BGPMetrics, error) {
	if b.metrics == nil {
		return nil, fmt.Errorf("server not started yet")
	}

	return b.metrics.metrics(), nil
}
