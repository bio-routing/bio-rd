package main

import (
	"os"
	"time"

	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/protocols/isis/server"
	"github.com/prometheus/common/log"
)

func main() {
	cfg := &config.ISISConfig{}

	s := server.NewISISServer(cfg)
	err := s.Start()
	if err != nil {
		log.Errorf("Unable to start ISIS server: %v", err)
		os.Exit(1)
	}

	go func() {
		t := time.NewTicker(time.Second * 10)
		for {
			<-t.C
			s.DumpLSDB()
		}
	}()

	select {}
}
