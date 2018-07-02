// +build ipv6

package main

import (
	"net"
	"time"

	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	"github.com/sirupsen/logrus"

	bnet "github.com/bio-routing/bio-rd/net"
)

func startServer(b server.BGPServer, rib *locRIB.LocRIB) {
	err := b.Start(&config.Global{
		Listen: true,
		LocalAddressList: []net.IP{
			net.IP{0x20, 0x01, 0x6, 0x78, 0x1, 0xe0, 0, 0, 0, 0, 0, 0, 0, 0, 0xca, 0xfe},
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
		AddPathSend: routingtable.ClientOptions{
			BestOnly: true,
		},
		ImportFilter: filter.NewAcceptAllFilter(),
		ExportFilter: filter.NewDrainFilter(),
		IPv6:         true,
	}, rib)
}
