package server

import (
	"net"

	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/routingtable"
)

type Peer struct {
	addr          net.IP
	asn           uint32
	fsm           *FSM
	rib           routingtable.RouteTableClient
	routerID      uint32
	addPathSend   routingtable.ClientOptions
	addPathRecv   bool
	optOpenParams []packet.OptParam
}

func NewPeer(c config.Peer, rib routingtable.RouteTableClient) (*Peer, error) {
	p := &Peer{
		addr:          c.PeerAddress,
		asn:           c.PeerAS,
		rib:           rib,
		addPathSend:   c.AddPathSend,
		addPathRecv:   c.AddPathRecv,
		optOpenParams: make([]packet.OptParam, 0),
	}
	p.fsm = NewFSM(p, c, rib)

	caps := make([]packet.Capability, 0)

	addPath := uint8(0)
	if c.AddPathRecv {
		addPath += packet.AddPathReceive
	}
	if !c.AddPathSend.BestOnly {
		addPath += packet.AddPathSend
	}

	if addPath > 0 {
		caps = append(caps, packet.Capability{
			Code: packet.AddPathCapabilityCode,
			Value: packet.AddPathCapability{
				AFI:         packet.IPv4AFI,
				SAFI:        packet.UnicastSAFI,
				SendReceive: addPath,
			},
		})
	}

	for _, cap := range caps {
		p.optOpenParams = append(p.optOpenParams, packet.OptParam{
			Type:  packet.CapabilitiesParamType,
			Value: cap,
		})
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
