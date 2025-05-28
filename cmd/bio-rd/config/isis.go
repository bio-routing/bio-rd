package config

const (
	defaultHelloInterval      = 9
	defaultHoldTime           = 27
	lspMinLifetime            = 350
	lspDefaultLifetimeSeconds = 1200
)

// ISIS config
type ISIS struct {
	// description: |
	//   Network entity title for this instance
	NETs []string `yaml:"NETs"`
	// description: |
	//   Configuration for the Level 1 adjacency
	Level1 *ISISLevel `yaml:"level1"`
	// description: |
	//   Configuration for the level 2 adjacency
	Level2 *ISISLevel `yaml:"level2"`
	// description: |
	//   Interface related configuration
	Interfaces []*ISISInterface `yaml:"interfaces"`
	// description: |
	//   Amount of time a link-state PDU should persist in the network
	//   Expressed in seconds
	LSPLifetime uint16 `yaml:"lsp_lifetime"`
}

// ISISLevel level config
type ISISLevel struct {
	// description: |
	//   Disables this level for the instance
	Disable bool `yaml:"disable"`
	// description: |
	//   Password for authentication
	AuthenticationKey string `yaml:"authentication_key"`
	// description: |
	//   Disable authentication for the Complete Sequence Number PDUs
	NoCSNPAuthentication bool `yaml:"no_csnp_authentication"`
	// description: |
	//   DIsable authentication for hello messages
	NoHelloAuthentication bool `yaml:"no_hello_authentication"`
	// description: |
	//   Disable authentication for the Partial Sequence Number PDUs
	NoPSNPAuthentication bool `yaml:"no_psnp_authentication"`
	// description: |
	//   Enable sending and receiving wide metrics only for this level
	WideMetricsOnly bool `yaml:"wide_metrics_only"`
}

// ISISInterface interface config
type ISISInterface struct {
	// description: |
	//   Name of the interface to configure
	Name string `yaml:"name"`
	// description: |
	//   Configure interface as passive
	Passive bool `yaml:"passive"`
	// description: |
	//   Configure interface as point-to-point
	PointToPoint bool `yaml:"point_to_point"`
	// description: |
	//   Level 1 configuration parameters for the interface
	Level1 *ISISInterfaceLevel `yaml:"level1"`
	// description: |
	//   Level 2 configuration parameters for the interface
	Level2 *ISISInterfaceLevel `yaml:"level2"`
}

// ISISInterfaceLevel interface level config
type ISISInterfaceLevel struct {
	// description: |
	//   Disable this level for the interface
	Disable bool `yaml:"disable"`
	// description: |
	//   Hello interval
	//   Expressed in seconds
	HelloInterval uint16 `yaml:"hello_interval"`
	// description: |
	//   Hold time
	//   Expressed in seconds
	HoldTime uint16 `yaml:"hold_time"`
	// description: |
	//   Metric for the interface in this level
	Metric uint32 `yaml:"metric"`
	// description: |
	//   Configures interface as passive
	Passive bool `yaml:"passive"`
	// description: |
	//   Configures the device priority to become a designated router for this level
	//   Value range: 0-127
	Priority uint8 `yaml:"priority"`
}

func (i *ISIS) loadDefaults() {
	if i.LSPLifetime == 0 {
		i.LSPLifetime = lspDefaultLifetimeSeconds
	}

	if i.LSPLifetime < lspMinLifetime {
		i.LSPLifetime = lspMinLifetime
	}

	for _, ifa := range i.Interfaces {
		ifa.loadDefaults()
	}
}

func (i *ISISInterface) loadDefaults() {
	if i.Level1 != nil {
		i.Level1.loadDefaults()
	}

	if i.Level2 != nil {
		i.Level2.loadDefaults()
	}
}

func (i *ISISInterfaceLevel) loadDefaults() {
	if i.HelloInterval == 0 {
		i.HelloInterval = defaultHelloInterval
	}

	if i.HoldTime == 0 {
		i.HoldTime = defaultHoldTime
	}
}

func (i *ISIS) InterfaceConfigured(name string) bool {
	for _, x := range i.Interfaces {
		if x.Name == name {
			return true
		}
	}

	return false
}
