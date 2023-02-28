package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/bio-routing/bio-rd/cmd/ris-lg/lg"
	"github.com/bio-routing/bio-rd/util/log"
	"github.com/sirupsen/logrus"
)

var (
	risAddr        = flag.String("ris.addr", "", "RIS address")
	httpListenAddr = flag.String("http.listen-addr", ":8080", "HTTP listening address")
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	log.SetLogger(log.NewLogrusWrapper(logger))

	flag.Parse()

	if *risAddr == "" {
		log.Errorf("ris.addr is a mandatory parameter")
		os.Exit(1)
	}

	lGlass, err := lg.New(*risAddr)
	if err != nil {
		log.Errorf("unable to create looking glass: %v", err)
		os.Exit(1)
	}

	http.HandleFunc("/", lGlass.Index)
	http.HandleFunc("/routes", lGlass.Routes)
	http.FileServer(http.FS(lg.Res))
	err = http.ListenAndServe(*httpListenAddr, nil)
	if err != nil {
		log.Errorf("http.ListenAndServe failed: %v", err)
		os.Exit(1)
	}
}
