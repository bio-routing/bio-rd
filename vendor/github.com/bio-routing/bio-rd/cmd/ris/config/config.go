package config

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// RISConfig is the config of RIS instance
type RISConfig struct {
	BMPServers []BMPServer `yaml:"bmp_servers"`
}

// BMPServer represent a BMP enable Router
type BMPServer struct {
	Address string `yaml:"address"`
	Port    uint16 `yaml:"port"`
}

// LoadConfig loads a RIS config
func LoadConfig(filepath string) (*RISConfig, error) {
	f, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read config file")
	}

	cfg := &RISConfig{}
	err = yaml.Unmarshal(f, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "Unmarshal failed")
	}

	return cfg, nil
}
