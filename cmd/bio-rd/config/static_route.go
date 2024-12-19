package config

type StaticRoute struct {
	// description: |
	//   Prefix for the route
	Prefix  string `yaml:"prefix"`
	// description: |
	//   Makes this route a blackhole
	Discard bool `yaml:"discard"`
	// description: |
	//   Next hop for the route
	NextHop string `yaml:"next_hop"`
	// description: |
	//   ??
	Resolve bool `yaml:"resolve"`
}
