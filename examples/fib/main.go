package main

import (
	"log"
	"net"
	"time"

	"github.com/bio-routing/bio-rd/config"
	bnet "github.com/bio-routing/bio-rd/net"
	api "github.com/bio-routing/bio-rd/protocols/bgp/api"
	"github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/protocols/fib"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func startBGPServer(b server.BGPServer, v *vrf.VRF) {
	apiSrv := server.NewBGPAPIServer(b)

	lis, err := net.Listen("tcp", ":1337")
	if err != nil {
		logrus.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	api.RegisterBgpServiceServer(grpcServer, apiSrv)
	go grpcServer.Serve(lis)

	err = b.Start(&config.Global{
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
		PeerAS:            65100,
		PeerAddress:       bnet.IPv4FromOctets(172, 17, 0, 2),
		LocalAddress:      bnet.IPv4FromOctets(169, 254, 100, 1),
		ReconnectInterval: time.Second * 15,
		HoldTime:          time.Second * 90,
		KeepAlive:         time.Second * 30,
		Passive:           true,
		RouterID:          b.RouterID(),
		RouteServerClient: true,
		IPv4: &config.AddressFamilyConfig{
			ImportFilter: filter.NewAcceptAllFilter(),
			ExportFilter: filter.NewAcceptAllFilter(),
			AddPathSend: routingtable.ClientOptions{
				MaxPaths: 10,
			},
			AddPathRecv: true,
		},
		VRF: v,
	})
}

func startFIB(f *fib.FIB) {
	err := f.Start()
	if err != nil {
		logrus.Fatalf("Unable to start FIB: %v", err)
	}
}

func main() {
	logrus.Printf("This is a linux router that speaks BGP\n")

	v, err := vrf.NewDefaultVRF()
	if err != nil {
		logrus.Fatal(err)
	}

	//b := server.NewBgpServer()
	//startBGPServer(b, v)

	go func() {
		rib, found := v.RIBByName("inet.0")
		if found {
			for {
				log.Print("\n\n### LocRIB DUmP:")
				log.Print(rib.Print())
				time.Sleep(5 * time.Second)
			}
		}
	}()

	f := fib.New(v)
	startFIB(f)

	select {}
}
