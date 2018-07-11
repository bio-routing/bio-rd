package config

type ISISConfig struct {
	NetworkEntityTitle []byte
	Interfaces         []ISISInterfaceConfig
}

type ISISInterfaceConfig struct {
	Name             string
	PointToPoint     bool
	Passive          bool
	ISISLevel2Config ISISLevelConfig
}

type ISISLevelConfig struct {
	HelloInterval uint16
	HoldTime      uint16
	Metric        uint32
}
