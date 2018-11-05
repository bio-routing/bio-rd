package main

import (
	"context"
	"fmt"
	"os"

	pb "github.com/bio-routing/bio-rd/cmd/bmp-streamer/pkg/bmpstreamer"
	"github.com/bio-routing/bio-rd/net"
	netapi "github.com/bio-routing/bio-rd/net/api"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	rtr, err := net.IPFromBytes([]byte{10, 0, 255, 0})
	if err != nil {
		log.Errorf("Unable to parse IP: %v", err)
		os.Exit(1)
	}

	fmt.Printf("Router: %s\n", rtr.String())

	conn, err := grpc.Dial("127.0.0.1:8081", grpc.WithInsecure())
	if err != nil {
		log.Errorf("Dial failed: %v", err)
		os.Exit(1)
	}
	defer conn.Close()

	c := pb.NewRIBServiceClient(conn)
	streamClient, err := c.AdjRIBInStream(context.Background(), &pb.AdjRIBInStreamRequest{
		Router: &netapi.IP{
			Higher:  rtr.Higher(),
			Lower:   rtr.Lower(),
			Version: netapi.IP_IPv4,
		},
	})

	if err != nil {
		log.Errorf("Unable to start streaming RPC: %v", err)
		os.Exit(1)
	}

	for {
		fmt.Printf("Reading stream...\n")
		u, err := streamClient.Recv()
		if err != nil {
			log.Errorf("stream stopped: %v", err)
			break
		}
		fmt.Printf("Update: %v\n", u)
	}

}
