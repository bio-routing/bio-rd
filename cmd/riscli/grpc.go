package main

import (
	"crypto/tls"
	"os"

	"github.com/bio-routing/bio-rd/util/log"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func SetupGRPCClient(c *cli.Context) *grpc.ClientConn {
	grpcDial := grpc.WithInsecure()
	if c.GlobalBool("tls") {
		config := &tls.Config{}
		grpcDial = grpc.WithTransportCredentials(credentials.NewTLS(config))
	}
	conn, err := grpc.Dial(c.GlobalString("ris"), grpcDial)
	if err != nil {
		log.WithError(err).Error("GRPC dial failed: %v")
		os.Exit(1)
	}
	return conn
}
