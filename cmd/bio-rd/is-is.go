package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bio-routing/bio-rd/cmd/bio-rd/config"
	"github.com/bio-routing/bio-rd/protocols/isis/server"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/bio-routing/bio-rd/util/log"
)

func configureProtocolsISIS(isis *config.ISIS) error {
	if len(isis.NETs) == 0 {
		return fmt.Errorf("no Network Entity Titles (NETs, ISO addresses) given")
	}

	nets, err := parseNETs(isis.NETs)
	if err != nil {
		return err
	}

	if isisSrv == nil {
		var err error
		isisSrv, err = server.New(nets, ds, isis.LSPLifetime)
		if err != nil {
			return fmt.Errorf("unable to create ISIS server: %w", err)
		}

		isisSrv.Start()
	}

	configuredInterfaces := isisSrv.GetInterfaceNames()
	for _, ifa := range isis.Interfaces {
		if strSliceContains(configuredInterfaces, ifa.Name) {
			log.Debugf("ISIS: Interface %q is already configured", ifa.Name)
			continue
		}

		log.Infof("ISIS: Adding interface %s to ISIS server", ifa.Name)
		err := isisSrv.AddInterface(&server.InterfaceConfig{
			Name:         ifa.Name,
			Passive:      ifa.Passive,
			PointToPoint: ifa.PointToPoint,
			Level1:       translateInterfaceLevelConfig(ifa.Level1),
			Level2:       translateInterfaceLevelConfig(ifa.Level2),
		})
		if err != nil {
			return fmt.Errorf("unable to add interface: %s: %w", ifa.Name, err)
		}
	}

	for _, ifaName := range configuredInterfaces {
		if !isis.InterfaceConfigured(ifaName) {
			err := isisSrv.RemoveInterface(ifaName)
			if err != nil {
				return fmt.Errorf("failed to remove ISIS interface %q", ifaName)
			}
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
			return nil, fmt.Errorf("unable to convert NET hex string into ints: %w", err)
		}

		n, err := types.ParseNET(b)
		if err != nil {
			return nil, fmt.Errorf("unable to parse NET %q: %w", net, err)
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
			return nil, fmt.Errorf("unable to parse int: %w", err)
		}

		ret = append(ret, uint8(x))
	}

	return ret, nil
}

func strSliceContains(haystack []string, needle string) bool {
	for _, x := range haystack {
		if x == needle {
			return true
		}
	}

	return false
}
