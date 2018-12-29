package main

import (
	"fmt"
	"time"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
	"github.com/sirupsen/logrus"
)

func strAddr(s string) uint32 {
	ret, _ := bnet.StrToAddr(s)
	return ret
}

func main() {
	logrus.Printf("This is a BGP speaker\n")

	b := server.NewBgpServer()
	v := vrf.New("master")
	rib := startServer(b, v)

	go func() {
		for {
			fmt.Printf("LocRIB count: %d\n", rib.Count())
			time.Sleep(time.Second * 10)
		}
	}()

	select {}
}
