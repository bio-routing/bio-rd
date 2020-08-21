package config

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// RISMirrorConfig is the config of RISMirror instance
type RISMirrorConfig struct {
	RIBConfigs []RIBConfig `yaml:"ribs"`
}

// RIBConfig is a RIB configuration
type RIBConfig struct {
	Router          string   `yaml:"router"`
	VRFs            []string `yaml:"vrfs"`
	IPVersions      []uint8  `yaml:"IPVersions"`
	SrcRISInstances []string `yaml:"source_ris_instances"`
}

// LoadConfig loads a RISMirror config
func LoadConfig(filepath string) (*RISMirrorConfig, error) {
	f, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read config file")
	}

	cfg := &RISMirrorConfig{}
	err = yaml.Unmarshal(f, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "Unmarshal failed")
	}

	return cfg, nil
}

// GetRISInstances returns a list of all RIS instances in the config
func (rismc *RISMirrorConfig) GetRISInstances() []string {
	instances := make(map[string]struct{})

	for _, r := range rismc.RIBConfigs {
		for _, s := range r.SrcRISInstances {
			instances[s] = struct{}{}
		}
	}

	ret := make([]string, 0)
	for instance := range instances {
		ret = append(ret, instance)
	}

	return ret
}
