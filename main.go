package main

import (
	"fmt"
	"net"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/taktv6/tbgp/config"
	"github.com/taktv6/tbgp/server"
)

func main() {
	fmt.Printf("This is a BGP speaker\n")

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
		PeerAS:       65201,
		PeerAddress:  net.IP([]byte{169, 254, 123, 1}),
		LocalAddress: net.IP([]byte{169, 254, 123, 0}),
		HoldTimer:    90,
		KeepAlive:    30,
		Passive:      true,
		RouterID:     b.RouterID(),
	})

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
