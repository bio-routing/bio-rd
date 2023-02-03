package main

import (
	"context"
	"fmt"
	"io"
	"os"

	pb "github.com/bio-routing/bio-rd/cmd/ris/api"
	"github.com/bio-routing/bio-rd/util/log"
	"github.com/urfave/cli"
)

// NewObserveRIBCommand creates a new observe rib command
func NewObserveRIBCommand() cli.Command {
	cmd := cli.Command{
		Name:  "observe-rib",
		Usage: "observes the RIB",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "4", Usage: "print IPv4 routes"},
			&cli.BoolFlag{Name: "6", Usage: "print IPv6 routes"},
			&cli.BoolFlag{Name: "allow-unready", Usage: "Do not wait for RIB readiness"},
		},
	}

	cmd.Action = func(c *cli.Context) error {
		conn := SetupGRPCClient(c)
		defer conn.Close()
		var err error

		afisafis := make([]pb.ObserveRIBRequest_AFISAFI, 0)
		reqIPv4, reqIPv6 := c.Bool("4"), c.Bool("6")
		if !reqIPv4 && !reqIPv6 {
			reqIPv4, reqIPv6 = true, true
		}
		if reqIPv4 {
			afisafis = append(afisafis, pb.ObserveRIBRequest_IPv4Unicast)
		}
		if reqIPv6 {
			afisafis = append(afisafis, pb.ObserveRIBRequest_IPv6Unicast)
		}

		client := pb.NewRoutingInformationServiceClient(conn)
		for _, afisafi := range afisafis {
			fmt.Printf(" --- Dump %s ---\n", pb.DumpRIBRequest_AFISAFI_name[int32(afisafi)])
			err = observeRIB(client, c.GlobalString("router"), c.GlobalUint64("vrf_id"), c.GlobalString("vrf"), afisafi, c.Bool("allow-unready"))
			if err != nil {
				log.Errorf("ObserveRIB failed: %v", err)
				os.Exit(1)
			}
		}

		return nil
	}

	return cmd
}

func observeRIB(c pb.RoutingInformationServiceClient, routerName string, vrfID uint64, vrf string, afisafi pb.ObserveRIBRequest_AFISAFI, allowUnready bool) error {
	req := &pb.ObserveRIBRequest{
		Router:          routerName,
		VrfId:           vrfID,
		Vrf:             vrf,
		Afisafi:         afisafi,
		AllowUnreadyRib: false,
	}
	if allowUnready {
		req.AllowUnreadyRib = true
	}
	client, err := c.ObserveRIB(context.Background(), req)
	if err != nil {
		return fmt.Errorf("unable to get client: %w", err)
	}

	for {
		r, err := client.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("receive failed: %w", err)
		}

		if r.EndOfRib {
			fmt.Printf("Received End-Of-RIB marker\n")
		}

		if r.Route != nil {
			printRoute(r.Route)
		}
	}
}
