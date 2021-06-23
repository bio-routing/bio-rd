package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bio-routing/bio-rd/cmd/bio-rd/config"
	"github.com/bio-routing/bio-rd/protocols/isis/server"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func configureProtocolsISIS(isis *config.ISIS) error {
	if len(isis.NETs) == 0 {
		return fmt.Errorf("No Network Entity Titles (NETs, ISO addresses) given")
	}

	nets, err := parseNETs(isis.NETs)
	if err != nil {
		return err
	}

	if isisSrv == nil {
		var err error
		isisSrv, err = server.New(nets, ds, isis.LSPLifetime)
		if err != nil {
			return errors.Wrap(err, "Unable to create ISIS server")
		}

		err = isisSrv.Start()
		if err != nil {
			return errors.Wrap(err, "Unable to start ISIS server")
		}
	}

	for _, ifa := range isis.Interfaces {
		log.Infof("ISIS: Adding interface %s to ISIS server", ifa.Name)
		err := isisSrv.AddInterface(&server.InterfaceConfig{
			Name:         ifa.Name,
			Passive:      ifa.Passive,
			PointToPoint: ifa.PointToPoint,
			Level1:       translateInterfaceLevelConfig(ifa.Level1),
			Level2:       translateInterfaceLevelConfig(ifa.Level2),
		})
		if err != nil {
			return errors.Wrapf(err, "Unable to add interface: %s", ifa.Name)
		}
	}

	return nil
}

func translateInterfaceLevelConfig(c *config.ISISInterfaceLevel) *server.InterfaceLevelConfig {
	if c == nil {
		return nil
	}

	return &server.InterfaceLevelConfig{
		HelloInterval: c.HelloInterval,
		HoldingTimer:  c.HoldTime,
		Metric:        c.Metric,
		Passive:       c.Passive,
		Priority:      c.Priority,
	}
}

func parseNETs(nets []string) ([]*types.NET, error) {
	ret := make([]*types.NET, 0, len(nets))

	for _, net := range nets {
		b, err := parseHexString(net)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to convert NET hex string into ints")
		}

		n, err := types.ParseNET(b)
		if err != nil {
			return nil, errors.Wrapf(err, "Unable to parse NET %q", net)
		}

		ret = append(ret, n)
	}

	return ret, nil
}

// parseHexString converts a string of hex numbers into a byte slice
func parseHexString(s string) ([]byte, error) {
	ret := make([]byte, 0)

	s = strings.Replace(s, ".", "", -1)
	runes := []rune(s)

	for i := 0; i < len(runes); i += 2 {
		x, err := strconv.ParseInt(fmt.Sprintf("%c%c", runes[i], runes[i+1]), 16, 8)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse int")
		}

		ret = append(ret, uint8(x))
	}

	return ret, nil
}
