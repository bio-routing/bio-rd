package main

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	mpb "github.com/bio-routing/bio-rd/proto/mgmt"
	"github.com/bio-routing/bio-rd/protocols/bgp/server"
)

type mgmtService struct {
	// managed protocol servers by name
	bgp map[string]server.BGPServer
	// grpc listen address
	listen string
}

func newMgmtService(listen string, bgp map[string]server.BGPServer) *mgmtService {
	return &mgmtService{
		listen: listen,
		bgp:    bgp,
	}
}

func (s *mgmtService) GetBGPServerStatus(ctx context.Context, req *mpb.GetBGPServerStatusRequest) (*mpb.GetBGPServerStatusResponse, error) {
	server, ok := s.bgp[req.Name]
	if !ok {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("BGP server %q not found", req.Name))
	}

	res := &mpb.GetBGPServerStatusResponse{
		RouterId: server.RouterID(),
		LocalAsn: server.LocalASN(),
	}

	for _, peer := range server.GetPeerInfoAll() {
		res.Peers = append(res.Peers, &mpb.GetBGPServerStatusResponse_PeerInfo{
			PeerAddress: peer.PeerAddr.String(),
			PeerAsn:     peer.PeerASN,
			LocalAsn:    peer.LocalASN,
		})
	}

	return res, nil
}

func (s *mgmtService) ListServers(ctx context.Context, req *mpb.ListServersRequest) (*mpb.ListServersResponse, error) {
	res := &mpb.ListServersResponse{}
	for k, _ := range s.bgp {
		res.Server = append(res.Server, &mpb.ListServersResponse_Server{
			Family: "bgp",
			Name:   k,
		})
	}
	return res, nil
}

func (s *mgmtService) start() {
	lis, err := net.Listen("tcp", s.listen)
	if err != nil {
		panic(fmt.Sprintf("Could not listen to management service: %v", err))
	}

	g := grpc.NewServer()
	mpb.RegisterManagementServer(g, s)
	reflection.Register(g)

	go func() {
		err := g.Serve(lis)
		if err != nil {
			panic(fmt.Sprintf("Could not start management service: %v", err))
		}
	}()
}
