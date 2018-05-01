package server

import (
	"net"

	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/rt"
)

type Peer struct {
	addr     net.IP
	asn      uint32
	fsm      *FSM
	vrf      *rt.RT
	routerID uint32
}

func NewPeer(c config.Peer, vrf *rt.RT) (*Peer, error) {
	p := &Peer{
		addr: c.PeerAddress,
		asn:  c.PeerAS,
		fsm:  NewFSM(c, vrf),
		vrf:  vrf,
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
