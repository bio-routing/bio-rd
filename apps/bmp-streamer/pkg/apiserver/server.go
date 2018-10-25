package apiserver

import (
	"fmt"

	pb "github.com/bio-routing/bio-rd/apps/bmp-streamer/pkg/bmpsrvapi"
	net "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
)

// APIServer implements the BMP server API
type APIServer struct {
	bmpServer *server.BMPServer
}

// New creates an new API server
func New(bmpServer *server.BMPServer) *APIServer {
	return &APIServer{
		bmpServer: bmpServer,
	}
}

func (u update) toRIBUpdate() *pb.RIBUpdate {
	toSend := &pb.RIBUpdate{
		Peer: &pb.IP{},
		Route: &pb.Route{
			Pfx: &pb.Prefix{
				Address: &pb.IP{},
			},
			Path: &pb.Path{
				Type: route.BGPPathType,
				BGPPath: &pb.BGPPath{
					NextHop: &pb.IP{},
					Source:  &pb.IP{},
				},
			},
		},
	}

	toSend.Advertisement = u.advertisement
	toSend.Peer.Lower = u.path.BGPPath.Source.Lower()
	toSend.Peer.Higher = u.path.BGPPath.Source.Higher()
	toSend.Peer.IPVersion = uint32(u.path.BGPPath.Source.IPVersion())
	toSend.Route.Pfx.Pfxlen = uint32(u.prefix.Pfxlen())
	toSend.Route.Pfx.Address.Lower = u.prefix.Addr().Lower()
	toSend.Route.Pfx.Address.Higher = u.prefix.Addr().Higher()
	toSend.Route.Pfx.Address.IPVersion = uint32(u.prefix.Addr().IPVersion())
	toSend.Route.Path.BGPPath.NextHop.Lower = u.path.BGPPath.NextHop.Lower()
	toSend.Route.Path.BGPPath.NextHop.Higher = u.path.BGPPath.NextHop.Higher()
	toSend.Route.Path.BGPPath.NextHop.IPVersion = uint32(u.path.BGPPath.NextHop.IPVersion())
	toSend.Route.Path.BGPPath.Source.Lower = u.path.BGPPath.Source.Lower()
	toSend.Route.Path.BGPPath.Source.Higher = u.path.BGPPath.Source.Higher()
	toSend.Route.Path.BGPPath.Source.IPVersion = uint32(u.path.BGPPath.Source.IPVersion())
	toSend.Route.Path.BGPPath.EBGP = u.path.BGPPath.EBGP
	toSend.Route.Path.BGPPath.BGPIdentifier = u.path.BGPPath.BGPIdentifier
	toSend.Route.Path.BGPPath.ClusterList = u.path.BGPPath.ClusterList
	toSend.Route.Path.BGPPath.Communities = u.path.BGPPath.Communities
	toSend.Route.Path.BGPPath.LocalPref = u.path.BGPPath.LocalPref
	toSend.Route.Path.BGPPath.MED = u.path.BGPPath.MED
	toSend.Route.Path.BGPPath.PathIdentifier = u.path.BGPPath.PathIdentifier
	toSend.Route.Path.BGPPath.Origin = uint32(u.path.BGPPath.Origin)

	toSend.Route.Path.BGPPath.LargeCommunities = make([]*pb.LargeCommunity, len(u.path.BGPPath.LargeCommunities))
	for i, com := range u.path.BGPPath.LargeCommunities {
		toSend.Route.Path.BGPPath.LargeCommunities[i] = &pb.LargeCommunity{
			GlobalAdministrator: com.GlobalAdministrator,
			DataPart1:           com.DataPart1,
			DataPart2:           com.DataPart2,
		}
	}

	toSend.Route.Path.BGPPath.ASPath = make([]*pb.ASPathSegment, len(u.path.BGPPath.ASPath))
	for i, pathSegment := range u.path.BGPPath.ASPath {
		newSegment := &pb.ASPathSegment{
			ASSequence: pathSegment.Type == types.ASSequence,
			ASNs:       make([]uint32, len(pathSegment.ASNs)),
		}
		copy(newSegment.ASNs, pathSegment.ASNs)
		toSend.Route.Path.BGPPath.ASPath[i] = newSegment
	}

	toSend.Route.Path.BGPPath.UnknownAttributes = make([]*pb.UnknownAttribute, len(u.path.BGPPath.UnknownAttributes))
	for i, attr := range u.path.BGPPath.UnknownAttributes {
		toSend.Route.Path.BGPPath.UnknownAttributes[i] = &pb.UnknownAttribute{
			Optional:   attr.Optional,
			Transitive: attr.Transitive,
			Partial:    attr.Partial,
			TypeCode:   uint32(attr.TypeCode),
			Value:      attr.Value,
		}
	}

	return toSend
}

// AdjRIBInStream offers RIBs as a stream
func (a *APIServer) AdjRIBInStream(req *pb.AdjRIBInStreamRequest, stream pb.RIBService_AdjRIBInStreamServer) error {
	r4 := newRIBClient()
	r6 := newRIBClient()

	addr := net.IP{}
	if req.Router.IPVersion == 4 {
		addr = net.IPv4(uint32(req.Router.Lower))
	} else if req.Router.IPVersion == 6 {
		addr = net.IPv6(req.Router.Higher, req.Router.Lower)
	} else {
		return fmt.Errorf("Unknown IP version: %d", req.Router.IPVersion)
	}

	ret := make(chan error)
	go func() {
		var err error
		u := update{}
		for {
			select {
			case u = <-r4.ch:
			case u = <-r6.ch:
			}

			toSend := u.toRIBUpdate()

			err = stream.Send(toSend)
			if err != nil {
				ret <- err
				return
			}
		}
	}()

	a.bmpServer.SubscribeRIBs(r4, addr.ToNetIP(), packet.IPv4AFI)
	defer a.bmpServer.UnsubscribeRIBs(r4, addr.ToNetIP(), packet.IPv4AFI)

	a.bmpServer.SubscribeRIBs(r6, addr.ToNetIP(), packet.IPv6AFI)
	defer a.bmpServer.UnsubscribeRIBs(r6, addr.ToNetIP(), packet.IPv6AFI)

	err := <-ret
	if err != nil {
		return fmt.Errorf("Stream ended: %v", err)
	}

	return nil
}

type update struct {
	advertisement bool
	prefix        net.Prefix
	path          *route.Path
}

type ribClient struct {
	ch chan update
}

func newRIBClient() *ribClient {
	return &ribClient{
		ch: make(chan update),
	}
}

func (r *ribClient) AddPath(pfx net.Prefix, path *route.Path) error {
	r.ch <- update{
		advertisement: true,
		prefix:        pfx,
		path:          path,
	}

	return nil
}

func (r *ribClient) RemovePath(pfx net.Prefix, path *route.Path) bool {
	r.ch <- update{
		advertisement: false,
		prefix:        pfx,
		path:          path,
	}

	return false
}

func (r *ribClient) UpdateNewClient(routingtable.RouteTableClient) error {
	return nil
}

func (r *ribClient) Register(routingtable.RouteTableClient) {
}

func (r *ribClient) RegisterWithOptions(routingtable.RouteTableClient, routingtable.ClientOptions) {
}

func (r *ribClient) Unregister(routingtable.RouteTableClient) {
}

func (r *ribClient) RouteCount() int64 {
	return -1
}
