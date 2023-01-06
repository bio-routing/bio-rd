package server

import (
	"context"
	"fmt"

	"github.com/bio-routing/bio-rd/protocols/bgp/api"
	"github.com/bio-routing/bio-rd/routingtable/vrf"

	bnet "github.com/bio-routing/bio-rd/net"
)

type BGPAPIServer struct {
	api.UnimplementedBgpServiceServer
	srv    BGPServer
	vrfReg *vrf.VRFRegistry
}

// NewBGPAPIServer creates a new BGP API Server
func NewBGPAPIServer(s BGPServer, vrfReg *vrf.VRFRegistry) *BGPAPIServer {
	return &BGPAPIServer{
		srv:    s,
		vrfReg: vrfReg,
	}
}

func (s *BGPAPIServer) ListSessions(ctx context.Context, in *api.ListSessionsRequest) (*api.ListSessionsResponse, error) {
	return nil, fmt.Errorf("not implemented yet")
}

// DumpRIBIn dumps the RIB in of a peer for a given AFI/SAFI
func (s *BGPAPIServer) DumpRIBIn(in *api.DumpRIBRequest, stream api.BgpService_DumpRIBInServer) error {
	v := s.getVRF(in)
	if v == nil {
		return fmt.Errorf("unable to find vrf %q", in.VrfName)
	}

	r := s.srv.GetRIBIn(v, bnet.IPFromProtoIP(in.Peer).Ptr(), uint16(in.Afi), uint8(in.Safi))
	if r == nil {
		return fmt.Errorf("unable to get AdjRIBIn")
	}

	for _, r := range r.Dump() {
		x := r.ToProto()
		err := stream.Send(x)
		if err != nil {
			return err
		}
	}

	return nil
}

// DumpRIBOut dumps the RIB out of a peer for a given AFI/SAFI
func (s *BGPAPIServer) DumpRIBOut(in *api.DumpRIBRequest, stream api.BgpService_DumpRIBOutServer) error {
	v := s.getVRF(in)
	if v == nil {
		return fmt.Errorf("unable to find vrf %q", in.VrfName)
	}

	r := s.srv.GetRIBOut(v, bnet.IPFromProtoIP(in.Peer).Ptr(), uint16(in.Afi), uint8(in.Safi))
	if r == nil {
		return fmt.Errorf("unable to get AdjRIBOut")
	}

	for _, r := range r.Dump() {
		x := r.ToProto()
		err := stream.Send(x)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *BGPAPIServer) getVRF(in *api.DumpRIBRequest) *vrf.VRF {
	if in.VrfName == "" {
		in.VrfName = vrf.DefaultVRFName
	}

	return s.vrfReg.GetVRFByName(in.VrfName)
}
