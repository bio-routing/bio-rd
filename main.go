package main

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"

	bnet "github.com/bio-routing/bio-rd/net"
)

func strAddr(s string) uint32 {
	ret, _ := bnet.StrToAddr(s)
	return ret
}

func main() {
	logrus.Printf("This is a BGP speaker\n")

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
		LocalAS:           6695,
		PeerAS:            65300,
		PeerAddress:       net.IP([]byte{169, 254, 200, 1}),
		LocalAddress:      net.IP([]byte{169, 254, 200, 0}),
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
		LocalAS:           6695,
		PeerAS:            65100,
		PeerAddress:       net.IP([]byte{169, 254, 100, 0}),
		LocalAddress:      net.IP([]byte{169, 254, 100, 1}),
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
			fmt.Print(rib.Print())
			time.Sleep(time.Second * 10)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
