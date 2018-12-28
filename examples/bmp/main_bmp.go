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

	rib4 := locRIB.New()
	rib6 := locRIB.New()
	b := server.NewServer()
	b.AddRouter(net.IP{10, 0, 255, 0}, 30119, rib4, rib6)

	go func() {
		for {
			fmt.Printf("LocRIB4 count: %d\n", rib4.Count())
			time.Sleep(time.Second * 10)
		}
	}()

	select {}
}
