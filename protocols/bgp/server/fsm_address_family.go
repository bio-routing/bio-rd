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

	initialized bool
}

func newFSMAddressFamily(afi uint16, safi uint8, params *familyParameters, fsm *FSM) *fsmAddressFamily {
	return &fsmAddressFamily{
		afi:          afi,
		safi:         safi,
		fsm:          fsm,
		rib:          params.rib,
		importFilter: params.importFilter,
		exportFilter: params.exportFilter,
	}
}

func (f *fsmAddressFamily) init(n *routingtable.Neighbor) {
	contributingASNs := f.rib.GetContributingASNs()

	f.adjRIBIn = adjRIBIn.New(f.importFilter, contributingASNs, f.fsm.peer.routerID, f.fsm.peer.clusterID)
	contributingASNs.Add(f.fsm.peer.localASN)
	f.adjRIBIn.Register(f.rib)

	f.adjRIBOut = adjRIBOut.New(n, f.exportFilter)
	clientOptions := routingtable.ClientOptions{
		BestOnly: true,
	}
	if f.fsm.options.AddPathRX {
		clientOptions = f.fsm.peer.addPathSend
	}

	f.updateSender = newUpdateSender(f.fsm, f.afi, f.safi)
	f.updateSender.Start(time.Millisecond * 5)

	f.adjRIBOut.Register(f.updateSender)
	f.rib.RegisterWithOptions(f.adjRIBOut, clientOptions)
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
	if f.afi == packet.IPv4AFI && f.safi == packet.UnicastSAFI {
		f.withdraws(u)
		f.updates(u)
	}

	if f.fsm.options.SupportsMultiProtocol {
		f.multiProtocolUpdates(u)
	}
}

func (f *fsmAddressFamily) withdraws(u *packet.BGPUpdate) {
	for r := u.WithdrawnRoutes; r != nil; r = r.Next {
		pfx := bnet.NewPfx(bnet.IPv4(r.IP), r.Pfxlen)
		f.adjRIBIn.RemovePath(pfx, nil)
	}
}

func (f *fsmAddressFamily) updates(u *packet.BGPUpdate) {
	for r := u.NLRI; r != nil; r = r.Next {
		pfx := bnet.NewPfx(bnet.IPv4(r.IP), r.Pfxlen)

		path := f.newRoutePath()
		f.processAttributes(u.PathAttributes, path)

		f.adjRIBIn.AddPath(pfx, path)
	}
}

func (f *fsmAddressFamily) multiProtocolUpdates(u *packet.BGPUpdate) {
	if !f.fsm.options.SupportsMultiProtocol {
		return
	}

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
	path.BGPPath.NextHop = nlri.NextHop

	for _, pfx := range nlri.Prefixes {
		f.adjRIBIn.AddPath(pfx, path)
	}
}

func (f *fsmAddressFamily) multiProtocolWithdraw(path *route.Path, nlri packet.MultiProtocolUnreachNLRI) {
	for _, pfx := range nlri.Prefixes {
		f.adjRIBIn.RemovePath(pfx, path)
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
