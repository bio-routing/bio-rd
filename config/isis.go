package config

type ISISConfig struct {
	NETs                       []NET
	Interfaces                 []ISISInterfaceConfig
	TrafficEngineeringRouterID [4]byte
}

type ISISInterfaceConfig struct {
	Name             string
	Passive          bool
	P2P              bool
	ISISLevel1Config *ISISLevelConfig
	ISISLevel2Config *ISISLevelConfig
}

type ISISLevelConfig struct {
	HelloInterval uint16
	HoldTime      uint16
	Metric        uint32
	Priority      uint8
}
