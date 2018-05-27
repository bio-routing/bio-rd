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
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
)

func main() {
	fmt.Printf("This is a BGP speaker\n")

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
		AdminEnabled: true,
		LocalAS:      65200,
		PeerAS:       65300,
		PeerAddress:  net.IP([]byte{169, 254, 200, 1}),
		LocalAddress: net.IP([]byte{169, 254, 200, 0}),
		HoldTimer:    90,
		KeepAlive:    30,
		Passive:      true,
		RouterID:     b.RouterID(),
		AddPathSend: routingtable.ClientOptions{
			MaxPaths: 10,
		},
	}, rib)

	time.Sleep(time.Second * 15)

	b.AddPeer(config.Peer{
		AdminEnabled: true,
		LocalAS:      65200,
		PeerAS:       65100,
		PeerAddress:  net.IP([]byte{169, 254, 100, 0}),
		LocalAddress: net.IP([]byte{169, 254, 100, 1}),
		HoldTimer:    90,
		KeepAlive:    30,
		Passive:      true,
		RouterID:     b.RouterID(),
		AddPathSend: routingtable.ClientOptions{
			MaxPaths: 10,
		},
		AddPathRecv: true,
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
