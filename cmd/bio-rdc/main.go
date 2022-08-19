package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	bnet "github.com/bio-routing/bio-rd/net"
	bgpapi "github.com/bio-routing/bio-rd/protocols/bgp/api"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/util/log"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var (
	bioAddr      = flag.String("bio-rd", "localhost:5566", "bio-rd grpc endpoint")
	cmd          = flag.String("cmd", "", "command to execute")
	bgpAPIClient bgpapi.BgpServiceClient
)

func main() {
	flag.Parse()

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	log.SetLogger(log.NewLogrusWrapper(logger))

	conn, err := grpc.Dial(*bioAddr, grpc.WithInsecure())
	if err != nil {
		log.Errorf("GRPC dial failed: %v", err)
		os.Exit(1)
	}
	defer conn.Close()

	bgpAPIClient = bgpapi.NewBgpServiceClient(conn)

	cmdParts := strings.Split(*cmd, " ")
	if len(cmdParts) == 0 {
		return
	}

	if cmdParts[0] == "show" {
		if len(cmdParts) == 1 {
			return
		}
		show(cmdParts[1:])
	}

}

func show(parts []string) {
	if parts[0] == "routes" {
		if len(parts) == 1 {
			return
		}

		showRoute(parts[1:])
	}
}

func showRoute(parts []string) {
	if parts[0] == "receive-protocol" {
		if len(parts) == 1 {
			return
		}

		if parts[1] == "bgp" {
			if len(parts) == 2 {
				return
			}

			showRouteReceiveBGP(parts[2:])
		}
	}
}

func showRouteReceiveBGP(parts []string) {
	if len(parts) == 0 {
		return
	}

	peer, err := bnet.IPFromString(parts[0])
	if err != nil {
		log.Errorf("unable to convert peer address: %v", err)
		return
	}

	c, err := bgpAPIClient.DumpRIBIn(context.Background(), &bgpapi.DumpRIBRequest{
		Peer: peer.ToProto(),
		Afi:  1,
		Safi: 1,
	})

	if err != nil {
		log.Errorf("Failed to get streaming RPC client: %v", err)
		return
	}

	for {
		r, err := c.Recv()
		if err == io.EOF {
			return
		}

		if err != nil {
			log.Errorf("Recv() failed: %v", err)
			return
		}

		rr := route.RouteFromProtoRoute(r, false)
		fmt.Println(rr.Print())
	}

}
