package main

import (
	"fmt"
	"net"
	"time"

	"github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/sirupsen/logrus"

	bnet "github.com/bio-routing/bio-rd/net"
)

func main() {
	logrus.Printf("This is a BMP speaker\n")

	b := server.NewServer()
	b.AddRouter(net.IP{10, 0, 255, 1}, 30119)

	go func() {
		for {
			for _, r := range b.GetRouters() {
				for _, v := range r.GetVRFs() {

					rib4 := v.IPv4UnicastRIB()
					c := rib4.Count()
					fmt.Printf("Router: %s VRF: %s IPv4 route count: %d\n", r.Name(), v.Name(), c)

					if v.RD() == 220434901565105 {
						for _, route := range rib4.Dump() {
							fmt.Printf("Pfx: %s\n", route.Prefix().String())
							for _, p := range route.Paths() {
								fmt.Printf("   %s\n", p.String())
							}
						}

						fmt.Printf("looking up 185.65.240.100\n")
						for _, r := range rib4.LPM(bnet.NewPfx(bnet.IPv4FromOctets(185, 65, 240, 100), 32)) {
							fmt.Printf("Pfx: %s\n", r.Prefix().String())
							for _, p := range r.Paths() {
								fmt.Printf("   %s\n", p.String())
							}
						}

						fmt.Printf("is 8.8.8.8 in closednet?\n")
						x := rib4.LPM(bnet.NewPfx(bnet.IPv4FromOctets(8, 8, 8, 8), 32))
						if len(x) == 0 {
							fmt.Printf("Nope\n")
						} else {
							fmt.Printf("Yep\n")
						}

						fmt.Printf("is 185.65.240.100 in closednet?\n")
						x = rib4.LPM(bnet.NewPfx(bnet.IPv4FromOctets(185, 65, 240, 0), 32))
						if len(x) == 0 {
							fmt.Printf("Nope\n")
						} else {
							fmt.Printf("Yep\n")
						}
					}
				}
			}

			time.Sleep(time.Second * 10)
		}
	}()

	select {}
}
