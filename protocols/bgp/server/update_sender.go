package server

import (
	"sync"
	"time"

	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"

	bnet "github.com/bio-routing/bio-rd/net"
	log "github.com/sirupsen/logrus"
)

// UpdateSender converts table changes into BGP update messages
type UpdateSender struct {
	routingtable.ClientManager
	fsm       *FSM
	afi       uint16
	safi      uint8
	iBGP      bool
	rrClient  bool
	toSendMu  sync.Mutex
	toSend    map[string]*pathPfxs
	destroyCh chan struct{}
}

type pathPfxs struct {
	path *route.Path
	pfxs []bnet.Prefix
}

func newUpdateSender(fsm *FSM, afi uint16, safi uint8) *UpdateSender {
	return &UpdateSender{
		fsm:       fsm,
		afi:       afi,
		safi:      safi,
		iBGP:      fsm.peer.localASN == fsm.peer.peerASN,
		rrClient:  fsm.peer.routeReflectorClient,
		destroyCh: make(chan struct{}),
		toSend:    make(map[string]*pathPfxs),
	}
}

// Start starts the update sender
func (u *UpdateSender) Start(aggrTime time.Duration) {
	go u.sender(aggrTime)
}

// Destroy destroys everything (with greetings to Hatebreed)
func (u *UpdateSender) Destroy() {
	u.destroyCh <- struct{}{}
}

// AddPath adds path p for pfx to toSend queue
func (u *UpdateSender) AddPath(pfx bnet.Prefix, p *route.Path) error {
	u.toSendMu.Lock()

	hash := p.BGPPath.ComputeHash()
	if _, exists := u.toSend[hash]; exists {
		u.toSend[hash].pfxs = append(u.toSend[hash].pfxs, pfx)
		u.toSendMu.Unlock()
		return nil
	}

	u.toSend[p.BGPPath.ComputeHash()] = &pathPfxs{
		path: p,
		pfxs: []bnet.Prefix{
			pfx,
		},
	}

	u.toSendMu.Unlock()
	return nil
}

// sender serializes BGP update messages
func (u *UpdateSender) sender(aggrTime time.Duration) {
	ticker := time.NewTicker(aggrTime)
	var err error
	var pathAttrs *packet.PathAttribute
	var budget int

	for {
		select {
		case <-u.destroyCh:
			return
		case <-ticker.C:
		}

		u.toSendMu.Lock()

		overhead := u.updateOverhead()

		for key, pathNLRIs := range u.toSend {
			budget = packet.MaxLen - packet.HeaderLen - packet.MinUpdateLen - int(pathNLRIs.path.BGPPath.Length()) - overhead

			pathAttrs, err = packet.PathAttributes(pathNLRIs.path, u.iBGP, u.rrClient)
			if err != nil {
				log.Errorf("Unable to get path attributes: %v", err)
				continue
			}

			updatesPrefixes := make([][]bnet.Prefix, 0, 1)
			prefixes := make([]bnet.Prefix, 0, 1)
			for _, pfx := range pathNLRIs.pfxs {
				budget -= int(packet.BytesInAddr(pfx.Pfxlen())) + 1

				if budget < 0 {
					updatesPrefixes = append(updatesPrefixes, prefixes)
					prefixes = make([]bnet.Prefix, 0, 1)
					budget = packet.MaxLen - int(pathNLRIs.path.BGPPath.Length()) - overhead
				}

				prefixes = append(prefixes, pfx)
			}
			if len(prefixes) > 0 {
				updatesPrefixes = append(updatesPrefixes, prefixes)
			}

			delete(u.toSend, key)
			u.toSendMu.Unlock()

			u.sendUpdates(pathAttrs, updatesPrefixes, pathNLRIs.path.BGPPath.PathIdentifier)
			u.toSendMu.Lock()
		}
		u.toSendMu.Unlock()
	}
}

func (u *UpdateSender) updateOverhead() int {
	if !u.fsm.options.SupportsMultiProtocol {
		return 0
	}

	addrLen := packet.IPv4AFI
	if u.afi == packet.IPv6AFI {
		addrLen = packet.IPv6Len
	}

	// since we are replacing the next hop attribute IPv4Len has to be subtracted, we also add another byte for extended length
	return packet.AFILen + packet.SAFILen + 1 + addrLen - packet.IPv4Len + 1
}

