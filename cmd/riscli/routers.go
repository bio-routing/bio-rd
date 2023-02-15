package main

import (
	"context"
	"fmt"
	"os"

	pb "github.com/bio-routing/bio-rd/cmd/ris/api"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
	"github.com/bio-routing/bio-rd/util/log"
	"github.com/urfave/cli"
)

// NewLPMCommand creates a new LPM command
func NewGetRoutersCommand() cli.Command {
	cmd := cli.Command{
		Name:  "routers",
		Usage: "get all routers and vrfs available",
		Flags: []cli.Flag{},
	}

	cmd.Action = func(c *cli.Context) error {
		conn := SetupGRPCClient(c)
		defer conn.Close()

		client := pb.NewRoutingInformationServiceClient(conn)
		err := getRouters(client)
		if err != nil {
			log.Errorf("Get Routers failed: %v", err)
			os.Exit(1)
		}

		return nil
	}

	return cmd
}

func getRouters(c pb.RoutingInformationServiceClient) error {
	resp, err := c.GetRouters(context.Background(), &pb.GetRoutersRequest{})
	if err != nil {
		return fmt.Errorf("unable to get client: %w", err)
	}
	for _, r := range resp.GetRouters() {
		fmt.Printf("Router %s at %s\n", r.SysName, r.Address)
		fmt.Println("VRFs:")
		for _, v := range r.VrfIds {
			fmt.Printf("%s (Numeric: %v)\n", vrf.RouteDistinguisherHumanReadable(v), v)
		}
		fmt.Println("---")
	}
	return nil
}
