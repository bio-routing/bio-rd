package main

import (
	"context"
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
	}

	cmd.Action = func(c *cli.Context) error {
		conn, err := grpc.Dial(c.GlobalString("ris"), grpc.WithInsecure())
		if err != nil {
			log.Errorf("GRPC dial failed: %v", err)
			os.Exit(1)
		}
		defer conn.Close()

		client := pb.NewRoutingInformationServiceClient(conn)
		err = dumpRIB(client, c.GlobalString("router"), c.GlobalUint64("vrf_id"))
		if err != nil {
			log.Errorf("DumpRIB failed: %v", err)
			os.Exit(1)
		}

		return nil
	}

	return cmd
}

func dumpRIB(c pb.RoutingInformationServiceClient, routerName string, vrfID uint64) error {
	client, err := c.DumpRIB(context.Background(), &pb.DumpRIBRequest{
		Router:  routerName,
		VrfId:   vrfID,
		Afisafi: pb.DumpRIBRequest_IPv4Unicast,
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
