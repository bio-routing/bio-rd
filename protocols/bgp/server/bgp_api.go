package server

import (
	"context"
	"fmt"

	"github.com/bio-routing/bio-rd/protocols/bgp/api"
	"google.golang.org/grpc"
)

type BgpApiServer struct {
}

func (s *BgpApiServer) ListSessions(ctx context.Context, in *api.ListSessionsRequest, opts ...grpc.CallOption) (*api.ListSessionsResponse, error) {
	return nil, fmt.Errorf("Not implemented yet.")
}
