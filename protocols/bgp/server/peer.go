package server

import (
	"net"

	"github.com/bio-routing/bio-rd/routingtable/locRIB"

	"github.com/bio-routing/bio-rd/config"
)

type Peer struct {
	addr     net.IP
	asn      uint32
	fsm      *FSM
	rib      *locRIB.LocRIB
	routerID uint32
}

func NewPeer(c config.Peer, rib *locRIB.LocRIB) (*Peer, error) {
	p := &Peer{
		addr: c.PeerAddress,
		asn:  c.PeerAS,
		fsm:  NewFSM(c, rib),
		rib:  rib,
	}
	return p, nil
}

func (p *Peer) GetAddr() net.IP {
	return p.addr
}

func (p *Peer) GetASN() uint32 {
	return p.asn
}

func (p *Peer) Start() {
	p.fsm.start()
	p.fsm.activate()
}
