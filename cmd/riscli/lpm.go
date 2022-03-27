package main

import (
	"context"
	"fmt"
	"os"

	pb "github.com/bio-routing/bio-rd/cmd/ris/api"
	bnet "github.com/bio-routing/bio-rd/net"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
)

// NewLPMCommand creates a new LPM command
func NewLPMCommand() cli.Command {
	cmd := cli.Command{
		Name:  "lpm",
		Usage: "longest prefix match",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "ip", Usage: "IP address"},
		},
	}

	cmd.Action = func(c *cli.Context) error {
		conn, err := grpc.Dial(c.GlobalString("ris"), grpc.WithInsecure())
		if err != nil {
			log.Errorf("GRPC dial failed: %v", err)
			os.Exit(1)
		}
		defer conn.Close()

		ipAddr, err := bnet.IPFromString(c.String("ip"))
		if err != nil {
			log.Fatalf("Unable to parse address: %v", err)
		}

		pfxLen := uint8(32)
		if !ipAddr.IsIPv4() {
			pfxLen = 128
		}
		pfx := bnet.NewPfx(ipAddr, pfxLen)

		client := pb.NewRoutingInformationServiceClient(conn)
		err = lpm(client, c.GlobalString("router"), c.GlobalUint64("vrf_id"), c.GlobalString("vrf"), pfx)
		if err != nil {
			log.Fatalf("LPM failed: %v", err)
		}

		return nil
	}

	return cmd
}

func lpm(c pb.RoutingInformationServiceClient, routerName string, vrfID uint64, vrf string, pfx bnet.Prefix) error {
	resp, err := c.LPM(context.Background(), &pb.LPMRequest{
		Router: routerName,
		VrfId:  vrfID,
		Vrf:    vrf,
		Pfx:    pfx.ToProto(),
	})
	if err != nil {
		return fmt.Errorf("unable to get client: %w", err)
	}

	for _, r := range resp.Routes {
		printRoute(r)
	}

	return nil
}
