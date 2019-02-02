package server

import (
	"context"
	"fmt"

	"github.com/bio-routing/bio-rd/protocols/bgp/api"
	"github.com/bio-routing/bio-rd/route"
	"google.golang.org/grpc"

	bnet "github.com/bio-routing/bio-rd/net"
	routeapi "github.com/bio-routing/bio-rd/route/api"
)

type BGPAPIServer struct {
	srv BGPServer
}

func (s *BGPAPIServer) ListSessions(ctx context.Context, in *api.ListSessionsRequest, opts ...grpc.CallOption) (*api.ListSessionsResponse, error) {
	return nil, fmt.Errorf("Not implemented yet.")
}

// DumpRIBIn dumps the RIB in of a peer for a given AFI/SAFI
func (s *BGPAPIServer) DumpRIBIn(ctx context.Context, in *api.DumpRIBRequest, opts ...grpc.CallOption) (*api.DumpRIBResponse, error) {
	dump := s.srv.DumpRIBIn(bnet.IPFromProtoIP(*in.Peer), uint16(in.Afi), uint8(in.Safi))

	return &api.DumpRIBResponse{
		Routes: routesToProto(dump),
	}, nil
}

// DumpRIBOut dumps the RIB out of a peer for a given AFI/SAFI
func (s *BGPAPIServer) DumpRIBOut(ctx context.Context, in *api.DumpRIBRequest, opts ...grpc.CallOption) (*api.DumpRIBResponse, error) {
	dump := s.srv.DumpRIBOut(bnet.IPFromProtoIP(*in.Peer), uint16(in.Afi), uint8(in.Safi))

	return &api.DumpRIBResponse{
		Routes: routesToProto(dump),
	}, nil
}

func routesToProto(dump []*route.Route) []*routeapi.Route {
	routes := make([]*routeapi.Route, len(dump))
	for i := range dump {
		routes[i] = dump[i].ToProto()
	}

	return routes
}
