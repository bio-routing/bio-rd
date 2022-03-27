package config

import "fmt"

type Protocols struct {
	BGP  *BGP  `yaml:"bgp"`
	ISIS *ISIS `yaml:"isis"`
}

func (p *Protocols) load(localAS uint32, policyOptions *PolicyOptions) error {
	err := p.BGP.load(localAS, policyOptions)
	if err != nil {
		return fmt.Errorf("BGP error: %w", err)
	}

	if p.ISIS != nil {
		p.ISIS.loadDefaults()
	}

	return nil
}
