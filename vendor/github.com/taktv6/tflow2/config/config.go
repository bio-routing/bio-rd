package config

import (
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// Config represents a yaml config file
type Config struct {
	AggregationPeriod    *int64  `yaml:"aggregation_period"`
	DefaultSNMPCommunity *string `yaml:"default_snmp_community"`
	Debug                *int    `yaml:"debug"`
	CompressionLevel     *int    `yaml:"compression_level"`
	DataDir              *string `yaml:"data_dir"`
	Anonymize            *bool   `yaml:"anonymize"`
	CacheTime            *int64  `yaml:"cache_time"`

	NetflowV9       *Server     `yaml:"netflow_v9"`
	IPFIX           *Server     `yaml:"ipfix"`
	Sflow           *Server     `yaml:"sflow"`
	Frontend        *Server     `yaml:"frontend"`
	BGPAugmentation *BGPAugment `yaml:"bgp_augmentation"`
	Agents          []Agent     `yaml:"agents"`
	Annotators      []Annotator `yaml:"annotators"`

	AgentsNameByIP map[string]string
}

// Annotator represents annotator configuration
type Annotator struct {
	Name   string
	Target string
}

// BGPAugment represents BGP augmentation configuration
type BGPAugment struct {
	Enabled     *bool   `yaml:"enabled"`
	BIRDSocket  *string `yaml:"bird_socket"`
	BIRD6Socket *string `yaml:"bird6_socket"`
}

// Server represents a server config
type Server struct {
	Enabled *bool   `yaml:"enabled"`
	Listen  *string `yaml:"listen"`
}

// Agent represents an agent config
type Agent struct {
	Name          *string `yaml:"name"`
	IPAddress     *string `yaml:"ip_address"`
	SNMPCommunity *string `yaml:"snmp_community"`
	SampleRate    *uint64 `yaml:"sample_rate"`
}

var (
	dfltAggregationPeriod    = int64(60)
	dfltDefaultSNMPCommunity = "public"
	dfltDebug                = 0
	dfltSampleRate           = uint64(1)
	dfltCompressionLevel     = 6
	dfltDataDir              = "data"
	dfltAnonymize            = false
	dfltCacheTime            = int64(1800)

	dfltNetflowV9Listen = strPtr(":2055")
	dfltNetflowV9       = Server{
		Enabled: boolPtr(true),
		Listen:  dfltNetflowV9Listen,
	}

	dfltServerEnabled = boolPtr(true)

	dfltIPFIXListen = strPtr(":4739")
	dfltIPFIX       = Server{
		Enabled: boolPtr(true),
		Listen:  dfltIPFIXListen,
	}

	dfltSflowListen = strPtr(":6343")
	dfltSflow       = Server{
		Enabled: boolPtr(true),
		Listen:  dfltSflowListen,
	}

	dfltFrontendListen = strPtr(":4444")
	dfltFrontend       = Server{
		Enabled: boolPtr(true),
		Listen:  dfltFrontendListen,
	}

	dfltBIRDSocket      = strPtr("/var/run/bird/bird.ctl")
	dfltBIRD6Socket     = strPtr("/var/run/bird/bird6.ctl")
	dfltBGPAugmentation = BGPAugment{
		Enabled:     boolPtr(false),
		BIRDSocket:  dfltBIRDSocket,
		BIRD6Socket: dfltBIRD6Socket,
	}
)

// New reads a configuration file and returns a Config
func New(filename string) (*Config, error) {
	cfgFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Unable to read config file %s: %v", filename, err)
	}

	cfg := &Config{}
	err = yaml.Unmarshal(cfgFile, cfg)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse yaml file: %v", err)
	}

	cfg.defaults()

	cfg.AgentsNameByIP = make(map[string]string)
	for _, agent := range cfg.Agents {
		if _, ok := cfg.AgentsNameByIP[*agent.IPAddress]; ok {
			return nil, fmt.Errorf("Duplicate agent: %s", *agent.Name)
		}
		cfg.AgentsNameByIP[*agent.IPAddress] = *agent.Name
	}

	return cfg, nil
}

