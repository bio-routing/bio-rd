package config

import (
	"fmt"
	"time"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/routingtable/filter"
)

type BGP struct {
	Groups []*BGPGroup `yaml:"groups"`
}

func (b *BGP) load(localAS uint32, policyOptions *PolicyOptions) error {
	for _, g := range b.Groups {
		err := g.load(localAS, policyOptions)
		if err != nil {
			return err
		}
	}

	return nil
}

type BGPGroup struct {
	Name              string `yaml:"name"`
	LocalAddress      string `yaml:"local_address"`
	LocalAddressIP    *bnet.IP
	TTL               uint8      `yaml:"ttl"`
	AuthenticationKey string     `yaml:"authentication_key"`
	PeerAS            uint32     `yaml:"peer_as"`
	LocalAS           uint32     `yaml:"local_as"`
	HoldTime          uint16     `yaml:"hold_time"`
	Multipath         *Multipath `yaml:"multipath"`
	Import            []string   `yaml:"import"`
	ImportFilterChain filter.Chain
	Export            []string `yaml:"export"`
	ExportFilterChain filter.Chain
	RouteServerClient bool           `yaml:"route_server_client"`
	Passive           bool           `yaml:"passive"`
	Neighbors         []*BGPNeighbor `yaml:"neighbors"`
}

func (bg *BGPGroup) load(localAS uint32, policyOptions *PolicyOptions) error {
	if bg.LocalAS == 0 {
		bg.LocalAS = localAS
	}

	if bg.LocalAddress != "" {
		a, err := bnet.IPFromString(bg.LocalAddress)
		if err != nil {
			return fmt.Errorf("unable to parse BGP local address: %q: %w", bg.LocalAddress, err)
		}

		bg.LocalAddressIP = a.Dedup()
	}

	if bg.HoldTime == 0 {
		bg.HoldTime = 90
	}

	for i := range bg.Import {
		f := policyOptions.getPolicyStatementFilter(bg.Import[i])
		if f == nil {
			return fmt.Errorf("policy statement %q undefined", bg.Import[i])
		}

		bg.ImportFilterChain = append(bg.ImportFilterChain, f)
	}

	for i := range bg.Export {
		f := policyOptions.getPolicyStatementFilter(bg.Export[i])
		if f == nil {
			return fmt.Errorf("policy statement %q undefined", bg.Export[i])
		}

		bg.ExportFilterChain = append(bg.ExportFilterChain, f)
	}

	for _, bn := range bg.Neighbors {
		if bn.RouteServerClient == nil {
			bn.RouteServerClient = &bg.RouteServerClient
		}

		if bn.Passive == nil {
			bn.Passive = &bg.Passive
		}

		if bn.LocalAddress == "" {
			bn.LocalAddressIP = bg.LocalAddressIP
		}

		if bn.TTL == 0 {
			bn.TTL = bg.TTL
		}

		if bn.AuthenticationKey == "" {
			bn.AuthenticationKey = bg.AuthenticationKey
		}

		if bn.LocalAS == 0 {
			bn.LocalAS = localAS
		}

		if bn.LocalAS == 0 {
			return fmt.Errorf("local_as 0 is invalid")
		}

		if bn.PeerAS == 0 {
			bn.PeerAS = bg.PeerAS
		}

		if bn.PeerAS == 0 {
			return fmt.Errorf("peer_as 0 is invalid")
		}

		if bn.HoldTime == 0 {
			bn.HoldTime = bg.HoldTime
		}

		err := bn.load(policyOptions)
		if err != nil {
			return err
		}
	}

	return nil
}

type Multipath struct {
	Enable    bool `yaml:"enable"`
	MulipleAS bool `yaml:"multiple_as"`
}

type BGPNeighbor struct {
	PeerAddress                string `yaml:"peer_address"`
	PeerAddressIP              *bnet.IP
	LocalAddress               string `yaml:"local_address"`
	LocalAddressIP             *bnet.IP
	Disabled                   bool   `yaml:"disabled"`
	TTL                        uint8  `yaml:"ttl"`
	AuthenticationKey          string `yaml:"authentication_key"`
	PeerAS                     uint32 `yaml:"peer_as"`
	LocalAS                    uint32 `yaml:"local_as"`
	HoldTime                   uint16 `yaml:"hold_time"`
	HoldTimeDuration           time.Duration
	Multipath                  *Multipath `yaml:"multipath"`
	Import                     []string   `yaml:"import"`
	ImportFilterChain          filter.Chain
	Export                     []string `yaml:"export"`
	ExportFilterChain          filter.Chain
	RouteServerClient          *bool  `yaml:"route_server_client"`
	Passive                    *bool  `yaml:"passive"`
	ClusterID                  string `yaml:"cluster_id"`
	ClusterIDIP                *bnet.IP
	IPv4                       *AddressFamilyConfig `yaml:"ipv4"`
	IPv6                       *AddressFamilyConfig `yaml:"ipv6"`
	AdvertiseIPv4MultiProtocol bool                 `yaml:"advertise_ipv4_multiprotocol"`
}

func (bn *BGPNeighbor) load(policyOptions *PolicyOptions) error {
	if bn.PeerAS == 0 {
		return fmt.Errorf("peer %q is lacking peer as number", bn.PeerAddress)
	}

	if bn.PeerAddress == "" {
		return fmt.Errorf("mandatory parameter BGP peer address is empty")
	}

	if bn.LocalAddress != "" {
		a, err := bnet.IPFromString(bn.LocalAddress)
		if err != nil {
			return fmt.Errorf("unable to parse BGP local address: %w", err)
		}

		bn.LocalAddressIP = a.Dedup()
	}

	b, err := bnet.IPFromString(bn.PeerAddress)
	if err != nil {
		return fmt.Errorf("unable to parse BGP peer address: %w", err)
	}

	bn.PeerAddressIP = b.Dedup()
	bn.HoldTimeDuration = time.Second * time.Duration(bn.HoldTime)

	for i := range bn.Import {
		f := policyOptions.getPolicyStatementFilter(bn.Import[i])
		if f == nil {
			return fmt.Errorf("policy statement %q undefined", bn.Import[i])
		}

		bn.ImportFilterChain = append(bn.ImportFilterChain, f)
	}

	for i := range bn.Export {
		f := policyOptions.getPolicyStatementFilter(bn.Export[i])
		if f == nil {
			return fmt.Errorf("policy statement %q undefined", bn.Export[i])
		}

		bn.ExportFilterChain = append(bn.ExportFilterChain, f)
	}
	return nil
}

type AddressFamilyConfig struct {
	AddPath *AddPathConfig
}

type AddPathConfig struct {
	Receive bool               `yaml:"receive"`
	Send    *AddPathSendConfig `yaml:"send"`
}

type AddPathSendConfig struct {
	Multipath bool  `yaml:"multipath"`
	PathCount uint8 `yaml:"path_count"`
}
