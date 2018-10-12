package main

import (
	"fmt"
	"net"
	"time"

	"github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.Printf("This is a BMP speaker\n")

	rib := locRIB.New()
	b := server.NewServer()
	b.AddRouter(net.IP{10, 0, 255, 0}, 30119, rib, nil)

	go func() {
		for {
			fmt.Printf("LocRIB count: %d\n", rib.Count())
			time.Sleep(time.Second * 10)
		}
	}()

	select {}
}