func (cfg *Config) defaults() {
	if cfg.AggregationPeriod == nil {
		cfg.AggregationPeriod = int64Ptr(dfltAggregationPeriod)
	}
	if cfg.DefaultSNMPCommunity == nil {
		cfg.DefaultSNMPCommunity = strPtr(dfltDefaultSNMPCommunity)
	}
	if cfg.Debug == nil {
		cfg.Debug = intPtr(dfltDebug)
	}
	if cfg.CompressionLevel == nil {
		cfg.CompressionLevel = intPtr(dfltCompressionLevel)
	}
	if cfg.DataDir == nil {
		cfg.DataDir = strPtr(dfltDataDir)
	}
	if cfg.Anonymize == nil {
		cfg.Anonymize = boolPtr(dfltAnonymize)
	}
	if cfg.CacheTime == nil {
		cfg.CacheTime = int64Ptr(dfltCacheTime)
	}

	if cfg.NetflowV9 == nil {
		cfg.NetflowV9 = srvPtr(dfltNetflowV9)
	}
	if cfg.NetflowV9.Listen == nil {
		cfg.NetflowV9.Listen = dfltNetflowV9Listen
	}
	if cfg.NetflowV9.Enabled == nil {
		cfg.NetflowV9.Enabled = dfltServerEnabled
	}

	if cfg.IPFIX == nil {
		cfg.IPFIX = srvPtr(dfltIPFIX)
	}
	if cfg.IPFIX.Listen == nil {
		cfg.IPFIX.Listen = dfltIPFIXListen
	}
	if cfg.IPFIX.Enabled == nil {
		cfg.IPFIX.Enabled = dfltServerEnabled
	}

	if cfg.Sflow == nil {
		cfg.Sflow = srvPtr(dfltSflow)
	}
	if cfg.Sflow.Listen == nil {
		cfg.Sflow.Listen = dfltSflowListen
	}
	if cfg.Sflow.Enabled == nil {
		cfg.Sflow.Enabled = dfltServerEnabled
	}

	if cfg.Frontend == nil {
		cfg.Frontend = srvPtr(dfltFrontend)
	}
	if cfg.Frontend.Listen == nil {
		cfg.Frontend.Listen = dfltFrontendListen
	}
	if cfg.Frontend.Enabled == nil {
		cfg.Frontend.Enabled = dfltServerEnabled
	}

	if cfg.BGPAugmentation == nil {
		cfg.BGPAugmentation = bgpPtr(dfltBGPAugmentation)
	}
	if cfg.BGPAugmentation.BIRDSocket == nil {
		cfg.BGPAugmentation.BIRDSocket = dfltBIRDSocket
	}
	if cfg.BGPAugmentation.BIRD6Socket == nil {
		cfg.BGPAugmentation.BIRD6Socket = dfltBIRD6Socket
	}

	if cfg.Agents != nil {
		for key, agent := range cfg.Agents {
			if agent.SNMPCommunity == nil {
				cfg.Agents[key].SNMPCommunity = strPtr(*cfg.DefaultSNMPCommunity)
			}
			if agent.SampleRate == nil {
				cfg.Agents[key].SampleRate = uint64Ptr(dfltSampleRate)
			}
		}
	}
}

func bgpPtr(bgp BGPAugment) *BGPAugment {
	return &bgp
}

func uint64Ptr(x uint64) *uint64 {
	return &x
}

func srvPtr(srv Server) *Server {
	return &srv
}

func strPtr(str string) *string {
	return &str
}

func boolPtr(v bool) *bool {
	return &v
}

func intPtr(x int) *int {
	return &x
}

func int64Ptr(x int64) *int64 {
	return &x
}
