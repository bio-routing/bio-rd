package server

import (
	"fmt"
	"io"
	"net"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/taktv6/tbgp/config"
	"github.com/taktv6/tbgp/packet"
)

const (
	uint16max  = 65535
	BGPVersion = 4
)

type BGPServer struct {
	listeners []*TCPListener
	acceptCh  chan *net.TCPConn
	peers     map[string]*Peer
	routerID  uint32
}

func NewBgpServer() *BGPServer {
	return &BGPServer{
		peers: make(map[string]*Peer),
	}
}

func (b *BGPServer) RouterID() uint32 {
	return b.routerID
}

func (b *BGPServer) Start(c *config.Global) error {
	if err := c.SetDefaultGlobalConfigValues(); err != nil {
		return fmt.Errorf("Failed to load defaults: %v", err)
	}

	fmt.Printf("ROUTER ID: %d\n", c.RouterID)
	b.routerID = c.RouterID

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

func (b *BGPServer) incomingConnectionWorker() {
	for {
		c := <-b.acceptCh
		fmt.Printf("Incoming connection!\n")
		fmt.Printf("Connection from: %v\n", c.RemoteAddr())

		peerAddr := strings.Split(c.RemoteAddr().String(), ":")[0]
		if _, ok := b.peers[peerAddr]; !ok {
			c.Close()
			log.WithFields(log.Fields{
				"source": c.RemoteAddr(),
			}).Warning("TCP connection from unknown source")
			continue
		}

		log.WithFields(log.Fields{
			"source": c.RemoteAddr(),
		}).Info("Incoming TCP connection")

		fmt.Printf("DEBUG: Sending incoming TCP connection to fsm for peer %s\n", peerAddr)
		b.peers[peerAddr].fsm.conCh <- c
		fmt.Printf("DEBUG: Sending done\n")
	}
}

func (b *BGPServer) AddPeer(c config.Peer) error {
	if c.LocalAS > uint16max || c.PeerAS > uint16max {
		return fmt.Errorf("32bit ASNs are not supported yet")
	}

	peer, err := NewPeer(c)
	if err != nil {
		return err
	}

	peer.routerID = c.RouterID
	peerAddr := peer.GetAddr().String()
	b.peers[peerAddr] = peer
	b.peers[peerAddr].Start()

	return nil
}

func recvMsg(c *net.TCPConn) (msg []byte, err error) {
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
