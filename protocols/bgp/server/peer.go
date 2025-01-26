package server

import (
	"fmt"
	"sync"
	"time"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
)

type peer struct {
	server    *bgpServer
	config    *PeerConfig
	addr      *bnet.IP
	localAddr *bnet.IP
	ttl       uint8
	passive   bool
	peerASN   uint32
	localASN  uint32

	// guarded by fsmsMu
	fsms   []*FSM
	fsmsMu sync.Mutex

	routerID                    uint32
	reconnectInterval           time.Duration
	keepaliveTime               time.Duration
	holdTime                    time.Duration
	optOpenParams               []packet.OptParam
	routeServerClient           bool
	routeReflectorClient        bool
	ipv4MultiProtocolAdvertised bool
	clusterID                   uint32
	peerRoleEnabled             bool
	peerRoleStrictMode          bool
	peerRoleLocal               uint8
	peerRoleAdvByPeer           bool
	peerRoleRemote              uint8

	vrf  *vrf.VRF
	ipv4 *peerAddressFamily
	ipv6 *peerAddressFamily

	adjRIBInFactory adjRIBInFactoryI
}

// PeerConfig defines the configuration for a BGP session
type PeerConfig struct {
	AuthenticationKey          string
	AdminEnabled               bool
	ReconnectInterval          time.Duration
	KeepAlive                  time.Duration
	HoldTime                   time.Duration
	LocalAddress               *bnet.IP
	PeerAddress                *bnet.IP
	TTL                        uint8
	LocalAS                    uint32
	PeerAS                     uint32
	Passive                    bool
	RouterID                   uint32
	RouteServerClient          bool
	RouteReflectorClient       bool
	RouteReflectorClusterID    uint32
	AdvertiseIPv4MultiProtocol bool
	PeerRole                   uint8
	PeerRoleStrictMode         bool
	IPv4                       *AddressFamilyConfig
	IPv6                       *AddressFamilyConfig
	VRF                        *vrf.VRF
	Description                string
}

// AddressFamilyConfig represents all configuration parameters specific for an address family
type AddressFamilyConfig struct {
	ImportFilterChain filter.Chain
	ExportFilterChain filter.Chain
	AddPathSend       routingtable.ClientOptions
	AddPathRecv       bool
}

// NeedsRestart determines if the peer needs a restart on cfg change
func (pc *PeerConfig) NeedsRestart(x *PeerConfig) bool {
	if pc.AuthenticationKey != x.AuthenticationKey {
		return true
	}

	if pc.LocalAS != x.LocalAS {
		return true
	}

	if pc.PeerAS != x.PeerAS {
		return true
	}

	if pc.LocalAddress != x.LocalAddress {
		return true
	}

	if pc.HoldTime != x.HoldTime {
		return true
	}

	if pc.RouteReflectorClient != x.RouteReflectorClient {
		return true
	}

	if pc.RouteServerClient != x.RouteServerClient {
		return true
	}

	if peerRoleEnabled(pc.PeerRole) {
		if pc.PeerRole != x.PeerRole {
			return true
		}

		if pc.PeerRoleStrictMode != x.PeerRoleStrictMode {
			return true
		}
	}

	if pc.VRF != x.VRF {
		return true
	}

	if pc.RouterID != x.RouterID {
		return true
	}

	if pc.Passive != x.Passive {
		return true
	}

	return false
}

// replaceImportFilterChain replaces a peers import filter chain
func (p *peer) replaceImportFilterChain(c filter.Chain) {
	p.fsmsMu.Lock()
	defer p.fsmsMu.Unlock()

	for _, fsm := range p.fsms {
		fsm.replaceImportFilterChain(c)
	}
}

// replaceExportFilterChain replaces a peers import filter chain
func (p *peer) replaceExportFilterChain(c filter.Chain) {
	p.fsmsMu.Lock()
	defer p.fsmsMu.Unlock()

	for _, fsm := range p.fsms {
		fsm.replaceExportFilterChain(c)
	}
}

func (p *peer) dumpRIBIn(afi uint16, safi uint8) []*route.Route {
	if len(p.fsms) != 1 {
		return nil
	}

	fsm := p.fsms[0]
	f := fsm.addressFamily(afi, safi)
	if f == nil {
		return nil
	}

	return f.dumpRIBIn()
}

func (p *peer) dumpRIBOut(afi uint16, safi uint8) []*route.Route {
	if len(p.fsms) != 1 {
		return nil
	}

	fsm := p.fsms[0]
	f := fsm.addressFamily(afi, safi)
	if f == nil {
		return nil
	}

	return f.dumpRIBOut()
}

type peerAddressFamily struct {
	rib *locRIB.LocRIB

	importFilterChain filter.Chain
	exportFilterChain filter.Chain

	addPathSend    routingtable.ClientOptions
	addPathReceive bool
}

func (p *peer) addressFamily(afi uint16, safi uint8) *peerAddressFamily {
	if safi != packet.SAFIUnicast {
		return nil
	}

	switch afi {
	case packet.AFIIPv4:
		return p.ipv4
	case packet.AFIIPv6:
		return p.ipv6
	default:
		return nil
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

		if p.shouldCeaseOnCollision(callingFSM) {
			fsm.cease()
		} else {
			return true
		}
	}

	return false
}

