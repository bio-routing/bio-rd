package main

import (
	"fmt"
	"net"
	"time"

	"github.com/bio-routing/bio-rd/protocols/bmp/server"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.Printf("This is a BMP speaker\n")

	rib := locRIB.New()
	b := server.NewServer()
	b.AddRouter(net.IP{127, 0, 0, 1}, 1234, rib, nil)

	go func() {
		for {
			fmt.Printf("LocRIB count: %d\n", rib.Count())
			time.Sleep(time.Second * 10)
		}
	}()

	select {}
}
