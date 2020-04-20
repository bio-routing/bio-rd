package config

import "github.com/pkg/errors"

// Protocols is routing protocol config
type Protocols struct {
	BGP  *BGP  `yaml:"bgp"`
	ISIS *ISIS `yaml:"isis"`
}

func (p *Protocols) load(localAS uint32, policyOptions *PolicyOptions) error {
	err := p.BGP.load(localAS, policyOptions)
	if err != nil {
		return errors.Wrap(err, "BGP error")
	}

	if p.ISIS != nil {
		p.ISIS.loadDefaults()
	}

	return nil
}
