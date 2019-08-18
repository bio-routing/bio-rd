package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "riscli"
	app.Usage = "RIS CLI"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "ris",
			Usage: "RIS GRPC address",
			Value: "",
		},
		cli.StringFlag{
			Name:  "router",
			Usage: "Router Name",
			Value: "",
		},
		cli.Uint64Flag{
			Name:  "vrf_id",
			Usage: "VRF ID",
			Value: 0,
		},
	}

	app.Commands = []cli.Command{
		NewDumpLocRIBCommand(),
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
