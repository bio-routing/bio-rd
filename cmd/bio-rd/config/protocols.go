package config

import "github.com/pkg/errors"

type Protocols struct {
	BGP *BGP `yaml:"bgp"`
}

func (p *Protocols) load(localAS uint32, policyOptions *PolicyOptions) error {
	err := p.BGP.load(localAS, policyOptions)
	if err != nil {
		return errors.Wrap(err, "BGP error")
	}

	return nil
}
