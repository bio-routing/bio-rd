package main

import (
	"os"
	"time"

	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/protocols/isis/server"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	log "github.com/sirupsen/logrus"
)

func main() {
	cfg := &config.ISISConfig{
		NETs: []config.NET{
			{
				AFI:      0x49,
				AreaID:   types.AreaID{0, 0x01, 0, 0x10},
				SystemID: types.SystemID{10, 20, 30, 40, 50, 60},
				SEL:      0x00,
			},
		},
		Interfaces: []config.ISISInterfaceConfig{
			{
				Name:    "virbr2",
				Passive: false,
				P2P:     true,
				ISISLevel2Config: &config.ISISLevelConfig{
					HelloInterval: 1,
					HoldTime:      3,
					Metric:        10,
					Priority:      0,
				},
			},
		},
		MinLSPTransmissionInterval: 100,
	}

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
