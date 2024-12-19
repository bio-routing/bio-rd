package config

import "fmt"

type Protocols struct {
	// description: |
	//   parameters for the bgp. See bgp.md file for details
	BGP  *BGP  `yaml:"bgp"`
	// description: |
	//   parameters for the is-is protocol. See is-is.md for details
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
