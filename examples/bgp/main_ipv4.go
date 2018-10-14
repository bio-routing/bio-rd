// +build !ipv6

package main

import (
	"net"
	"time"

	"github.com/bio-routing/bio-rd/routingtable/locRIB"

	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/sirupsen/logrus"

	bnet "github.com/bio-routing/bio-rd/net"
)

func startServer(b server.BGPServer, rib *locRIB.LocRIB) {
	err := b.Start(&config.Global{
		Listen: true,
		LocalAddressList: []net.IP{
			net.IPv4(169, 254, 100, 1),
			net.IPv4(169, 254, 200, 0),
		},
	})
	if err != nil {
		logrus.Fatalf("Unable to start BGP server: %v", err)
	}

	b.AddPeer(config.Peer{
		AdminEnabled:      true,
		LocalAS:           65200,
		PeerAS:            65300,
		PeerAddress:       bnet.IPv4FromOctets(172, 17, 0, 3),
		LocalAddress:      bnet.IPv4FromOctets(169, 254, 200, 0),
		ReconnectInterval: time.Second * 15,
		HoldTime:          time.Second * 90,
		KeepAlive:         time.Second * 30,
		Passive:           true,
		RouterID:          b.RouterID(),
		IPv4: &config.AddressFamilyConfig{
			RIB:          rib,
			ImportFilter: filter.NewAcceptAllFilter(),
			ExportFilter: filter.NewAcceptAllFilter(),
			AddPathSend: routingtable.ClientOptions{
				MaxPaths: 10,
			},
			AddPathRecv: true,
		},
		RouteServerClient: true,
	})

	b.AddPeer(config.Peer{
		AdminEnabled:      true,
		LocalAS:           65200,
		PeerAS:            65100,
		PeerAddress:       bnet.IPv4FromOctets(172, 17, 0, 2),
		LocalAddress:      bnet.IPv4FromOctets(169, 254, 100, 1),
		ReconnectInterval: time.Second * 15,
		HoldTime:          time.Second * 90,
		KeepAlive:         time.Second * 30,
		Passive:           true,
		RouterID:          b.RouterID(),
		RouteServerClient: true,
		IPv4: &config.AddressFamilyConfig{
			RIB:          rib,
			ImportFilter: filter.NewAcceptAllFilter(),
			ExportFilter: filter.NewAcceptAllFilter(),
			AddPathSend: routingtable.ClientOptions{
				MaxPaths: 10,
			},
			AddPathRecv: true,
		},
	})
}
