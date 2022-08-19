package main

import (
	"os"

	"github.com/urfave/cli"

	"github.com/bio-routing/bio-rd/util/log"
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
		cli.StringFlag{
			Name:  "vrf",
			Usage: "VRF",
			Value: "",
		},
	}

	app.Commands = []cli.Command{
		NewObserveRIBCommand(),
		NewDumpLocRIBCommand(),
		NewLPMCommand(),
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
