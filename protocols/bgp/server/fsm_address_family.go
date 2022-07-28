package server

import (
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

	initialized bool
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
	New(exportFilterChain filter.Chain, contributingASNs *routingtable.ContributingASNs, routerID uint32, clusterID uint32, addPathRX bool) routingtable.AdjRIBIn
}

type adjRIBInFactory struct{}

func (a adjRIBInFactory) New(exportFilterChain filter.Chain, contributingASNs *routingtable.ContributingASNs, routerID uint32, clusterID uint32, addPathRX bool) routingtable.AdjRIBIn {
	return adjRIBIn.New(exportFilterChain, contributingASNs, routerID, clusterID, addPathRX)
}

func (f *fsmAddressFamily) init(n *routingtable.Neighbor) {
	contributingASNs := f.rib.GetContributingASNs()

	f.adjRIBIn = f.fsm.peer.adjRIBInFactory.New(f.importFilterChain, contributingASNs, f.fsm.peer.routerID, f.fsm.peer.clusterID, f.addPathRX)
	contributingASNs.Add(f.fsm.peer.localASN)

	f.adjRIBIn.Register(f.rib)

	f.adjRIBOut = adjRIBOut.New(f.rib, n, f.exportFilterChain, !f.addPathTX.BestOnly)

	f.updateSender = newUpdateSender(f)
	f.updateSender.Start(time.Millisecond * 5)

	f.adjRIBOut.Register(f.updateSender)

	f.rib.RegisterWithOptions(f.adjRIBOut, f.addPathTX)
	f.initialized = true
}

func (f *fsmAddressFamily) bmpInit() {
	f.adjRIBIn = f.fsm.peer.adjRIBInFactory.New(filter.NewAcceptAllFilterChain(), &routingtable.ContributingASNs{}, f.fsm.peer.routerID, f.fsm.peer.clusterID, f.addPathRX)

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

func (f *fsmAddressFamily) processUpdate(u *packet.BGPUpdate, postPolicy bool) {
	if f.safi != packet.SAFIUnicast {
		return
	}

	f.multiProtocolUpdates(u)
	if f.afi == packet.AFIIPv4 {
		f.withdraws(u, postPolicy)
		f.updates(u, postPolicy)
	}
}

func (f *fsmAddressFamily) withdraws(u *packet.BGPUpdate, postPolicy bool) {
	for r := u.WithdrawnRoutes; r != nil; r = r.Next {
		f.adjRIBIn.RemovePath(r.Prefix, &route.Path{
			PostPolicy: postPolicy,
		})
	}
}

func (f *fsmAddressFamily) updates(u *packet.BGPUpdate, postPolicy bool) {
	for r := u.NLRI; r != nil; r = r.Next {
		path := f.newRoutePath()
		path.PostPolicy = postPolicy
		f.processAttributes(u.PathAttributes, path)

		f.adjRIBIn.AddPath(r.Prefix, path)
	}
}

func (f *fsmAddressFamily) multiProtocolUpdates(u *packet.BGPUpdate) {
	path := f.newRoutePath()
	f.processAttributes(u.PathAttributes, path)

	for pa := u.PathAttributes; pa != nil; pa = pa.Next {
		switch pa.TypeCode {
		case packet.MultiProtocolReachNLRICode:
			f.multiProtocolUpdate(path, pa.Value.(packet.MultiProtocolReachNLRI))
		case packet.MultiProtocolUnreachNLRICode:
			f.multiProtocolWithdraw(path, pa.Value.(packet.MultiProtocolUnreachNLRI))
		}
	}
}

func (f *fsmAddressFamily) newRoutePath() *route.Path {
	return &route.Path{
		Type: route.BGPPathType,
		BGPPath: &route.BGPPath{
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

	path.BGPPath.BGPPathA.NextHop = nlri.NextHop

	for n := nlri.NLRI; n != nil; n = n.Next {
		f.adjRIBIn.AddPath(n.Prefix, path)
	}
}

func (f *fsmAddressFamily) multiProtocolWithdraw(path *route.Path, nlri packet.MultiProtocolUnreachNLRI) {
	if f.afi != nlri.AFI || f.safi != nlri.SAFI {
		return
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
		case packet.MultiProtocolReachNLRICode:
		case packet.MultiProtocolUnreachNLRICode:
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
