package main

import (
	"time"

	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"

	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	log "github.com/sirupsen/logrus"

	bnet "github.com/bio-routing/bio-rd/net"
)

func startBGPServer(b server.BGPServer, rib *locRIB.LocRIB, cfg *config.Global) {
	err := b.Start(cfg)
	if err != nil {
		log.Fatalf("Unable to start BGP server: %v", err)
	}

	b.AddPeer(config.Peer{
		AdminEnabled:      true,
		LocalAS:           65200,
		PeerAS:            65100,
		PeerAddress:       bnet.IPv4FromOctets(169, 254, 0, 1),
		LocalAddress:      bnet.IPv4FromOctets(169, 254, 0, 2),
		ReconnectInterval: time.Second * 20,
		HoldTime:          time.Second * 20,
		KeepAlive:         time.Second * 20,
		Passive:           false,
		RouterID:          b.RouterID(),

		//AddPathSend: routingtable.ClientOptions{
		//	MaxPaths: 10,
		//},
		//RouteServerClient: true,
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
