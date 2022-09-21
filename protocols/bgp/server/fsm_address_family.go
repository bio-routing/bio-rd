package server

import (
	"sync/atomic"
	"time"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/adjRIBIn"
	"github.com/bio-routing/bio-rd/routingtable/adjRIBOut"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
)

// fsmAddressFamily holds RIBs and the UpdateSender of an peer for an AFI/SAFI combination
type fsmAddressFamily struct {
	afi  uint16
	safi uint8
	fsm  *FSM

	adjRIBIn  routingtable.AdjRIBIn
	adjRIBOut routingtable.AdjRIBOut
	rib       *locRIB.LocRIB

	importFilterChain filter.Chain
	exportFilterChain filter.Chain

	updateSender *UpdateSender

	addPathTX routingtable.ClientOptions
	addPathRX bool

	multiProtocol bool

	initialized            bool
	endOfRIBMarkerReceived atomic.Bool
}

func newFSMAddressFamily(afi uint16, safi uint8, family *peerAddressFamily, fsm *FSM) *fsmAddressFamily {
	return &fsmAddressFamily{
		afi:               afi,
		safi:              safi,
		fsm:               fsm,
		rib:               family.rib,
		importFilterChain: family.importFilterChain,
		exportFilterChain: family.exportFilterChain,
		addPathTX: routingtable.ClientOptions{
			BestOnly: true,
		},
	}
}

func (f *fsmAddressFamily) replaceImportFilterChain(c filter.Chain) {
	if c.Equal(f.importFilterChain) {
		return
	}

	f.importFilterChain = c
	f.adjRIBIn.ReplaceFilterChain(c)
}

func (f *fsmAddressFamily) replaceExportFilterChain(c filter.Chain) {
	if c.Equal(f.exportFilterChain) {
		return
	}

	f.exportFilterChain = c
	f.adjRIBOut.ReplaceFilterChain(c)
}

func (f *fsmAddressFamily) dumpRIBOut() []*route.Route {
	return f.adjRIBOut.Dump()
}

func (f *fsmAddressFamily) dumpRIBIn() []*route.Route {
	return f.adjRIBIn.Dump()
}

type adjRIBInFactoryI interface {
	New(exportFilterChain filter.Chain, contributingASNs *routingtable.ContributingASNs, sessionAttrs routingtable.SessionAttrs) routingtable.AdjRIBIn
}

type adjRIBInFactory struct{}

func (a adjRIBInFactory) New(exportFilterChain filter.Chain, contributingASNs *routingtable.ContributingASNs, sessionAttrs routingtable.SessionAttrs) routingtable.AdjRIBIn {
	return adjRIBIn.New(exportFilterChain, contributingASNs, sessionAttrs)
}

func (f *fsmAddressFamily) init() {
	contributingASNs := f.rib.GetContributingASNs()
	sessionAttrs := f.getSessionAttrs()

	f.adjRIBIn = f.fsm.peer.adjRIBInFactory.New(f.importFilterChain, contributingASNs, sessionAttrs)
	contributingASNs.Add(f.fsm.peer.localASN)

	f.adjRIBIn.Register(f.rib)

	f.adjRIBOut = adjRIBOut.New(f.rib, sessionAttrs, f.exportFilterChain)

	f.updateSender = newUpdateSender(f)
	f.updateSender.Start(time.Millisecond * 5)

	f.adjRIBOut.Register(f.updateSender)

	f.rib.RegisterWithOptions(f.adjRIBOut, f.addPathTX)
	f.initialized = true
}

func (f *fsmAddressFamily) getSessionAttrs() routingtable.SessionAttrs {
	rip, _ := bnet.IPFromBytes(f.fsm.bmpRouterAddress)

	return routingtable.SessionAttrs{
		RouterID:             f.fsm.peer.routerID,
		PeerIP:               f.fsm.peer.addr,
		LocalIP:              f.fsm.peer.localAddr,
		Type:                 route.BGPPathType,
		IBGP:                 f.fsm.peer.localASN == f.fsm.peer.peerASN,
		LocalASN:             f.fsm.peer.localASN,
		PeerASN:              f.fsm.peer.peerASN,
		RouteServerClient:    f.fsm.peer.routeServerClient,
		RouteReflectorClient: f.fsm.peer.routeReflectorClient,
		ClusterID:            f.fsm.peer.clusterID,
		AddPathRX:            f.addPathRX,
		AddPathTX:            !f.addPathTX.BestOnly,

		PeerRoleEnabled:    f.fsm.peer.peerRoleEnabled,
		PeerRoleStrictMode: f.fsm.peer.peerRoleStrictMode,
		PeerRoleLocal:      f.fsm.peer.peerRoleLocal,
		PeerRoleAdvByPeer:  f.fsm.peer.peerRoleAdvByPeer,
		PeerRoleRemote:     f.fsm.peer.peerRoleRemote,

		// Only relevant for BMP use
		RouterIP: rip,
	}
}

