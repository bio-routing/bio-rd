package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	PolicyOptions    *PolicyOptions     `yaml:"policy_options"`
	RoutingInstances []*RoutingInstance `yaml:"routing_instances"`
	RoutingOptions   *RoutingOptions    `yaml:"routing_options"`
	Protocols        *Protocols         `yaml:"protocols"`
}

func (c *Config) load() error {
	if c.RoutingOptions == nil {
		return fmt.Errorf("config is lacking routing_options")
	}

	if c.PolicyOptions != nil {
		err := c.PolicyOptions.load()
		if err != nil {
			return fmt.Errorf("unable to load policy_options: %w", err)
		}
	}

	err := c.RoutingOptions.load()
	if err != nil {
		return fmt.Errorf("error in routing_options: %w", err)
	}

	for _, ri := range c.RoutingInstances {
		err := ri.load()
		if ri != nil {
			return err
		}
	}

	if c.Protocols != nil {
		localAS := c.RoutingOptions.AutonomousSystem

		err := c.Protocols.load(localAS, c.PolicyOptions)
		if err != nil {
			return fmt.Errorf("Failed to load protocols: %w", err)
		}
	}

	return nil
}

// GetConfig gets the configuration
func GetConfig(filePath string) (*Config, error) {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to read file: %w", err)
	}

	c := &Config{}
	err = yaml.Unmarshal(file, c)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal: %w", err)
	}

	err = c.load()
	if err != nil {
		return nil, err
	}

	return c, nil
}
