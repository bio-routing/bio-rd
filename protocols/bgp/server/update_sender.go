package server

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/util/log"
)

// UpdateSender converts table changes into BGP update messages
type UpdateSender struct {
	clientManager *routingtable.ClientManager
	fsm           *FSM
	addressFamily *fsmAddressFamily
	options       *packet.EncodeOptions
	iBGP          bool
	rrClient      bool
	toSendMu      sync.Mutex
	toSend        map[string]*pathPfxs
	destroyCh     chan struct{}
	wg            sync.WaitGroup
}

type pathPfxs struct {
	path *route.Path
	pfxs []*bnet.Prefix
}

func newUpdateSender(f *fsmAddressFamily) *UpdateSender {
	u := &UpdateSender{
		fsm:           f.fsm,
		addressFamily: f,
		iBGP:          f.fsm.peer.localASN == f.fsm.peer.peerASN,
		rrClient:      f.fsm.peer.routeReflectorClient,
		destroyCh:     make(chan struct{}),
		toSend:        make(map[string]*pathPfxs),
		options: &packet.EncodeOptions{
			Use32BitASN: f.fsm.supports4OctetASN,
			UseAddPath:  !f.addPathTX.BestOnly,
		},
	}
	u.clientManager = routingtable.NewClientManager(u)

	return u
}

// ClientCount is here to satisfy an interface
func (u *UpdateSender) ClientCount() uint64 {
	return 0
}

// Start starts the update sender
func (u *UpdateSender) Start(aggrTime time.Duration) {
	u.wg.Add(1)
	go u.sender(aggrTime)
}

// Destroy destroys everything (with greetings to Hatebreed)
func (u *UpdateSender) Destroy() {
	u.destroyCh <- struct{}{}
}

func (u *UpdateSender) AddPathInitialDump(pfx *bnet.Prefix, p *route.Path) error {
	return u.AddPath(pfx, p)
}

// AddPath adds path p for pfx to toSend queue
func (u *UpdateSender) AddPath(pfx *bnet.Prefix, p *route.Path) error {
	u.toSendMu.Lock()
	defer u.toSendMu.Unlock()

	hash := p.BGPPath.ComputeHashWithPathID()
	if _, exists := u.toSend[hash]; exists {
		u.toSend[hash].pfxs = append(u.toSend[hash].pfxs, pfx)
		return nil
	}

	u.toSend[hash] = &pathPfxs{
		path: p,
		pfxs: []*bnet.Prefix{
			pfx,
		},
	}

	return nil
}

// Dump is here to fulfill an interface
func (u *UpdateSender) Dump() []*route.Route {
	return nil
}

func (u *UpdateSender) EndOfRIB() {
	u.toSendMu.Lock()
	defer u.toSendMu.Unlock()

	u._flush()
	u.sendEndOfRIB()
}

func (u *UpdateSender) sendEndOfRIB() {
	update := &packet.BGPUpdate{
		SAFI: u.addressFamily.safi,
	}
	err := serializeAndSendUpdate(u.fsm.con, update, u.options)
	if err != nil {
		log.Errorf("Failed to serialize and send end of RIB marker: %v", err)
	}
}

// sender serializes BGP update messages
func (u *UpdateSender) sender(aggrTime time.Duration) {
	ticker := time.NewTicker(aggrTime)

	for {
		select {
		case <-u.destroyCh:
			u.wg.Done()
			return
		case <-ticker.C:
		}

		u.toSendMu.Lock()
		for key, pathNLRIs := range u.toSend {
			pathAttrs, updatesPrefixes, pathID := u._getUpdateInformation(pathNLRIs)

			delete(u.toSend, key)
			u.toSendMu.Unlock()

			u.sendUpdates(pathAttrs, updatesPrefixes, pathID)
			u.toSendMu.Lock()
		}
		u.toSendMu.Unlock()
	}
}

func (u *UpdateSender) _getUpdateInformation(pathNLRIs *pathPfxs) (*packet.PathAttribute, [][]*bnet.Prefix, uint32) {
	budget := u.getBudget(pathNLRIs)

	pathAttrs, err := packet.PathAttributes(pathNLRIs.path, u.iBGP, u.rrClient)
	if err != nil {
		log.Errorf("unable to get path attributes: %v", err)
		return nil, nil, 0 // FIXME
	}

	updatesPrefixes := make([][]*bnet.Prefix, 0, 1)
	prefixes := make([]*bnet.Prefix, 0, 1)
	for _, pfx := range pathNLRIs.pfxs {
		budget -= int(packet.BytesInAddr(pfx.Len())) + 1

		if u.options.UseAddPath {
			budget -= packet.PathIdentifierLen
		}

		if budget < 0 {
			updatesPrefixes = append(updatesPrefixes, prefixes)
			prefixes = make([]*bnet.Prefix, 0, 1)
			budget = u.getBudget(pathNLRIs)
		}

		prefixes = append(prefixes, pfx)
	}
	if len(prefixes) > 0 {
		updatesPrefixes = append(updatesPrefixes, prefixes)
	}

	return pathAttrs, updatesPrefixes, pathNLRIs.path.BGPPath.PathIdentifier
}

