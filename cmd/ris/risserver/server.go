package risserver

import (
	"context"
	"fmt"

	"github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"

	pb "github.com/bio-routing/bio-rd/cmd/ris/api"
	bnet "github.com/bio-routing/bio-rd/net"
	netapi "github.com/bio-routing/bio-rd/net/api"
	routeapi "github.com/bio-routing/bio-rd/route/api"
)

// Server represents an RoutingInformationService server
type Server struct {
	bmp *server.BMPServer
}

// NewServer creates a new server
func NewServer(b *server.BMPServer) *Server {
	return &Server{
		bmp: b,
	}
}

func (s Server) getRIB(rtr string, vrfID uint64, p *netapi.Prefix) (*locRIB.LocRIB, error) {
	r := s.bmp.GetRouter(rtr)
	if r == nil {
		return nil, fmt.Errorf("Unable to get router %q", rtr)
	}

	v := r.GetVRF(vrfID)
	if v == nil {
		return nil, fmt.Errorf("Unable to get VRF %d", vrfID)
	}

	if p == nil {
		return nil, fmt.Errorf("Not prefix given")
	}

	var rib *locRIB.LocRIB
	switch p.Address.Version {
	case netapi.IP_IPv4:
		rib = v.IPv4UnicastRIB()
	case netapi.IP_IPv6:
		rib = v.IPv6UnicastRIB()
	default:
		return nil, fmt.Errorf("Unknown afi")
	}

	if rib == nil {
		return nil, fmt.Errorf("Unable to get RIB")
	}

	return rib, nil
}

// LPM provides a longest prefix match service
func (s *Server) LPM(ctx context.Context, req *pb.LPMRequest) (*pb.LPMResponse, error) {
	rib, err := s.getRIB(req.Router, req.VrfId, req.Pfx)
	if err != nil {
		return nil, err
	}

	routes := rib.LPM(bnet.NewPrefixFromProtoPrefix(*req.Pfx))
	res := &pb.LPMResponse{
		Routes: make([]*routeapi.Route, 0, len(routes)),
	}
	for _, route := range routes {
		res.Routes = append(res.Routes, route.ToProto())
	}

	return res, nil
}

// Get gets a prefix (exact match)
func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	rib, err := s.getRIB(req.Router, req.VrfId, req.Pfx)
	if err != nil {
		return nil, err
	}

	route := rib.Get(bnet.NewPrefixFromProtoPrefix(*req.Pfx))
	if route == nil {
		return &pb.GetResponse{
			Routes: make([]*routeapi.Route, 0, 0),
		}, nil
	}

	return &pb.GetResponse{
		Routes: []*routeapi.Route{
			route.ToProto(),
		},
	}, nil
}

// GetLonger gets all more specifics of a prefix
func (s *Server) GetLonger(ctx context.Context, req *pb.GetLongerRequest) (*pb.GetLongerResponse, error) {
	rib, err := s.getRIB(req.Router, req.VrfId, req.Pfx)
	if err != nil {
		return nil, err
	}

	routes := rib.GetLonger(bnet.NewPrefixFromProtoPrefix(*req.Pfx))
	res := &pb.GetLongerResponse{
		Routes: make([]*routeapi.Route, 0, len(routes)),
	}
	for _, route := range routes {
		res.Routes = append(res.Routes, route.ToProto())
	}

	return res, nil
}

func (s *Server) AdjRIBInStream(req *pb.AdjRIBInStreamRequest, srv pb.RoutingInformationService_AdjRIBInStreamServer) error {
	return nil
}
