package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// RISConfig is the config of RIS instance
type RISConfig struct {
	BMPServers     []BMPServer `yaml:"bmp_servers"`
	IgnorePeerASNs []uint32    `yaml:"ignore_peer_asns"`
}

// BMPServer represent a BMP enable Router
type BMPServer struct {
	Address string `yaml:"address"`
	Port    uint16 `yaml:"port"`
	Passive bool   `yaml:"passive"`
}

// LoadConfig loads a RIS config
func LoadConfig(filepath string) (*RISConfig, error) {
	f, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("unable to read config file: %w", err)
	}

	cfg := &RISConfig{}
	err = yaml.Unmarshal(f, cfg)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal failed: %w", err)
	}

	return cfg, nil
}
