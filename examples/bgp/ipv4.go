package main

import (
	"time"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
)

func addPeersIPv4(b server.BGPServer, v *vrf.VRF) {
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
		IPv4: &server.AddressFamilyConfig{
			ImportFilterChain: filter.NewAcceptAllFilterChain(),
			ExportFilterChain: filter.NewAcceptAllFilterChain(),
			AddPathSend: routingtable.ClientOptions{
				MaxPaths: 10,
			},
			AddPathRecv: true,
		},
		RouteServerClient: true,
		VRF:               v,
	})
}
