package config

import (
	"fmt"
	"time"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/routingtable/filter"
)

const (
	DefaultHoldTimeSeconds = 90
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
	Name                 string `yaml:"name"`
	LocalAddress         string `yaml:"local_address"`
	LocalAddressIP       *bnet.IP
	TTL                  uint8      `yaml:"ttl"`
	AuthenticationKey    string     `yaml:"authentication_key"`
	PeerAS               uint32     `yaml:"peer_as"`
	LocalAS              uint32     `yaml:"local_as"`
	HoldTime             uint16     `yaml:"hold_time"`
	Multipath            *Multipath `yaml:"multipath"`
	Import               []string   `yaml:"import"`
	ImportFilterChain    filter.Chain
	Export               []string `yaml:"export"`
	ExportFilterChain    filter.Chain
	RouteServerClient    *bool  `yaml:"route_server_client"`
	RouteReflectorClient *bool  `yaml:"route_reflector_client"`
	ClusterID            string `yaml:"cluster_id"`
	ClusterIDIP          *bnet.IP
	Passive              *bool                `yaml:"passive"`
	Neighbors            []*BGPNeighbor       `yaml:"neighbors"`
	IPv4                 *AddressFamilyConfig `yaml:"ipv4"`
	IPv6                 *AddressFamilyConfig `yaml:"ipv6"`
	RoutingInstance      string               `yaml:"routing_instance"`
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

	if bg.ClusterID != "" {
		a, err := bnet.IPFromString(bg.ClusterID)
		if err != nil {
			return fmt.Errorf("unable to parse BGP cluster identifier: %w", err)
		}

		bg.ClusterIDIP = a.Dedup()
	}

	if bg.HoldTime == 0 {
		bg.HoldTime = DefaultHoldTimeSeconds
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
			bn.RouteServerClient = bg.RouteServerClient
		}

		if bn.Passive == nil {
			bn.Passive = bg.Passive
		}

		if bn.LocalAddress == "" {
			bn.LocalAddressIP = bg.LocalAddressIP
		}

		if bn.ClusterID == "" {
			bn.ClusterIDIP = bg.ClusterIDIP
		}

		if bn.TTL == 0 {
			bn.TTL = bg.TTL
		}

		if bn.AuthenticationKey == "" {
			bn.AuthenticationKey = bg.AuthenticationKey
		}

		if bn.LocalAS == 0 {
			bn.LocalAS = bg.LocalAS
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

		if bn.Multipath == nil {
			bn.Multipath = bg.Multipath
		}

		if len(bn.RoutingInstance) == 0 {
			bn.RoutingInstance = bg.RoutingInstance
		}

		if bn.IPv4 == nil {
			bn.IPv4 = bg.IPv4
		}

		if bn.IPv6 == nil {
			bn.IPv6 = bg.IPv6
		}

		if bn.RouteReflectorClient == nil {
			bn.RouteReflectorClient = bg.RouteReflectorClient
		}

		bn.ImportFilterChain = bg.ImportFilterChain
		bn.ExportFilterChain = bg.ExportFilterChain

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
	RouteReflectorClient       *bool  `yaml:"route_reflector_client"`
	Passive                    *bool  `yaml:"passive"`
	ClusterID                  string `yaml:"cluster_id"`
	ClusterIDIP                *bnet.IP
	IPv4                       *AddressFamilyConfig `yaml:"ipv4"`
	IPv6                       *AddressFamilyConfig `yaml:"ipv6"`
	AdvertiseIPv4MultiProtocol bool                 `yaml:"advertise_ipv4_multiprotocol"`
	RoutingInstance            string               `yaml:"routing_instance"`
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

	if bn.ClusterID != "" {
		a, err := bnet.IPFromString(bn.ClusterID)
		if err != nil {
			return fmt.Errorf("unable to parse BGP cluster identifier: %w", err)
		}

		bn.ClusterIDIP = a.Dedup()
	}

	b, err := bnet.IPFromString(bn.PeerAddress)
	if err != nil {
		return fmt.Errorf("unable to parse BGP peer address: %w", err)
	}

	bn.PeerAddressIP = b.Dedup()
	bn.HoldTimeDuration = time.Second * time.Duration(bn.HoldTime)

	if len(bn.Import) > 0 {
		bn.ImportFilterChain = filter.Chain{}
	}
	for i := range bn.Import {
		f := policyOptions.getPolicyStatementFilter(bn.Import[i])
		if f == nil {
			return fmt.Errorf("policy statement %q undefined", bn.Import[i])
		}

		bn.ImportFilterChain = append(bn.ImportFilterChain, f)
	}

	if len(bn.Export) > 0 {
		bn.ExportFilterChain = filter.Chain{}
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
	AddPath         *AddPathConfig `yaml:"add_path"`
	NextHopExtended bool           `yaml:"next_hop_extended"`
}

type AddPathConfig struct {
	Receive bool               `yaml:"receive"`
	Send    *AddPathSendConfig `yaml:"send"`
}

type AddPathSendConfig struct {
	Multipath bool  `yaml:"multipath"`
	PathCount uint8 `yaml:"path_count"`
}
