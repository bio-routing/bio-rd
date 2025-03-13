package risserver

import (
	"context"
	"fmt"

	"github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/prometheus/client_golang/prometheus"

	pb "github.com/bio-routing/bio-rd/cmd/ris/api"
	bnet "github.com/bio-routing/bio-rd/net"
	netapi "github.com/bio-routing/bio-rd/net/api"
	routeapi "github.com/bio-routing/bio-rd/route/api"
)

var risObserveFIBClients *prometheus.GaugeVec

func init() {
	risObserveFIBClients = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "bio",
			Subsystem: "ris",
			Name:      "observe_fib_clients",
			Help:      "number of observe FIB clients per router/vrf/afisafi",
		},
		[]string{
			"router",
			"vrf",
			"afisafi",
		},
	)
	prometheus.MustRegister(risObserveFIBClients)
}

// Server represents an RoutingInformationService server
type Server struct {
	pb.UnimplementedRoutingInformationServiceServer
	bmp server.BMPReceiverInterface
}

// NewServer creates a new server
func NewServer(b server.BMPReceiverInterface) *Server {
	return &Server{
		bmp: b,
	}
}

func wrapGetRIBErr(err error, rtr string, vrfID uint64, version netapi.IP_Version) error {
	return fmt.Errorf("unable to get RIB (%s/%s/v%d): %w", rtr, vrf.RouteDistinguisherHumanReadable(vrfID), version, err)
}

func wrapRIBNotReadyErr(err error, rtr string, vrfID uint64, version netapi.IP_Version) error {
	return fmt.Errorf("RIB not ready yet (%s/%s/v%d): %w", rtr, vrf.RouteDistinguisherHumanReadable(vrfID), version, err)
}

func ipVersionFromProto(v netapi.IP_Version) uint16 {
	switch v {
	case netapi.IP_IPv4:
		return 4
	case netapi.IP_IPv6:
		return 6
	}

	return 0
}

func (s Server) getRIB(rtr string, vrfID uint64, ipVersion netapi.IP_Version) (*locRIB.LocRIB, error) {
	r := s.bmp.GetRouter(rtr)
	if r == nil {
		return nil, fmt.Errorf("unable to get router")
	}

	v := r.GetVRF(vrfID)
	if v == nil {
		return nil, fmt.Errorf("unable to get VRF")
	}

	var rib *locRIB.LocRIB
	switch ipVersion {
	case netapi.IP_IPv4:
		rib = v.IPv4UnicastRIB()
	case netapi.IP_IPv6:
		rib = v.IPv6UnicastRIB()
	default:
		return nil, fmt.Errorf("unknown afi")
	}

	if rib == nil {
		return nil, fmt.Errorf("unable to get RIB")
	}

	return rib, nil
}

// LPM provides a longest prefix match service
func (s *Server) LPM(ctx context.Context, req *pb.LPMRequest) (*pb.LPMResponse, error) {
	vrfID, err := getVRFID(req)
	if err != nil {
		return nil, err
	}

	rib, err := s.getRIB(req.Router, vrfID, req.Pfx.Address.Version)
	if err != nil {
		return nil, wrapGetRIBErr(err, req.Router, vrfID, req.Pfx.Address.Version)
	}

	routes := rib.LPM(bnet.NewPrefixFromProtoPrefix(req.Pfx))
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
	route, err := s.getRoutesFromRouter(req.Router, req)
	if err != nil {
		return nil, fmt.Errorf("unable to get prefix %s for router %s: %w", req.Pfx, req.Router, err)
	}

	return &pb.GetResponse{
		Routes: []*routeapi.Route{
			route.ToProto(),
		},
	}, nil
}

// GetPrefixFromAllRouters gets a prefix from all configured routers (exact match)
func (s *Server) GetPrefixFromAllRouters(ctx context.Context, req *pb.GetPrefixFromAllRoutersRequest) (*pb.GetResponse, error) {
	routesFromAllRouters := make([]*routeapi.Route, 0)
	routers := s.bmp.GetRouters()
	for _, router := range routers {
		route, err := s.getRoutesFromRouter(router.Address().String(), req)
		if err != nil {
			return nil, fmt.Errorf("unable to get prefix from router %s: %w", router.Name(), err)
		}
		if route != nil {
			routesFromAllRouters = append(routesFromAllRouters, route.ToProto())
		}
	}
	return &pb.GetResponse{
		Routes: routesFromAllRouters,
	}, nil
}

