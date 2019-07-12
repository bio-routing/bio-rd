package main

import (
	"context"

	pb "github.com/bio-routing/bio-rd/cmd/ris/api"
	"github.com/pkg/errors"
)

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
