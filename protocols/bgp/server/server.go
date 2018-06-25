package server

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync"

	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	log "github.com/sirupsen/logrus"
)

const (
	BGPVersion = 4
)

type bgpServer struct {
	listeners []*TCPListener
	acceptCh  chan *net.TCPConn
	peers     sync.Map
	routerID  uint32
	localASN  uint32
}

type BGPServer interface {
	RouterID() uint32
	Start(*config.Global) error
	AddPeer(config.Peer, *locRIB.LocRIB) error
	GetPeerInfoAll() map[string]PeerInfo
}

func NewBgpServer() BGPServer {
	return &bgpServer{}
}

func (b *bgpServer) GetPeerInfoAll() map[string]PeerInfo {
	res := make(map[string]PeerInfo)
	b.peers.Range(func(key, value interface{}) bool {
		name := key.(string)
		peer := value.(*peer)

		res[name] = peer.snapshot()

		return true
	})
	return res
}

func (b *bgpServer) RouterID() uint32 {
	return b.routerID
}

func (b *bgpServer) Start(c *config.Global) error {
	if err := c.SetDefaultGlobalConfigValues(); err != nil {
		return fmt.Errorf("Failed to load defaults: %v", err)
	}

	log.Infof("ROUTER ID: %d\n", c.RouterID)
	b.routerID = c.RouterID
	b.localASN = c.LocalASN

	if c.Listen {
		acceptCh := make(chan *net.TCPConn, 4096)
		for _, addr := range c.LocalAddressList {
			l, err := NewTCPListener(addr, c.Port, acceptCh)
			if err != nil {
				return fmt.Errorf("Failed to start TCPListener for %s: %v", addr.String(), err)
			}
			b.listeners = append(b.listeners, l)
		}
		b.acceptCh = acceptCh

		go b.incomingConnectionWorker()
	}

	return nil
}

func (b *bgpServer) incomingConnectionWorker() {
	for {
		c := <-b.acceptCh

		peerAddr := strings.Split(c.RemoteAddr().String(), ":")[0]
		peerInterface, ok := b.peers.Load(peerAddr)
		if !ok {
			c.Close()
			log.WithFields(log.Fields{
				"source": c.RemoteAddr(),
			}).Warning("TCP connection from unknown source")
			continue
		}
		peer := peerInterface.(*peer)

		log.WithFields(log.Fields{
			"source": c.RemoteAddr(),
		}).Info("Incoming TCP connection")

		log.WithField("Peer", peerAddr).Debug("Sending incoming TCP connection to fsm for peer")
		fsm := NewActiveFSM2(peer)
		fsm.state = newActiveState(fsm)
		fsm.startConnectRetryTimer()

		peer.fsmsMu.Lock()
		peer.fsms = append(peer.fsms, fsm)
		peer.fsmsMu.Unlock()

		go fsm.run()
		fsm.conCh <- c
	}
}

func (b *bgpServer) AddPeer(c config.Peer, rib *locRIB.LocRIB) error {
	peer, err := newPeer(c, rib, b)
	if err != nil {
		return err
	}

	peer.routerID = c.RouterID
	peerAddr := peer.GetAddr().String()
	b.peers.Store(peerAddr, peer)
	peer.Start()

	return nil
}

func recvMsg(c net.Conn) (msg []byte, err error) {
	buffer := make([]byte, packet.MaxLen)
	_, err = io.ReadFull(c, buffer[0:packet.MinLen])
	if err != nil {
		return nil, fmt.Errorf("Read failed: %v", err)
	}

	l := int(buffer[16])*256 + int(buffer[17])
	toRead := l
	_, err = io.ReadFull(c, buffer[packet.MinLen:toRead])
	if err != nil {
		return nil, fmt.Errorf("Read failed: %v", err)
	}

	return buffer, nil
}
