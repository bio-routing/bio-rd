package main

import (
	"context"
	"fmt"
	"os"

	pb "github.com/bio-routing/bio-rd/cmd/ris/api"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
)

// NewDumpLocRIBCommand creates a new dump local rib command
func NewDumpLocRIBCommand() cli.Command {
	cmd := cli.Command{
		Name:  "dump-loc-rib",
		Usage: "dump loc RIB",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "4", Usage: "print IPv4 routes"},
			&cli.BoolFlag{Name: "6", Usage: "print IPv6 routes"},
		},
	}

	cmd.Action = func(c *cli.Context) error {
		conn, err := grpc.Dial(c.GlobalString("ris"), grpc.WithInsecure())
		if err != nil {
			log.Errorf("GRPC dial failed: %v", err)
			os.Exit(1)
		}
		defer conn.Close()

		afisafis := make([]pb.DumpRIBRequest_AFISAFI, 0)
		req_ipv4, req_ipv6 := c.Bool("4"), c.Bool("6")
		if !req_ipv4 && !req_ipv6 {
			req_ipv4, req_ipv6 = true, true
		}
		if req_ipv4 {
			afisafis = append(afisafis, pb.DumpRIBRequest_IPv4Unicast)
		}
		if req_ipv6 {
			afisafis = append(afisafis, pb.DumpRIBRequest_IPv6Unicast)
		}

		client := pb.NewRoutingInformationServiceClient(conn)
		for _, afisafi := range afisafis {
			fmt.Printf(" --- Dump %s ---\n", pb.DumpRIBRequest_AFISAFI_name[int32(afisafi)])
			err = dumpRIB(client, c.GlobalString("router"), c.GlobalUint64("vrf_id"), afisafi)
			if err != nil {
				log.Errorf("DumpRIB failed: %v", err)
			}
		}

		return nil
	}

	return cmd
}

func dumpRIB(c pb.RoutingInformationServiceClient, routerName string, vrfID uint64, afisafi pb.DumpRIBRequest_AFISAFI) error {
	client, err := c.DumpRIB(context.Background(), &pb.DumpRIBRequest{
		Router:  routerName,
		VrfId:   vrfID,
		Afisafi: afisafi,
	})
	if err != nil {
		return errors.Wrap(err, "Unable to get client")
	}

	for {
		r, err := client.Recv()
		if err != nil {
			return errors.Wrap(err, "Received failed")
		}

		printRoute(r.Route)
	}
}
