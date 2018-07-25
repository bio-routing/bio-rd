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
	States   []string
}

type peer struct {
	server   *bgpServer
	addr     bnet.IP
	peerASN  uint32
	localASN uint32

	// guarded by fsmsMu
	fsms   []*FSM
	fsmsMu sync.Mutex

	routerID             uint32
	reconnectInterval    time.Duration
	keepaliveTime        time.Duration
	holdTime             time.Duration
	optOpenParams        []packet.OptParam
	routeServerClient    bool
	routeReflectorClient bool
	clusterID            uint32

	ipv4 *peerAddressFamily
	ipv6 *peerAddressFamily
}

type peerAddressFamily struct {
	rib *locRIB.LocRIB

	importFilter *filter.Filter
	exportFilter *filter.Filter

	addPathSend    routingtable.ClientOptions
	addPathReceive bool
}

func (p *peer) snapshot() PeerInfo {
	p.fsmsMu.Lock()
	defer p.fsmsMu.Unlock()
	states := make([]string, 0, len(p.fsms))
	for _, fsm := range p.fsms {
		fsm.stateMu.RLock()
		states = append(states, stateName(fsm.state))
		fsm.stateMu.RUnlock()
	}
	return PeerInfo{
		PeerAddr: p.addr,
		PeerASN:  p.peerASN,
		LocalASN: p.localASN,
		States:   states,
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
func newPeer(c config.Peer, server *bgpServer) (*peer, error) {
	if c.LocalAS == 0 {
		c.LocalAS = server.localASN
	}

	p := &peer{
		server:               server,
		addr:                 c.PeerAddress,
		peerASN:              c.PeerAS,
		localASN:             c.LocalAS,
		fsms:                 make([]*FSM, 0),
		reconnectInterval:    c.ReconnectInterval,
		keepaliveTime:        c.KeepAlive,
		holdTime:             c.HoldTime,
		optOpenParams:        make([]packet.OptParam, 0),
		routeServerClient:    c.RouteServerClient,
		routeReflectorClient: c.RouteReflectorClient,
		clusterID:            c.RouteReflectorClusterID,
	}

	if c.IPv4 != nil {
		p.ipv4 = &peerAddressFamily{
			rib:            c.IPv4.RIB,
			importFilter:   filterOrDefault(c.IPv4.ImportFilter),
			exportFilter:   filterOrDefault(c.IPv4.ExportFilter),
			addPathReceive: c.IPv4.AddPathRecv,
			addPathSend:    c.IPv4.AddPathSend,
		}
	}

	// If we are a route reflector and no ClusterID was set, use our RouterID
	if p.routeReflectorClient && p.clusterID == 0 {
		p.clusterID = c.RouterID
	}

	caps := make(packet.Capabilities, 0)

	caps = append(caps, addPathCapabilities(c)...)

	caps = append(caps, asn4Capability(c))

	if c.IPv6 != nil {
		p.ipv6 = &peerAddressFamily{
			rib:            c.IPv6.RIB,
			importFilter:   filterOrDefault(c.IPv6.ImportFilter),
			exportFilter:   filterOrDefault(c.IPv6.ExportFilter),
			addPathReceive: c.IPv6.AddPathRecv,
			addPathSend:    c.IPv6.AddPathSend,
		}
		caps = append(caps, multiProtocolCapability(packet.IPv6AFI))
	}

	p.optOpenParams = append(p.optOpenParams, packet.OptParam{
		Type:  packet.CapabilitiesParamType,
		Value: caps,
	})

	p.fsms = append(p.fsms, NewActiveFSM(p))

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

func addPathCapabilities(c config.Peer) []packet.Capability {
	caps := make([]packet.Capability, 0)

	enabled, cap := addPathCapabilityForFamily(c.IPv4, packet.IPv4AFI, packet.UnicastSAFI)
	if enabled {
		caps = append(caps, cap)
	}

	enabled, cap = addPathCapabilityForFamily(c.IPv6, packet.IPv6AFI, packet.UnicastSAFI)
	if enabled {
		caps = append(caps, cap)
	}

	return caps
}

func addPathCapabilityForFamily(f *config.AddressFamilyConfig, afi uint16, safi uint8) (enabled bool, cap packet.Capability) {
	if f == nil {
		return false, packet.Capability{}
	}

	addPath := uint8(0)
	if f.AddPathRecv {
		addPath += packet.AddPathReceive
	}
	if !f.AddPathSend.BestOnly {
		addPath += packet.AddPathSend
	}

	if addPath == 0 {
		return false, packet.Capability{}
	}

	return true, packet.Capability{
		Code: packet.AddPathCapabilityCode,
		Value: packet.AddPathCapability{
			AFI:         afi,
			SAFI:        safi,
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
