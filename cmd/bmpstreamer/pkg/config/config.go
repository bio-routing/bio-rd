package config

import (
	"fmt"
	"io/ioutil"
	"net"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	BMPRouters []BMPRouter `yaml:"bmp_routers"`
	HTTPPort   uint16      `yaml:"http_port"`
	GRPCPort   uint16      `yaml:"grpc_port"`
}

type BMPRouter struct {
	Addr net.IP `yaml:"addr"`
	Port uint16 `yaml:"port"`
}

func LoadConfig(cfgFilePath string) (*Config, error) {
	cfgFile, err := ioutil.ReadFile(cfgFilePath)
	if err != nil {
		return nil, fmt.Errorf("Unable to read config file: %v", err)
	}

	cfg := &Config{}
	yaml.Unmarshal(cfgFile, cfg)

	return cfg, nil
}