func (u *UpdateSender) sendUpdates(pathAttrs *packet.PathAttribute, updatePrefixes [][]bnet.Prefix, pathID uint32) {
	var err error
	for _, prefixes := range updatePrefixes {
		update := u.updateMessageForPrefixes(prefixes, pathAttrs, pathID)
		if update == nil {
			log.Errorf("Failed to create update: Neighbor does not support multi protocol.")
			return
		}

		err = serializeAndSendUpdate(u.fsm.con, update, u.fsm.options)
		if err != nil {
			log.Errorf("Failed to serialize and send: %v", err)
		}
	}
}

func (u *UpdateSender) updateMessageForPrefixes(pfxs []bnet.Prefix, pa *packet.PathAttribute, pathID uint32) *packet.BGPUpdate {
	if u.afi == packet.IPv4AFI && u.safi == packet.UnicastSAFI {
		return u.bgpUpdate(pfxs, pa, pathID)
	}

	if u.fsm.options.SupportsMultiProtocol {
		return u.bgpUpdateMultiProtocol(pfxs, pa, pathID)
	}

	return nil
}

func (u *UpdateSender) bgpUpdate(pfxs []bnet.Prefix, pa *packet.PathAttribute, pathID uint32) *packet.BGPUpdate {
	update := &packet.BGPUpdate{
		PathAttributes: pa,
	}

	var nlri *packet.NLRI
	for _, pfx := range pfxs {
		nlri = &packet.NLRI{
			PathIdentifier: pathID,
			IP:             pfx.Addr().ToUint32(),
			Pfxlen:         pfx.Pfxlen(),
			Next:           update.NLRI,
		}
		update.NLRI = nlri
	}

	return update
}

func (u *UpdateSender) bgpUpdateMultiProtocol(pfxs []bnet.Prefix, pa *packet.PathAttribute, pathID uint32) *packet.BGPUpdate {
	pa, nextHop := u.copyAttributesWithoutNextHop(pa)

	attrs := &packet.PathAttribute{
		TypeCode: packet.MultiProtocolReachNLRICode,
		Value: packet.MultiProtocolReachNLRI{
			AFI:      u.afi,
			SAFI:     u.safi,
			NextHop:  nextHop,
			Prefixes: pfxs,
		},
	}
	attrs.Next = pa

	return &packet.BGPUpdate{
		PathAttributes: attrs,
	}
}

func (u *UpdateSender) copyAttributesWithoutNextHop(pa *packet.PathAttribute) (attrs *packet.PathAttribute, nextHop bnet.IP) {
	var curCopy, lastCopy *packet.PathAttribute
	for cur := pa; cur != nil; cur = cur.Next {
		if cur.TypeCode == packet.NextHopAttr {
			nextHop = cur.Value.(bnet.IP)
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
func (u *UpdateSender) RemovePath(pfx bnet.Prefix, p *route.Path) bool {
	err := u.withdrawPrefix(pfx, p)
	if err != nil {
		log.Errorf("Unable to withdraw prefix: %v", err)
		return false
	}
	return true
}

func (u *UpdateSender) withdrawPrefix(pfx bnet.Prefix, p *route.Path) error {
	if u.fsm.options.SupportsMultiProtocol {
		return withDrawPrefixesMultiProtocol(u.fsm.con, u.fsm.options, pfx, u.afi, u.safi)
	}

	return withDrawPrefixesAddPath(u.fsm.con, u.fsm.options, pfx, p)
}

// UpdateNewClient does nothing
func (u *UpdateSender) UpdateNewClient(client routingtable.RouteTableClient) error {
	log.Warningf("BGP Update Sender: UpdateNewClient not implemented")
	return nil
}

// RouteCount returns the number of stored routes
func (u *UpdateSender) RouteCount() int64 {
	log.Warningf("BGP Update Sender: RouteCount not implemented")
	return 0
}
