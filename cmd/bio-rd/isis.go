package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bio-routing/bio-rd/cmd/bio-rd/config"
	"github.com/bio-routing/bio-rd/protocols/isis/server"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/pkg/errors"
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
		isisSrv = server.New(nets, ds)
		err := isisSrv.Start()
		if err != nil {
			return errors.Wrap(err, "Unable to start ISIS server")
		}
	}

	for _, ifa := range isis.Interfaces {
		err := isisSrv.AddInterface(&server.InterfaceConfig{})
		if err != nil {
			return errors.Wrapf(err, "Unable to add interface: %s", ifa.Name)
		}
	}

	return fmt.Errorf("ISIS not implemented yet")
}

func parseNETs(nets []string) ([]*types.NET, error) {
	ret := make([]*types.NET, 0, len(nets))

	for _, net := range nets {
		n, err := types.ParseNET([]byte(net))
		if err != nil {
			return nil, errors.Wrapf(err, "Unable to parse NET %q", net)
		}

		ret = append(ret, n)
	}

	return ret, nil
}

func netStringToByteSlice(s string) ([]byte, error) {
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
