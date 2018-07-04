package server

import (
	"sync"
	"time"

	"github.com/bio-routing/bio-rd/config"
	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
)

type PeerInfo struct {
	PeerAddr bnet.IP
	PeerASN  uint32
	LocalASN uint32
}

type peer struct {
	server   *bgpServer
	addr     bnet.IP
	peerASN  uint32
	localASN uint32

	// guarded by fsmsMu
	fsms   []*FSM
	fsmsMu sync.Mutex

	rib                  *locRIB.LocRIB
	routerID             uint32
	addPathSend          routingtable.ClientOptions
	addPathRecv          bool
	reconnectInterval    time.Duration
	keepaliveTime        time.Duration
	holdTime             time.Duration
	optOpenParams        []packet.OptParam
	importFilter         *filter.Filter
	exportFilter         *filter.Filter
	routeServerClient    bool
	routeReflectorClient bool
	clusterID            uint32
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
func newPeer(c config.Peer, rib *locRIB.LocRIB, server *bgpServer) (*peer, error) {
	if c.LocalAS == 0 {
		c.LocalAS = server.localASN
	}
	p := &peer{
		server:               server,
		addr:                 c.PeerAddress,
		peerASN:              c.PeerAS,
		localASN:             c.LocalAS,
		fsms:                 make([]*FSM, 0),
		rib:                  rib,
		addPathSend:          c.AddPathSend,
		addPathRecv:          c.AddPathRecv,
		reconnectInterval:    c.ReconnectInterval,
		keepaliveTime:        c.KeepAlive,
		holdTime:             c.HoldTime,
		optOpenParams:        make([]packet.OptParam, 0),
		importFilter:         filterOrDefault(c.ImportFilter),
		exportFilter:         filterOrDefault(c.ExportFilter),
		routeServerClient:    c.RouteServerClient,
		routeReflectorClient: c.RouteReflectorClient,
		clusterID:            c.RouteReflectorClusterID,
	}

	// If we are a route reflector and no ClusterID was set, use our RouterID
	if p.routeReflectorClient && p.clusterID == 0 {
		p.clusterID = c.RouterID
	}

	p.fsms = append(p.fsms, NewActiveFSM2(p))

	caps := make(packet.Capabilities, 0)

	addPathEnabled, addPathCap := handleAddPathCapability(c)
	if addPathEnabled {
		caps = append(caps, addPathCap)
	}

	caps = append(caps, asn4Capability(c))

	if c.IPv6 {
		caps = append(caps, multiProtocolCapability(packet.IPv6AFI))
	}

	p.optOpenParams = append(p.optOpenParams, packet.OptParam{
		Type:  packet.CapabilitiesParamType,
		Value: caps,
	})

	return p, nil
}

func asn4Capability(c config.Peer) packet.Capability {
	return packet.Capability{
		Code: packet.ASN4CapabilityCode,
		Value: packet.ASN4Capability{
			ASN4: c.LocalAS,
		},
	}
}

func multiProtocolCapability(afi uint16) packet.Capability {
	return packet.Capability{
		Code: packet.MultiProtocolCapabilityCode,
		Value: packet.MultiProtocolCapability{
			AFI:  afi,
			SAFI: packet.UnicastSAFI,
		},
	}
}

func handleAddPathCapability(c config.Peer) (bool, packet.Capability) {
	addPath := uint8(0)
	if c.AddPathRecv {
		addPath += packet.AddPathReceive
	}
	if !c.AddPathSend.BestOnly {
		addPath += packet.AddPathSend
	}

	if addPath == 0 {
		return false, packet.Capability{}
	}

	return true, packet.Capability{
		Code: packet.AddPathCapabilityCode,
		Value: packet.AddPathCapability{
			AFI:         packet.IPv4AFI,
			SAFI:        packet.UnicastSAFI,
			SendReceive: addPath,
		},
	}
}

func filterOrDefault(f *filter.Filter) *filter.Filter {
	if f != nil {
		return f
	}

	return filter.NewDrainFilter()
}

// GetAddr returns the IP address of the peer
func (p *peer) GetAddr() bnet.IP {
	return p.addr
}

func (p *peer) Start() {
	p.fsms[0].start()
}