func (s *Server) getRoutesFromRouter(router string, req GetRouteRequestI) (*route.Route, error) {
	vrfID, err := getVRFID(req)
	if err != nil {
		return nil, err
	}
	rib, err := s.getRIB(router, vrfID, req.GetPfx().Address.GetVersion())
	if err != nil {
		return nil, wrapGetRIBErr(err, router, vrfID, req.GetPfx().Address.GetVersion())
	}

	route := rib.Get(bnet.NewPrefixFromProtoPrefix(req.GetPfx()))
	return route, nil
}

type GetRouteRequestI interface {
	GetVrfId() uint64
	GetVrf() string
	GetPfx() *netapi.Prefix
}

type GetPrefixFromAllRoutersRequestWrapper struct {
	req *pb.GetPrefixFromAllRoutersRequest
}

func (r *GetPrefixFromAllRoutersRequestWrapper) GetPfx() *netapi.Prefix {
	return r.req.Pfx
}

func (r *GetPrefixFromAllRoutersRequestWrapper) GetVrfId() uint64 {
	return r.req.VrfId
}

func (r *GetPrefixFromAllRoutersRequestWrapper) GetVrf() string {
	return r.req.Vrf
}

type GetRequestWrapper struct {
	req *pb.GetRequest
}

func (r *GetRequestWrapper) GetPfx() *netapi.Prefix {
	return r.req.Pfx
}

func (r *GetRequestWrapper) GetVrf() string {
	return r.req.Vrf
}

func (r *GetRequestWrapper) GetVrfId() uint64 {
	return r.req.VrfId
}

// GetLonger gets all more specifics of a prefix
func (s *Server) GetLonger(ctx context.Context, req *pb.GetLongerRequest) (*pb.GetLongerResponse, error) {
	vrfID, err := getVRFID(req)
	if err != nil {
		return nil, err
	}

	rib, err := s.getRIB(req.Router, vrfID, req.Pfx.Address.Version)
	if err != nil {
		return nil, wrapGetRIBErr(err, req.Router, vrfID, req.Pfx.Address.Version)
	}

	routes := rib.GetLonger(bnet.NewPrefixFromProtoPrefix(req.Pfx))
	res := &pb.GetLongerResponse{
		Routes: make([]*routeapi.Route, 0, len(routes)),
	}
	for _, route := range routes {
		res.Routes = append(res.Routes, route.ToProto())
	}

	return res, nil
}

// ObserveRIB implements the ObserveRIB RPC
func (s *Server) ObserveRIB(req *pb.ObserveRIBRequest, stream pb.RoutingInformationService_ObserveRIBServer) error {
	vrfID, err := getVRFID(req)
	if err != nil {
		return status.New(codes.Unavailable, err.Error()).Err()
	}

	ipVersion := netapi.IP_IPv4
	switch req.Afisafi {
	case pb.ObserveRIBRequest_IPv4Unicast:
		ipVersion = netapi.IP_IPv4
	case pb.ObserveRIBRequest_IPv6Unicast:
		ipVersion = netapi.IP_IPv6
	default:
		return status.New(codes.InvalidArgument, "Unknown AFI/SAFI").Err()
	}

	rib, err := s.getRIB(req.Router, vrfID, ipVersion)
	if err != nil {
		return status.New(codes.Unavailable, wrapGetRIBErr(err, req.Router, vrfID, ipVersion).Error()).Err()
	}

	if !req.AllowUnreadyRib {
		if ready, err := s.bmp.GetRouter(req.Router).Ready(vrfID, ipVersionFromProto(ipVersion)); !ready {
			return status.New(codes.Unavailable, wrapRIBNotReadyErr(err, req.Router, vrfID, ipVersion).Error()).Err()
		}
	}

	risObserveFIBClients.WithLabelValues(req.Router, fmt.Sprintf("%d", req.VrfId), fmt.Sprintf("%d", req.Afisafi)).Inc()
	defer risObserveFIBClients.WithLabelValues(req.Router, fmt.Sprintf("%d", req.VrfId), fmt.Sprintf("%d", req.Afisafi)).Dec()

	fifo := newUpdateFIFO()
	rc := newRIBClient(fifo)
	ret := make(chan error)

	go func(fifo *updateFIFO) {
		var err error

		for {
			for _, toSend := range fifo.dequeue() {
				err = stream.Send(toSend)
				if err != nil {
					ret <- err
					return
				}
			}
		}
	}(fifo)

	rib.RegisterWithOptions(rc, routingtable.ClientOptions{
		MaxPaths: 100,
	})
	defer rib.Unregister(rc)

	select {
	case <-rc.stopped:
		return status.New(codes.Aborted, "ribClient got stopped (probably RIB disappeared)").Err()
	case err = <-ret:
		if err != nil {
			return status.New(codes.Unknown, fmt.Sprintf("Stream ended: %v", err)).Err()
		}
	}

	return nil
}

