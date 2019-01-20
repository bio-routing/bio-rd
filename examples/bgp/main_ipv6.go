// +build ipv6

package main

import (
	"net"
	"time"

	"github.com/bio-routing/bio-rd/config"
	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
	"github.com/sirupsen/logrus"
)

func startServer(b server.BGPServer, v *vrf.VRF) *locRIB.LocRIB {
	rib, err := v.CreateIPv6UnicastLocRIB("inet6.0")
	if err != nil {
		logrus.Fatal(err)
	}

	err = b.Start(&config.Global{
		Listen: true,
		LocalAddressList: []net.IP{
			{0x20, 0x01, 0x6, 0x78, 0x1, 0xe0, 0, 0, 0, 0, 0, 0, 0, 0, 0xca, 0xfe},
		},
	})
	if err != nil {
		logrus.Fatalf("Unable to start BGP server: %v", err)
	}

	b.AddPeer(config.Peer{
		AdminEnabled:      true,
		LocalAS:           65200,
		PeerAS:            202739,
		PeerAddress:       bnet.IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0, 0, 0, 0, 1),
		LocalAddress:      bnet.IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0, 0, 0, 0, 0xcafe),
		ReconnectInterval: time.Second * 15,
		HoldTime:          time.Second * 90,
		KeepAlive:         time.Second * 30,
		Passive:           true,
		RouterID:          b.RouterID(),
		IPv6: &config.AddressFamilyConfig{
			ImportFilter: filter.NewAcceptAllFilter(),
			ExportFilter: filter.NewDrainFilter(),
			AddPathSend: routingtable.ClientOptions{
				BestOnly: true,
			},
		},
		VRF: v,
	})

	b.AddPeer(config.Peer{
		AdminEnabled:      true,
		LocalAS:           65200,
		PeerAS:            65400,
		PeerAddress:       bnet.IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0xcafe, 0, 0, 0, 5),
		LocalAddress:      bnet.IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0, 0, 0, 0, 0xcafe),
		ReconnectInterval: time.Second * 15,
		HoldTime:          time.Second * 90,
		KeepAlive:         time.Second * 30,
		Passive:           true,
		RouterID:          b.RouterID(),
		IPv6: &config.AddressFamilyConfig{
			ImportFilter: filter.NewDrainFilter(),
			ExportFilter: filter.NewAcceptAllFilter(),
			AddPathSend: routingtable.ClientOptions{
				BestOnly: true,
			},
		},
	})

	return rib
}
