package config

const (
	defaultHelloInterval = 9
	defaultHoldTime      = 27
	lspMinLifetime       = 350
)

// ISIS config
type ISIS struct {
	NETs        []string         `yaml:"NETs"`
	Level1      *ISISLevel       `yaml:"level1"`
	Level2      *ISISLevel       `yaml:"level2"`
	Interfaces  []*ISISInterface `yaml:"interfaces"`
	LSPLifetime uint16           `yaml:"lsp_lifetime"`
}

//ISISLevel level config
type ISISLevel struct {
	Disable               bool   `yaml:"disable"`
	AuthenticationKey     string `yaml:"authentication_key"`
	NoCSNPAuthentication  bool   `yaml:"no_csnp_authentication"`
	NoHelloAuthentication bool   `yaml:"no_hello_authentication"`
	NoPSNPAuthentication  bool   `yaml:"no_psnp_authentication"`
	WideMetricsOnly       bool   `yaml:"wide_metrics_only"`
}

// ISISInterface interface config
type ISISInterface struct {
	Name         string              `yaml:"name"`
	Passive      bool                `yaml:"passive"`
	PointToPoint bool                `yaml:"point_to_point"`
	Level1       *ISISInterfaceLevel `yaml:"level1"`
	Level2       *ISISInterfaceLevel `yaml:"level2"`
}

// ISISInterfaceLevel interface level config
type ISISInterfaceLevel struct {
	Disable       bool   `yaml:"disable"`
	HelloInterval uint16 `yaml:"hello_interval"`
	HoldTime      uint16 `yaml:"hold_time"`
	Metric        uint32 `yaml:"metric"`
	Passive       bool   `yaml:"passive"`
	Priority      uint8  `yaml:"priority"`
}

func (i *ISIS) loadDefaults() {
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
