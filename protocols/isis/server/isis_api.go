package server

import (
	"context"
	"time"

	netapi "github.com/bio-routing/bio-rd/net/api"
	"github.com/bio-routing/bio-rd/protocols/isis/api"
)

type ISISAPIServer struct {
	api.UnimplementedIsisServiceServer
	srv ISISServer
}

// NewISISAPIServer creates a new ISIS API Server
func NewISISAPIServer(s ISISServer) *ISISAPIServer {
	return &ISISAPIServer{
		srv: s,
	}
}

func (s *ISISAPIServer) ListAdjacencies(context.Context, *api.ListAdjacenciesRequest) (*api.ListAdjacenciesResponse, error) {
	res := &api.ListAdjacenciesResponse{
		Adjacencies: make([]*api.Adjacency, 0),
	}

	for _, a := range s.srv.GetAdjacencies() {
		addrs := make([]*netapi.IP, 0, len(a.IPAddresses))
		for _, addr := range a.IPAddresses {
			addrs = append(addrs, addr.ToProto())
		}

		adj := &api.Adjacency{
			Name:               a.Name,
			SystemId:           a.SystemID[:],
			Address:            a.Address[:],
			InterfaceName:      a.InterfaceName,
			Level:              uint32(a.Level),
			Priority:           uint32(a.Priority),
			IpAddresses:        addrs,
			LastTransitionUnix: a.LastStateChange.Unix(),
			ExpiresInSeconds:   uint32(a.Timeout.Sub(time.Now()).Seconds()),
			Status:             api.Adjacency_State(a.Status),
		}

		res.Adjacencies = append(res.Adjacencies, adj)
	}

	return res, nil
}
