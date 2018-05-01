package main

import (
	"fmt"
	"net"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/rt"
)

func main() {
	fmt.Printf("This is a BGP speaker\n")

	VRF := rt.New(true)
	b := server.NewBgpServer()

	err := b.Start(&config.Global{
		Listen: true,
	})
	if err != nil {
		logrus.Fatalf("Unable to start BGP server: %v", err)
	}

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
	}, VRF)

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
	}, VRF)

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