func (u *UpdateSender) _flush() {
	for key, pathNLRIs := range u.toSend {
		pathAttrs, updatesPrefixes, pathID := u._getUpdateInformation(pathNLRIs)
		delete(u.toSend, key)

		u.sendUpdates(pathAttrs, updatesPrefixes, pathID)
	}
}

func (u *UpdateSender) getBudget(pathNLRIs *pathPfxs) int {
	return packet.MaxLen - packet.HeaderLen - packet.MinUpdateLen - int(pathNLRIs.path.BGPPath.Length()) - u.updateOverhead()
}

func (u *UpdateSender) updateOverhead() int {
	if u.addressFamily.afi == packet.AFIIPv4 && !u.addressFamily.multiProtocol {
		return 0
	}

	addrLen := packet.AFIIPv4
	if u.addressFamily.afi == packet.AFIIPv6 {
		addrLen = packet.IPv6Len
	}

	// since we are replacing the next hop attribute IPv4Len has to be subtracted, we also add another byte for extended length
	return packet.AFILen + packet.SAFILen + 1 + addrLen - packet.IPv4Len + 1
}

func (u *UpdateSender) sendUpdates(pathAttrs *packet.PathAttribute, updatePrefixes [][]*bnet.Prefix, pathID uint32) {
	var err error

	for _, prefixes := range updatePrefixes {
		update := u.updateMessageForPrefixes(prefixes, pathAttrs, pathID)
		if update == nil {
			log.Errorf("Failed to create update: Neighbor does not support multi protocol.")
			return
		}

		err = serializeAndSendUpdate(u.fsm.con, update, u.options)
		if err != nil {
			log.Errorf("Failed to serialize and send: %v", err)
		}
		atomic.AddUint64(&u.fsm.counters.updatesSent, 1)
	}
}

func (u *UpdateSender) updateMessageForPrefixes(pfxs []*bnet.Prefix, pa *packet.PathAttribute, pathID uint32) *packet.BGPUpdate {
	if u.addressFamily.afi == packet.AFIIPv4 && !u.addressFamily.multiProtocol {
		return u.bgpUpdate(pfxs, pa, pathID)
	}

	if u.addressFamily.multiProtocol {
		return u.bgpUpdateMultiProtocol(pfxs, pa, pathID)
	}

	return nil
}

func (u *UpdateSender) bgpUpdate(pfxs []*bnet.Prefix, pa *packet.PathAttribute, pathID uint32) *packet.BGPUpdate {
	update := &packet.BGPUpdate{
		PathAttributes: pa,
		SAFI:           u.addressFamily.safi,
	}

	var nlri *packet.NLRI
	for _, pfx := range pfxs {
		nlri = &packet.NLRI{
			PathIdentifier: pathID,
			Prefix:         pfx,
			Next:           update.NLRI,
		}
		update.NLRI = nlri
	}

	return update
}

func (u *UpdateSender) bgpUpdateMultiProtocol(pfxs []*bnet.Prefix, pa *packet.PathAttribute, pathID uint32) *packet.BGPUpdate {
	pa, nextHop := u.copyAttributesWithoutNextHop(pa)

	attrs := &packet.PathAttribute{
		TypeCode: packet.MultiProtocolReachNLRIAttr,
		Value: packet.MultiProtocolReachNLRI{
			AFI:     u.addressFamily.afi,
			SAFI:    u.addressFamily.safi,
			NextHop: nextHop,
			NLRI:    u.nlriForPrefixes(pfxs, pathID),
		},
	}
	attrs.Next = pa

	return &packet.BGPUpdate{
		PathAttributes: attrs,
		SAFI:           u.addressFamily.safi,
	}
}

func (u *UpdateSender) nlriForPrefixes(pfxs []*bnet.Prefix, pathID uint32) *packet.NLRI {
	var prev, res *packet.NLRI
	for _, pfx := range pfxs {
		cur := &packet.NLRI{
			Prefix:         pfx,
			PathIdentifier: pathID,
		}

		if res == nil {
			res = cur
			prev = cur
			continue
		}

		prev.Next = cur
		prev = cur
	}

	return res
}

