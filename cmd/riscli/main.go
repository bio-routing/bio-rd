package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/bio-routing/bio-rd/util/log"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	log.SetLogger(log.NewLogrusWrapper(logger))

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
		cli.BoolFlag{
			Name:  "tls",
			Usage: "use a tls-encrypted grpc connection to the ris",
		},
	}

	app.Commands = []cli.Command{
		NewObserveRIBCommand(),
		NewDumpLocRIBCommand(),
		NewLPMCommand(),
		NewGetRoutersCommand(),
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
