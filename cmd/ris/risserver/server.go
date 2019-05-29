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

// LPM provides a longest prefix match service
func (s *Server) LPM(ctx context.Context, req *pb.LPMRequest) (*pb.LPMResponse, error) {
	r := s.bmp.GetRouter(req.Router)
	if r == nil {
		return nil, fmt.Errorf("Unable to get router %q", req.Router)
	}

	v := r.GetVRF(req.VrfId)
	if v == nil {
		return nil, fmt.Errorf("Unable to get VRF %d", req.VrfId)
	}

	if req.Pfx == nil {
		return nil, fmt.Errorf("Not prefix given")
	}

	pfx := bnet.NewPrefixFromProtoPrefix(*req.Pfx)
	var rib *locRIB.LocRIB
	switch req.Pfx.Address.Version {
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

	routes := rib.LPM(pfx)
	res := &pb.LPMResponse{
		Routes: make([]*routeapi.Route, 0, len(routes)),
	}
	for _, route := range routes {
		res.Routes = append(res.Routes, route.ToProto())
	}

	return res, nil
}