// DumpRIB implements the DumpRIB RPC
func (s *Server) DumpRIB(req *pb.DumpRIBRequest, stream pb.RoutingInformationService_DumpRIBServer) error {
	vrfID, err := getVRFID(req)
	if err != nil {
		return err
	}

	ipVersion := netapi.IP_IPv4
	switch req.Afisafi {
	case pb.DumpRIBRequest_IPv4Unicast:
		ipVersion = netapi.IP_IPv4
	case pb.DumpRIBRequest_IPv6Unicast:
		ipVersion = netapi.IP_IPv6
	default:
		return fmt.Errorf("uknown AFI/SAFI")
	}

	rib, err := s.getRIB(req.Router, vrfID, ipVersion)
	if err != nil {
		return wrapGetRIBErr(err, req.Router, vrfID, ipVersion)
	}

	toSend := &pb.DumpRIBReply{
		Route: &routeapi.Route{
			Paths: make([]*routeapi.Path, 1),
		},
	}

	routes := rib.Dump()
	for i := range routes {
		if !s.filterRIB(req.GetFilter(), routes[i]) {
			continue
		}
		toSend.Route = routes[i].ToProto()

		err = stream.Send(toSend)
		if err != nil {
			return err
		}
	}

	return nil
}

// filterRIB returns true for routes passing the filter or if the filter is nil
func (s *Server) filterRIB(rf *pb.RIBFilter, route *route.Route) bool {
	if rf == nil {
		return true
	}

	rfOrig := rf.GetOriginatingAsn()
	if rfOrig != 0 && !route.IsBGPOriginatedBy(rfOrig) {
		return false
	}
	if rf.GetMinLength() != 0 && uint32(route.Pfxlen()) < rf.GetMinLength() {
		return false
	}
	if rf.GetMaxLength() != 0 && uint32(route.Pfxlen()) > rf.GetMaxLength() {
		return false
	}
	return true
}

// GetRouters implements the GetRouters RPC
func (s *Server) GetRouters(c context.Context, request *pb.GetRoutersRequest) (*pb.GetRoutersResponse, error) {
	resp := &pb.GetRoutersResponse{}
	routers := s.bmp.GetRouters()
	for _, r := range routers {
		vrfs := r.GetVRFs()
		vrfIDs := make([]uint64, 0, len(vrfs))
		for _, vrf := range vrfs {
			vrfIDs = append(vrfIDs, vrf.RD())
		}
		resp.Routers = append(resp.Routers, &pb.Router{
			SysName: r.Name(),
			VrfIds:  vrfIDs,
			Address: r.Address().String(),
		})
	}
	return resp, nil
}

type RequestWithVRF interface {
	GetVrfId() uint64
	GetVrf() string
}

func getVRFID(req RequestWithVRF) (uint64, error) {
	if req.GetVrf() != "" {
		vrfID, err := vrf.ParseHumanReadableRouteDistinguisher(req.GetVrf())
		if err != nil {
			return 0, fmt.Errorf("unable to parse VRF: %w", err)
		}

		return vrfID, nil
	}

	return req.GetVrfId(), nil
}
