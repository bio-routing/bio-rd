package server

import (
	"context"
	"fmt"

	"github.com/bio-routing/bio-rd/protocols/bgp/api"
	"github.com/bio-routing/bio-rd/protocols/bgp/metrics"
	"github.com/bio-routing/bio-rd/route"
	"github.com/pkg/errors"

	bnet "github.com/bio-routing/bio-rd/net"
	routeapi "github.com/bio-routing/bio-rd/route/api"
)

type BGPAPIServer struct {
	srv BGPServer
}

// NewBGPAPIServer creates a new BGP API Server
func NewBGPAPIServer(s BGPServer) *BGPAPIServer {
	return &BGPAPIServer{
		srv: s,
	}
}

// ListSessions lists all sessions the BGP server currently has
func (s *BGPAPIServer) ListSessions(ctx context.Context, in *api.ListSessionsRequest) (*api.ListSessionsResponse, error) {
	bgpMetrics, err := s.srv.Metrics()
	if err != nil {
		return nil, errors.Wrap(err, "Could not get peer metrics")
	}

	sessions := make([]*api.Session, 0)
	for _, peerIP := range s.srv.GetPeers() {
		peer := s.srv.GetPeerConfig(peerIP)
		// find metrics for peer
		var peerMetrics *metrics.BGPPeerMetrics
		for _, peerMetricsEntry := range bgpMetrics.Peers {
			if *peerMetricsEntry.IP == *peer.PeerAddress {
				peerMetrics = peerMetricsEntry
				break
			}
		}
		if peerMetrics == nil {
			return nil, fmt.Errorf("Could not find metrics for neighbor %s", peer.PeerAddress)
		}

		if in.Filter != nil {
			if in.Filter.NeighborIp != nil {
				filterNeighbor := bnet.IPFromProtoIP(in.Filter.NeighborIp)
				if *filterNeighbor != *peerIP {
					continue
				}
			}
			if in.Filter.VrfName != "" {
				if in.Filter.VrfName != peerMetrics.VRF {
					continue
				}
			}
		}

		estSince := peerMetrics.Since.Unix()
		if estSince < 0 {
			// time not set, peer probably not up
			estSince = 0
		}
		var routesReceived, routesSent uint64
		for _, afiPeerMetrics := range peerMetrics.AddressFamilies {
			routesReceived += afiPeerMetrics.RoutesReceived
			routesSent += afiPeerMetrics.RoutesSent
		}

		session := &api.Session{
			LocalAddress:    peer.LocalAddress.ToProto(),
			NeighborAddress: peer.PeerAddress.ToProto(),
			LocalAsn:        peer.LocalAS,
			PeerAsn:         peer.PeerAS,
			Status:          peerMetrics.GetStateAsProto(),
			Stats: &api.SessionStats{
				MessagesIn:     peerMetrics.UpdatesReceived,
				MessagesOut:    peerMetrics.UpdatesSent,
				RoutesReceived: routesReceived,
				RoutesExported: routesSent,
			},
			EstablishedSince: uint64(estSince),
			Description:      peer.Description,
			VrfName:          peerMetrics.VRF,
		}
		sessions = append(sessions, session)
	}

	resp := &api.ListSessionsResponse{
		Sessions: sessions,
	}
	return resp, nil
}

// DumpRIBIn dumps the RIB in of a peer for a given AFI/SAFI
func (s *BGPAPIServer) DumpRIBIn(in *api.DumpRIBRequest, stream api.BgpService_DumpRIBInServer) error {
	r := s.srv.GetRIBIn(bnet.IPFromProtoIP(in.Peer), uint16(in.Afi), uint8(in.Safi))
	if r == nil {
		return fmt.Errorf("Unable to get AdjRIBIn")
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
	r := s.srv.GetRIBOut(bnet.IPFromProtoIP(in.Peer), uint16(in.Afi), uint8(in.Safi))
	if r == nil {
		return fmt.Errorf("Unable to get AdjRIBOut")
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

func routesToProto(dump []*route.Route) []*routeapi.Route {
	routes := make([]*routeapi.Route, len(dump))
	for i := range dump {
		routes[i] = dump[i].ToProto()
	}

	return routes
}
