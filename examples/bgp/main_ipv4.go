//go:build !ipv6
// +build !ipv6

package main

import (
	"net"
	"time"

	"github.com/bio-routing/bio-rd/routingtable/vrf"
	"google.golang.org/grpc"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/sirupsen/logrus"

	api "github.com/bio-routing/bio-rd/protocols/bgp/api"
)

func startServer(b server.BGPServer, v *vrf.VRF) {
	apiSrv := server.NewBGPAPIServer(b)

	lis, err := net.Listen("tcp", ":1337")
	if err != nil {
		logrus.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	api.RegisterBgpServiceServer(grpcServer, apiSrv)
	go grpcServer.Serve(lis)

	err = b.Start()
	if err != nil {
		logrus.Fatalf("Unable to start BGP server: %v", err)
	}

	b.AddPeer(server.PeerConfig{
		AdminEnabled:      true,
		LocalAS:           65200,
		PeerAS:            65300,
		PeerAddress:       bnet.IPv4FromOctets(172, 17, 0, 3).Ptr(),
		LocalAddress:      bnet.IPv4FromOctets(169, 254, 200, 0).Ptr(),
		ReconnectInterval: time.Second * 15,
		HoldTime:          time.Second * 90,
		KeepAlive:         time.Second * 30,
		Passive:           true,
		RouterID:          b.RouterID(),
		IPv4: &server.AddressFamilyConfig{
			ImportFilterChain: filter.NewAcceptAllFilterChain(),
			ExportFilterChain: filter.NewAcceptAllFilterChain(),
			AddPathSend: routingtable.ClientOptions{
				MaxPaths: 10,
			},
		},
		RouteServerClient: true,
		VRF:               v,
	})

	b.AddPeer(server.PeerConfig{
		AdminEnabled:      true,
		LocalAS:           65200,
		PeerAS:            65100,
		PeerAddress:       bnet.IPv4FromOctets(172, 17, 0, 2).Ptr(),
		LocalAddress:      bnet.IPv4FromOctets(169, 254, 100, 1).Ptr(),
		ReconnectInterval: time.Second * 15,
		HoldTime:          time.Second * 90,
		KeepAlive:         time.Second * 30,
		Passive:           true,
		RouterID:          b.RouterID(),
		RouteServerClient: true,
		IPv4: &server.AddressFamilyConfig{
			ImportFilterChain: filter.NewAcceptAllFilterChain(),
			ExportFilterChain: filter.NewAcceptAllFilterChain(),
			AddPathSend: routingtable.ClientOptions{
				MaxPaths: 10,
			},
			AddPathRecv: true,
		},
		VRF: v,
	})
}
