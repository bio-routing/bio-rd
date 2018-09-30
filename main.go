package main

import (
	"time"
	"github.com/sirupsen/logrus"

	"github.com/bio-routing/bio-rd/config"
	isisserver "github.com/bio-routing/bio-rd/protocols/isis/server"
	"github.com/bio-routing/bio-rd/protocols/isis/types"

	bnet "github.com/bio-routing/bio-rd/net"
)

func strAddr(s string) uint32 {
	ret, _ := bnet.StrToAddr(s)
	return ret
}

func main() {
	logrus.Printf("This is an ISIS speaker\n")

	isis := isisserver.NewISISServer(config.ISISConfig{
		NETs: []config.NET{
			{
				AFI:      0x49,
				AreaID:   []byte{0, 1, 0, 10},
				SystemID: types.SystemID{0, 0, 0, 0, 0, 4},
			},
		},
		Interfaces: []config.ISISInterfaceConfig{
			{
				Name:    "virbr3",
				P2P:     true,
				Passive: false,
				ISISLevel2Config: &config.ISISLevelConfig{
					HelloInterval: 1,
					HoldTime:      4,
					Metric:        10,
					Priority:      64,
				},
			},
		},
		MinLSPTransmissionInterval: time.Second * 5,
	})

	err := isis.Start()
	if err != nil {
		logrus.Errorf("Unable to start ISIS: %v", err)
	}

	/*logrus.Printf("This is a BGP speaker\n")

	rib := locRIB.New()
	b := server.NewBgpServer()

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
		AddPathSend: routingtable.ClientOptions{
			MaxPaths: 10,
		},
		ImportFilter:      filter.NewAcceptAllFilter(),
		ExportFilter:      filter.NewAcceptAllFilter(),
		RouteServerClient: true,
	}, rib)

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
		AddPathSend: routingtable.ClientOptions{
			MaxPaths: 10,
		},
		AddPathRecv:       true,
		ImportFilter:      filter.NewAcceptAllFilter(),
		ExportFilter:      filter.NewAcceptAllFilter(),
		RouteServerClient: true,
	}, rib)

	go func() {
		for {
			fmt.Printf("LocRIB count: %d\n", rib.Count())
			time.Sleep(time.Second * 10)
		}
	}()*/

	select {}
}
