package config

import (
	"fmt"
	"io/ioutil"
	"net"

	"github.com/bio-routing/bio-rd/routingtable/vrf"
	"gopkg.in/yaml.v2"
)

// RISMirrorConfig is the config of RISMirror instance
type RISMirrorConfig struct {
	RIBConfigs []*RIBConfig `yaml:"ribs"`
}

// RIBConfig is a RIB configuration
type RIBConfig struct {
	Router          string `yaml:"router"`
	router          net.IP
	VRFs            []string `yaml:"vrfs"`
	vrfs            []uint64
	IPVersions      []uint8  `yaml:"IPVersions"`
	SrcRISInstances []string `yaml:"source_ris_instances"`
}

// GetRouter gets a routers IP address
func (rc *RIBConfig) GetRouter() net.IP {
	return rc.router
}

// GetVRFs gets a routers VRFs
func (rc *RIBConfig) GetVRFs() []uint64 {
	return rc.vrfs
}

// LoadConfig loads a RISMirror config
func LoadConfig(filepath string) (*RISMirrorConfig, error) {
	f, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("Unable to read config file: %w", err)
	}

	cfg := &RISMirrorConfig{}
	err = yaml.Unmarshal(f, cfg)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal failed: %w", err)
	}

	for _, rc := range cfg.RIBConfigs {
		err := rc.loadRouter()
		if err != nil {
			return nil, fmt.Errorf("Unable to load router config: %w", err)
		}

		err = rc.loadVRFs()
		if err != nil {
			return nil, fmt.Errorf("Unable to load VRFs: %w", err)
		}
	}

	return cfg, nil
}

func (r *RIBConfig) loadRouter() error {
	addr := net.ParseIP(r.Router)
	if addr == nil {
		return fmt.Errorf("Unable to parse routers IP: %q", r.Router)
	}

	r.router = addr
	return nil
}

func (r *RIBConfig) loadVRFs() error {
	for _, vrfHuman := range r.VRFs {
		vrfRD, err := vrf.ParseHumanReadableRouteDistinguisher(vrfHuman)
		if err != nil {
			return fmt.Errorf("Unable to parse VRF identifier: %w", err)
		}

		r.vrfs = append(r.vrfs, vrfRD)
	}

	return nil
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
