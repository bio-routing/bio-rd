package server

import (
	"net"
	"sync"
	"time"

	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"
)

type PeerInfo struct {
	PeerAddr net.IP
	PeerASN  uint32
	LocalASN uint32
}

type peer struct {
	server   *bgpServer
	addr     net.IP
	peerASN  uint32
	localASN uint32

	// guarded by fsmsMu
	fsms   []*FSM
	fsmsMu sync.Mutex

	rib               routingtable.RouteTableClient
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

func (p *peer) snapshot() PeerInfo {
	return PeerInfo{
		PeerAddr: p.addr,
		PeerASN:  p.peerASN,
		LocalASN: p.localASN,
	}
}

func (p *peer) collisionHandling(callingFSM *FSM) bool {
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

// NewPeer creates a new peer with the given config. If an connection is established, the adjRIBIN of the peer is connected
// to the given rib. To actually connect the peer, call Start() on the returned peer.
func newPeer(c config.Peer, rib routingtable.RouteTableClient, server *bgpServer) (*peer, error) {
	if c.LocalAS == 0 {
		c.LocalAS = server.localASN
	}
	p := &peer{
		server:            server,
		addr:              c.PeerAddress,
		peerASN:           c.PeerAS,
		localASN:          c.LocalAS,
		fsms:              make([]*FSM, 0),
		rib:               rib,
		addPathSend:       c.AddPathSend,
		addPathRecv:       c.AddPathRecv,
		reconnectInterval: c.ReconnectInterval,
		keepaliveTime:     c.KeepAlive,
		holdTime:          c.HoldTime,
		optOpenParams:     make([]packet.OptParam, 0),
		importFilter:      filterOrDefault(c.ImportFilter),
		exportFilter:      filterOrDefault(c.ExportFilter),
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

func filterOrDefault(f *filter.Filter) *filter.Filter {
	if f != nil {
		return f
	}

	return filter.NewDrainFilter()
}

// GetAddr returns the IP address of the peer
func (p *peer) GetAddr() net.IP {
	return p.addr
}

func (p *peer) Start() {
	p.fsms[0].start()
}
