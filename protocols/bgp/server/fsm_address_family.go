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

	adjRIBIn  routingtable.RouteTableClient
	adjRIBOut routingtable.RouteTableClient
	rib       *locRIB.LocRIB

	importFilter *filter.Filter
	exportFilter *filter.Filter

	updateSender *UpdateSender

	addPathTX routingtable.ClientOptions
	addPathRX bool

	multiProtocol bool

	initialized bool
}

func newFSMAddressFamily(afi uint16, safi uint8, family *peerAddressFamily, fsm *FSM) *fsmAddressFamily {
	return &fsmAddressFamily{
		afi:          afi,
		safi:         safi,
		fsm:          fsm,
		rib:          family.rib,
		importFilter: family.importFilter,
		exportFilter: family.exportFilter,
	}
}

func (f *fsmAddressFamily) init(n *routingtable.Neighbor) {
	contributingASNs := f.rib.GetContributingASNs()

	f.adjRIBIn = adjRIBIn.New(f.importFilter, contributingASNs, f.fsm.peer.routerID, f.fsm.peer.clusterID, f.addPathRX)
	contributingASNs.Add(f.fsm.peer.localASN)

	f.adjRIBIn.Register(f.rib)

	f.adjRIBOut = adjRIBOut.New(n, f.exportFilter, !f.addPathTX.BestOnly)

	f.updateSender = newUpdateSender(f.fsm, f.afi, f.safi)
	f.updateSender.Start(time.Millisecond * 5)

	f.adjRIBOut.Register(f.updateSender)

	f.rib.RegisterWithOptions(f.adjRIBOut, f.addPathTX)
}

func (f *fsmAddressFamily) bmpInit() {
	f.adjRIBIn = adjRIBIn.New(filter.NewAcceptAllFilter(), &routingtable.ContributingASNs{}, f.fsm.peer.routerID, f.fsm.peer.clusterID, f.addPathRX)

	if f.rib != nil {
		f.adjRIBIn.Register(f.rib)
	}
}

func (f *fsmAddressFamily) bmpDispose() {
	f.rib.GetContributingASNs().Remove(f.fsm.peer.localASN)

	f.adjRIBIn.(*adjRIBIn.AdjRIBIn).Flush()

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

func (f *fsmAddressFamily) processUpdate(u *packet.BGPUpdate) {
	if f.safi != packet.UnicastSAFI {
		return
	}

	f.multiProtocolUpdates(u)
	if f.afi == packet.IPv4AFI {
		f.withdraws(u)
		f.updates(u)
	}
}

func (f *fsmAddressFamily) withdraws(u *packet.BGPUpdate) {
	for r := u.WithdrawnRoutes; r != nil; r = r.Next {
		f.adjRIBIn.RemovePath(r.Prefix, nil)
	}
}

func (f *fsmAddressFamily) updates(u *packet.BGPUpdate) {
	for r := u.NLRI; r != nil; r = r.Next {
		path := f.newRoutePath()
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
			Source: f.fsm.peer.addr,
			EBGP:   f.fsm.peer.localASN != f.fsm.peer.peerASN,
		},
	}
}

func (f *fsmAddressFamily) multiProtocolUpdate(path *route.Path, nlri packet.MultiProtocolReachNLRI) {
	if f.afi != nlri.AFI || f.safi != nlri.SAFI {
		return
	}

	path.BGPPath.NextHop = nlri.NextHop

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
			path.BGPPath.Origin = pa.Value.(uint8)
		case packet.LocalPrefAttr:
			path.BGPPath.LocalPref = pa.Value.(uint32)
		case packet.MEDAttr:
			path.BGPPath.MED = pa.Value.(uint32)
		case packet.NextHopAttr:
			path.BGPPath.NextHop = pa.Value.(bnet.IP)
		case packet.ASPathAttr:
			path.BGPPath.ASPath = pa.Value.(types.ASPath)
			path.BGPPath.ASPathLen = path.BGPPath.ASPath.Length()
		case packet.AggregatorAttr:
			aggr := pa.Value.(types.Aggregator)
			path.BGPPath.Aggregator = &aggr
		case packet.AtomicAggrAttr:
			path.BGPPath.AtomicAggregate = true
		case packet.CommunitiesAttr:
			path.BGPPath.Communities = pa.Value.([]uint32)
		case packet.LargeCommunitiesAttr:
			path.BGPPath.LargeCommunities = pa.Value.([]types.LargeCommunity)
		case packet.OriginatorIDAttr:
			path.BGPPath.OriginatorID = pa.Value.(uint32)
		case packet.ClusterListAttr:
			path.BGPPath.ClusterList = pa.Value.([]uint32)
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
