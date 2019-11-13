package config

type StaticRoute struct {
	Prefix  string
	Discard bool
	NextHop string
	Resolve bool
}
