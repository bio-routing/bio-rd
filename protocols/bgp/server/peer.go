package server

import (
	"net"
	"sync"
	"time"

	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"

	"github.com/bio-routing/bio-rd/config"
)

type Peer struct {
	server            *BGPServer
	addr              net.IP
	peerASN           uint32
	localASN          uint32
	fsms              []*FSM2
	fsmsMu            sync.Mutex
	rib               *locRIB.LocRIB
	routerID          uint32
	addPathSend       routingtable.ClientOptions
	addPathRecv       bool
	reconnectInterval time.Duration
	keepaliveTime     time.Duration
	holdTime          time.Duration
	optOpenParams     []packet.OptParam
	importFilter      *filter.Filter
	exportFilter      *filter.Filter
}

func (p *Peer) collisionHandling(callingFSM *FSM2) bool {
	p.fsmsMu.Lock()
	defer p.fsmsMu.Unlock()

	for _, fsm := range p.fsms {
		if callingFSM == fsm {
			continue
		}

		fsm.stateMu.RLock()
		isEstablished := isEstablishedState(fsm.state)
		isOpenConfirm := isOpenConfirmState(fsm.state)
		fsm.stateMu.RUnlock()

		if isEstablished {
			return true
		}

		if !isOpenConfirm {
			continue
		}

		if p.routerID < callingFSM.neighborID {
			fsm.cease()
		} else {
			return true
		}
	}

	return false
}

func isOpenConfirmState(s state) bool {
	switch s.(type) {
	case openConfirmState:
		return true
	}

	return false
}

func isEstablishedState(s state) bool {
	switch s.(type) {
	case establishedState:
		return true
	}

	return false
}

func NewPeer(c config.Peer, rib *locRIB.LocRIB, server *BGPServer) (*Peer, error) {
	p := &Peer{
		server:            server,
		addr:              c.PeerAddress,
		peerASN:           c.PeerAS,
		localASN:          c.LocalAS,
		fsms:              make([]*FSM2, 0),
		rib:               rib,
		addPathSend:       c.AddPathSend,
		addPathRecv:       c.AddPathRecv,
		reconnectInterval: c.ReconnectInterval,
		keepaliveTime:     c.KeepAlive,
		holdTime:          c.HoldTimer,
		optOpenParams:     make([]packet.OptParam, 0),
	}
	p.fsms = append(p.fsms, NewActiveFSM2(p))

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

func (p *Peer) Start() {
	p.fsms[0].start()
}