func (f *fsmAddressFamily) bmpInit() {
	f.adjRIBIn = f.fsm.peer.adjRIBInFactory.New(filter.NewAcceptAllFilterChain(), &routingtable.ContributingASNs{}, f.getSessionAttrs())

	if f.rib != nil {
		f.adjRIBIn.Register(f.rib)
	}

	f.initialized = true
}

func (f *fsmAddressFamily) bmpDispose() {
	f.rib.GetContributingASNs().Remove(f.fsm.peer.localASN)

	f.adjRIBIn.Flush()

	f.adjRIBIn.Unregister(f.rib)

	f.adjRIBIn = nil
}

func (f *fsmAddressFamily) dispose() {
	if !f.initialized {
		return
	}

	f.rib.GetContributingASNs().Remove(f.fsm.peer.localASN)
	f.adjRIBIn.Unregister(f.rib)
	f.rib.Unregister(f.adjRIBOut)
	f.adjRIBOut.Unregister(f.updateSender)
	f.updateSender.Destroy()

	f.adjRIBIn = nil
	f.adjRIBOut = nil

	f.initialized = false
}

func (f *fsmAddressFamily) processUpdate(u *packet.BGPUpdate, bmpPostPolicy bool, timestamp uint32) {
	if f.safi != packet.SAFIUnicast {
		return
	}

	f.multiProtocolUpdates(u, bmpPostPolicy, timestamp)
	if f.afi == packet.AFIIPv4 {
		if u.IsEndOfRIBMarker() {
			f.endOfRIBMarkerReceived.Store(true)
		}

		f.withdraws(u, bmpPostPolicy, timestamp)
		f.updates(u, bmpPostPolicy, timestamp)
	}
}

func (f *fsmAddressFamily) withdraws(u *packet.BGPUpdate, bmpPostPolicy bool, timestamp uint32) {
	for r := u.WithdrawnRoutes; r != nil; r = r.Next {
		f.adjRIBIn.RemovePath(r.Prefix, &route.Path{
			LTime: timestamp,
			BGPPath: &route.BGPPath{
				BMPPostPolicy:  bmpPostPolicy,
				PathIdentifier: r.PathIdentifier,
			},
		})
	}
}

func (f *fsmAddressFamily) updates(u *packet.BGPUpdate, bmpPostPolicy bool, timestamp uint32) {
	for r := u.NLRI; r != nil; r = r.Next {
		path := f.newRoutePath(bmpPostPolicy, timestamp)
		f.processAttributes(u.PathAttributes, path)
		path.BGPPath.PathIdentifier = u.NLRI.PathIdentifier

		f.adjRIBIn.AddPath(r.Prefix, path)
	}
}

func (f *fsmAddressFamily) multiProtocolUpdates(u *packet.BGPUpdate, bmpPostPolicy bool, timestamp uint32) {
	path := f.newRoutePath(bmpPostPolicy, timestamp)
	f.processAttributes(u.PathAttributes, path)

	mpReachNLRI, mpUnreachNLRI := getMPReachAndUnreachNLRIs(u)

	if mpReachNLRI != nil {
		f.multiProtocolUpdate(path, *mpReachNLRI)
	}

	if mpUnreachNLRI != nil {
		f.multiProtocolWithdraw(path, *mpUnreachNLRI)
	}

	if mpReachNLRI != nil && mpUnreachNLRI != nil {
		if mpReachNLRI.NLRI == nil && mpUnreachNLRI.NLRI == nil {
			f.endOfRIBMarkerReceived.Store(true)
		}
	}
}