func (p *peer) shouldCeaseOnCollision(callingFSM *FSM) bool {
	// RFC6286: For a BGP speaker that supports the AS-wide Unique BGP Identifier,
	// the procedures for connection collision resolution are extended as
	// follows to deal with the case in which the two BGP speakers share the
	// same BGP Identifier (thus, it is only applicable to an external
	// peer):
	//
	//    If the BGP Identifiers of the peers involved in the connection
	//    collision are identical, then the connection initiated by the BGP
	//    speaker with the larger AS number is preserved.
	if p.routerID == callingFSM.neighborID {
		return p.localASN < callingFSM.peer.peerASN
	}

	// RFC4271 collision handling
	return p.routerID < callingFSM.neighborID
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
func newPeer(c PeerConfig, server *bgpServer) (*peer, error) {
	p := &peer{
		server:               server,
		config:               &c,
		addr:                 c.PeerAddress,
		localAddr:            c.LocalAddress,
		ttl:                  c.TTL,
		passive:              c.Passive,
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
		peerRoleEnabled:      peerRoleEnabled(c.PeerRole),
		peerRoleStrictMode:   c.PeerRoleStrictMode,
		peerRoleLocal:        translatePeerRole(c.PeerRole),
		vrf:                  c.VRF,
		adjRIBInFactory:      adjRIBInFactory{},
	}

	if c.IPv4 != nil {
		p.ipv4 = &peerAddressFamily{
			rib:               c.VRF.IPv4UnicastRIB(),
			importFilterChain: filterOrDefault(c.IPv4.ImportFilterChain),
			exportFilterChain: filterOrDefault(c.IPv4.ExportFilterChain),
			addPathReceive:    c.IPv4.AddPathRecv,
			addPathSend:       c.IPv4.AddPathSend,
		}

		if p.ipv4.rib == nil {
			return nil, fmt.Errorf("no RIB for IPv4 unicast configured")
		}
	}

	// If we are a route reflector and no ClusterID was set, use our RouterID
	if p.routeReflectorClient && p.clusterID == 0 {
		p.clusterID = c.RouterID
	}

	caps := make(packet.Capabilities, 0)

	caps = append(caps, addPathCapabilities(c)...)

	caps = append(caps, asn4Capability(c))

	if c.IPv4 != nil && c.AdvertiseIPv4MultiProtocol {
		caps = append(caps, multiProtocolCapability(packet.AFIIPv4))
		p.ipv4MultiProtocolAdvertised = true
	}

	if c.IPv6 != nil {
		p.ipv6 = &peerAddressFamily{
			rib:               c.VRF.IPv6UnicastRIB(),
			importFilterChain: filterOrDefault(c.IPv6.ImportFilterChain),
			exportFilterChain: filterOrDefault(c.IPv6.ExportFilterChain),
			addPathReceive:    c.IPv6.AddPathRecv,
			addPathSend:       c.IPv6.AddPathSend,
		}
		caps = append(caps, multiProtocolCapability(packet.AFIIPv6))

		if p.ipv6.rib == nil {
			return nil, fmt.Errorf("no RIB for IPv6 unicast configured")
		}
	}

	// Activate Peer Role capability for eBGP neighbors if configured
	if p.localASN != p.peerASN && peerRoleEnabled(c.PeerRole) {
		caps = append(caps, peerRoleCapability(c))
	}

	p.optOpenParams = append(p.optOpenParams, packet.OptParam{
		Type:  packet.CapabilitiesParamType,
		Value: caps,
	})

	if !p.passive {
		p.fsms = append(p.fsms, NewActiveFSM(p))
	}

	return p, nil
}

func asn4Capability(c PeerConfig) packet.Capability {
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
			SAFI: packet.SAFIUnicast,
		},
	}
}

func addPathCapabilities(c PeerConfig) []packet.Capability {
	caps := make([]packet.Capability, 0)

	enabled, cap := addPathCapabilityForFamily(c.IPv4, packet.AFIIPv4, packet.SAFIUnicast)
	if enabled {
		caps = append(caps, cap)
	}

	enabled, cap = addPathCapabilityForFamily(c.IPv6, packet.AFIIPv6, packet.SAFIUnicast)
	if enabled {
		caps = append(caps, cap)
	}

	return caps
}

func addPathCapabilityForFamily(f *AddressFamilyConfig, afi uint16, safi uint8) (enabled bool, cap packet.Capability) {
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
			packet.AddPathCapabilityTuple{
				AFI:         afi,
				SAFI:        safi,
				SendReceive: addPath,
			},
		},
	}
}

func peerRoleCapability(c PeerConfig) packet.Capability {
	return packet.Capability{
		Code: packet.PeerRoleCapabilityCode,
		Value: packet.PeerRoleCapability{
			PeerRole: c.PeerRole,
		},
	}
}

func filterOrDefault(c filter.Chain) filter.Chain {
	if len(c) != 0 {
		return c
	}

	return filter.NewDrainFilterChain()
}

// GetAddr returns the IP address of the peer
func (p *peer) GetAddr() *bnet.IP {
	return p.addr
}

func (p *peer) Start() {
	p.fsms[0].start()
}

// Stop stops a peer BGP session
func (p *peer) stop() {
	p.fsmsMu.Lock()
	defer p.fsmsMu.Unlock()

	for _, fsm := range p.fsms {
		fsm.eventCh <- ManualStop
	}
}

func (p *peer) isEBGP() bool {
	return p.localASN != p.peerASN
}

func (p *peer) getBindDev() string {
	if p.vrf.Name() != vrf.DefaultVRFName {
		return p.vrf.Name()
	}

	return ""
}

func (p *peer) peerKey() PeerKey {
	return PeerKey{
		vrf:        p.vrf,
		neighborIP: p.addr.Dedup(),
	}
}
