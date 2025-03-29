package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	// description: |
	//   A list of policy statements, filters and prefix lists that can be used to filter route imports and exports.
	//   <a href="policy.md">Detailed documentation</a>
	//   Example:
	//   policy_statements:
	//     - name: "PeerA-In"
	//         terms:
	//           - name: "Reject_certain_stuff"
	//             from:
	//               route_filters:
	//                  - prefix: "198.51.100.0/24"
	//                    matcher: "orlonger"
	//                  - prefix: "203.0.113.0/25"
	//                    matcher: "exact"
	//                  - prefix: "203.0.113.128/25"
	//                    matcher: "exact"
	PolicyOptions *PolicyOptions `yaml:"policy_options"`
	// description: |
	//    List of routing instances
	//    <a href="routing_instance.md">Configuration parameters</a>
	RoutingInstances []*RoutingInstance `yaml:"routing_instances"`
	// description: |
	//   Routing options
	//   <a href="routing_options.md">parameter documentation</a>
	//   Allowed values:
	//   - <a href="static_route.md">static_routes</a>
	//   - router_id
	//   - autonomous_system
	RoutingOptions *RoutingOptions `yaml:"routing_options"`
	// description: |
	//   Here you define the specific configuration parameters for each protocol.
	//   <a href="protocols.md">documentation</a>
	//   Available protocols:
	//     - <a href="bgp.md">bgp</a>
	//     - <a href="isis.md">is-is</a>
	Protocols *Protocols `yaml:"protocols"`
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
	file, err := os.ReadFile(filePath)
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
