package main

import (
	"flag"
	"os"

	"github.com/prometheus/common/log"
	"google.golang.org/grpc"

	pb "github.com/bio-routing/bio-rd/cmd/ris/api"
)

var (
	risAddress = flag.String("ris", "10.11.2.7:4321", "RIS GRPC address")
	routerName = flag.String("router", "", "Router name")
	vrfID      = flag.Uint64("vrf_id", 0, "VRF ID")
)

func main() {
	flag.Parse()

	conn, err := grpc.Dial(*risAddress, grpc.WithInsecure())
	if err != nil {
		log.Errorf("GRPC dial failed: %v", err)
		os.Exit(1)
	}
	defer conn.Close()

	c := pb.NewRoutingInformationServiceClient(conn)
	err = dumpRIB(c, *routerName, *vrfID)
	if err != nil {
		log.Errorf("DumpRIB failed: %v", err)
		os.Exit(1)
	}
}