func getMPReachAndUnreachNLRIs(u *packet.BGPUpdate) (reach *packet.MultiProtocolReachNLRI, unreach *packet.MultiProtocolUnreachNLRI) {
	for pa := u.PathAttributes; pa != nil; pa = pa.Next {
		if pa.TypeCode == packet.MultiProtocolReachNLRIAttr {
			r := pa.Value.(packet.MultiProtocolReachNLRI)
			reach = &r
		}

		if pa.TypeCode == packet.MultiProtocolUnreachNLRIAttr {
			ur := pa.Value.(packet.MultiProtocolUnreachNLRI)
			unreach = &ur
		}
	}

	return reach, unreach
}

func (f *fsmAddressFamily) newRoutePath(bmpPostPolicy bool, timestamp uint32) *route.Path {
	return &route.Path{
		LTime: timestamp,
		Type:  route.BGPPathType,
		BGPPath: &route.BGPPath{
			BMPPostPolicy: bmpPostPolicy,
			BGPPathA: &route.BGPPathA{
				Source: f.fsm.peer.addr,
				EBGP:   f.fsm.peer.localASN != f.fsm.peer.peerASN,
			},
		},
	}
}

func (f *fsmAddressFamily) multiProtocolUpdate(path *route.Path, nlri packet.MultiProtocolReachNLRI) {
	if f.afi != nlri.AFI || f.safi != nlri.SAFI {
		return
	}

	path.BGPPath.PathIdentifier = nlri.NLRI.PathIdentifier
	path.BGPPath.BGPPathA.NextHop = nlri.NextHop

	for n := nlri.NLRI; n != nil; n = n.Next {
		f.adjRIBIn.AddPath(n.Prefix, path)
	}
}

func (f *fsmAddressFamily) multiProtocolWithdraw(path *route.Path, nlri packet.MultiProtocolUnreachNLRI) {
	if f.afi != nlri.AFI || f.safi != nlri.SAFI {
		return
	}

	if nlri.NLRI != nil {
		path.BGPPath.PathIdentifier = nlri.NLRI.PathIdentifier
	}

	for cur := nlri.NLRI; cur != nil; cur = cur.Next {
		f.adjRIBIn.RemovePath(cur.Prefix, path)
	}
}

func (f *fsmAddressFamily) processAttributes(attrs *packet.PathAttribute, path *route.Path) {
	for pa := attrs; pa != nil; pa = pa.Next {
		switch pa.TypeCode {
		case packet.OriginAttr:
			path.BGPPath.BGPPathA.Origin = pa.Value.(uint8)
		case packet.LocalPrefAttr:
			path.BGPPath.BGPPathA.LocalPref = pa.Value.(uint32)
		case packet.MEDAttr:
			path.BGPPath.BGPPathA.MED = pa.Value.(uint32)
		case packet.NextHopAttr:
			path.BGPPath.BGPPathA.NextHop = pa.Value.(*bnet.IP)
		case packet.ASPathAttr:
			path.BGPPath.ASPath = pa.Value.(*types.ASPath)
			path.BGPPath.ASPathLen = path.BGPPath.ASPath.Length()
		case packet.AggregatorAttr:
			aggr := pa.Value.(types.Aggregator)
			path.BGPPath.BGPPathA.Aggregator = &aggr
		case packet.AtomicAggrAttr:
			path.BGPPath.BGPPathA.AtomicAggregate = true
		case packet.CommunitiesAttr:
			path.BGPPath.Communities = pa.Value.(*types.Communities)
		case packet.LargeCommunitiesAttr:
			path.BGPPath.LargeCommunities = pa.Value.(*types.LargeCommunities)
		case packet.OriginatorIDAttr:
			path.BGPPath.BGPPathA.OriginatorID = pa.Value.(uint32)
		case packet.ClusterListAttr:
			path.BGPPath.ClusterList = pa.Value.(*types.ClusterList)
		case packet.MultiProtocolReachNLRIAttr:
		case packet.MultiProtocolUnreachNLRIAttr:
		default:
			unknownAttr := f.processUnknownAttribute(pa)
			if unknownAttr != nil {
				path.BGPPath.UnknownAttributes = append(path.BGPPath.UnknownAttributes, *unknownAttr)
			}
		}
	}
}

func (f *fsmAddressFamily) processUnknownAttribute(attr *packet.PathAttribute) *types.UnknownPathAttribute {
	if !attr.Transitive {
		return nil
	}

	u := &types.UnknownPathAttribute{
		Transitive: true,
		Optional:   attr.Optional,
		Partial:    attr.Partial,
		TypeCode:   attr.TypeCode,
		Value:      attr.Value.([]byte),
	}

	return u
}
