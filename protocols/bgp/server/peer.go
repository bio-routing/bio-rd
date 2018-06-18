package server

import (
	"net"

	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"

	"time"
)

type Peer struct {
	addr              net.IP
	asn               uint32
	fsm               *FSM
	rib               routingtable.RouteTableClient
	routerID          uint32
	addPathSend       routingtable.ClientOptions
	addPathRecv       bool
	optOpenParams     []packet.OptParam
	reconnectInterval time.Duration
	importFilter      *filter.Filter
	exportFilter      *filter.Filter
}

// NewPeer creates a new peer with the given config. If an connection is established, the adjRIBIN of the peer is connected
// to the given rib. To actually connect the peer, call Start() on the returned peer.
func NewPeer(c config.Peer, rib routingtable.RouteTableClient) (*Peer, error) {
	p := &Peer{
		addr:              c.PeerAddress,
		asn:               c.PeerAS,
		rib:               rib,
		addPathSend:       c.AddPathSend,
		addPathRecv:       c.AddPathRecv,
		optOpenParams:     make([]packet.OptParam, 0),
		reconnectInterval: c.ReconnectInterval,
		importFilter:      filterOrDefault(c.ImportFilter),
		exportFilter:      filterOrDefault(c.ExportFilter),
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

// GetAddr returns the IP address of the peer
func (p *Peer) GetAddr() net.IP {
	return p.addr
}

// GetASN returns the configured AS number of the peer
func (p *Peer) GetASN() uint32 {
	return p.asn
}

// Start the peers fsm. It starts from the Idle state and will get an ManualStart event. To trigger
// reconnects if the fsm falls back into the Idle state, every reconnectInterval a ManualStart event is send.
// The default value for reconnectInterval is 30 seconds.
func (p *Peer) Start() {
	p.fsm.start()
	if p.reconnectInterval == 0 {
		p.reconnectInterval = 30 * time.Second
	}
	t := time.Tick(p.reconnectInterval)
	go func() {
		for {
			<-t
			p.fsm.activate()
		}
	}()
}

func filterOrDefault(f *filter.Filter) *filter.Filter {
	if f != nil {
		return f
	}

	return filter.NewDrainFilter()
}