func (u *UpdateSender) copyAttributesWithoutNextHop(pa *packet.PathAttribute) (attrs *packet.PathAttribute, nextHop *bnet.IP) {
	var curCopy, lastCopy *packet.PathAttribute
	for cur := pa; cur != nil; cur = cur.Next {
		if cur.TypeCode == packet.NextHopAttr {
			nextHop = cur.Value.(*bnet.IP)
		} else {
			curCopy = cur.Copy()

			if lastCopy == nil {
				attrs = curCopy
			} else {
				lastCopy.Next = curCopy
			}
			lastCopy = curCopy
		}
	}

	return attrs, nextHop
}

// RemovePath withdraws prefix `pfx` from a peer
func (u *UpdateSender) RemovePath(pfx *bnet.Prefix, p *route.Path) bool {
	err := u.withdrawPrefix(u.fsm.con, pfx, p)
	if err != nil {
		log.Errorf("unable to withdraw prefix: %v", err)
		return false
	}

	return true
}

func (u *UpdateSender) withdrawPrefix(out io.Writer, pfx *bnet.Prefix, p *route.Path) error {
	if p.Type != route.BGPPathType {
		return errors.New("wrong path type, expected BGPPathType")
	}

	if p.BGPPath == nil {
		return errors.New("got nil BGPPath")
	}

	if u.addressFamily.afi == packet.AFIIPv4 && !u.addressFamily.multiProtocol {
		return u.withdrawPrefixIPv4(out, pfx, p)
	}

	if !u.addressFamily.multiProtocol {
		return fmt.Errorf(packet.AFIName(u.addressFamily.afi) + " was not negotiated")
	}

	return u.withdrawPrefixMultiProtocol(out, pfx, p)
}

func (u *UpdateSender) withdrawPrefixIPv4(out io.Writer, pfx *bnet.Prefix, p *route.Path) error {
	update := &packet.BGPUpdate{
		SAFI: packet.SAFIUnicast,
		WithdrawnRoutes: &packet.NLRI{
			PathIdentifier: p.BGPPath.PathIdentifier,
			Prefix:         pfx,
		},
	}

	return serializeAndSendUpdate(out, update, u.options)
}

func (u *UpdateSender) withdrawPrefixMultiProtocol(out io.Writer, pfx *bnet.Prefix, p *route.Path) error {
	pathID := uint32(0)
	if p.BGPPath != nil {
		pathID = p.BGPPath.PathIdentifier
	}

	update := &packet.BGPUpdate{
		PathAttributes: &packet.PathAttribute{
			TypeCode: packet.MultiProtocolUnreachNLRIAttr,
			Value: packet.MultiProtocolUnreachNLRI{
				AFI:  u.addressFamily.afi,
				SAFI: u.addressFamily.safi,
				NLRI: &packet.NLRI{
					PathIdentifier: pathID,
					Prefix:         pfx,
				},
			},
		},
	}

	return serializeAndSendUpdate(out, update, u.options)
}

// UpdateNewClient does nothing
func (u *UpdateSender) UpdateNewClient(client routingtable.RouteTableClient) error {
	log.Error("BGP Update Sender: UpdateNewClient not implemented")
	return nil
}

// RouteCount returns the number of stored routes
func (u *UpdateSender) RouteCount() int64 {
	log.Error("BGP Update Sender: RouteCount not implemented")
	return 0
}

// Register registers a client for updates
func (a *UpdateSender) Register(client routingtable.RouteTableClient) {
	a.clientManager.RegisterWithOptions(client, routingtable.ClientOptions{BestOnly: true})
}

// RegisterWithOptions registers a client with options for updates
func (a *UpdateSender) RegisterWithOptions(client routingtable.RouteTableClient, opt routingtable.ClientOptions) {
	a.clientManager.RegisterWithOptions(client, opt)
}

// Unregister unregisters a client
func (a *UpdateSender) Unregister(client routingtable.RouteTableClient) {
	a.clientManager.Unregister(client)
}

// ReplaceFilterChain is here to fulfill an interface
func (a *UpdateSender) ReplaceFilterChain(c filter.Chain) {

}

// ReplacePath is here to fulfill an interface
func (a *UpdateSender) ReplacePath(*bnet.Prefix, *route.Path, *route.Path) {

}

// RefreshRoute is here to fultill an interface
func (a *UpdateSender) RefreshRoute(*bnet.Prefix, []*route.Path) {

}

// Dispose is here to fulfill an interface
func (u *UpdateSender) Dispose() {}
