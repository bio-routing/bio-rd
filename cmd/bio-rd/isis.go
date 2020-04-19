package main

import (
	"fmt"

	"github.com/bio-routing/bio-rd/cmd/bio-rd/config"
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

	return fmt.Errorf("ISIS not implemented yet")
}

func parseNETs(nets []string) ([]types.NET, error) {
	ret := make([]types.NET, 0, len(nets))

	for _, net := range nets {
		n, err := types.parseNETs(net)
		if err != nil {
			return nil, errors.Wrapf(err, "Unable to parse NET %q", net)
		}

		ret = append(ret, n)
	}

	return ret, nil
}
