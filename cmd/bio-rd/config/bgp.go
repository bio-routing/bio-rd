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
	// description: |
	//   List of BGP Peer groups. All peers *must* belong to a group
	//   If a parameter is configured in the group and in the neighbor level, the neighbor is used
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
	// description: |
	//   Name for the group
	Name string `yaml:"name"`
	// description: |
	//   Local address for the peering
	LocalAddress string `yaml:"local_address"`
	// docgen:nodoc
	LocalAddressIP *bnet.IP
	// description: |
	//   Maximum allowed TTL for routes from peers belonging to this group
	TTL uint8 `yaml:"ttl"`
	// description: |
	//   MD5 secret for the session
	AuthenticationKey string `yaml:"authentication_key"`
	// description: |
	//   Peer AS number
	PeerAS uint32 `yaml:"peer_as"`
	// description: |
	//   Local AS number
	LocalAS uint32 `yaml:"local_as"`
	// description: |
	//   Hold timer
	HoldTime uint16 `yaml:"hold_time"`
	// description: |
	//   Enables multipath routes
	Multipath *Multipath `yaml:"multipath"`
	// description: |
	//   List of import filters.
	//   Example:
	//   import: ["ACCEPT_ALL"]
	//   # this example assumes that the a policy named ACCEPT_ALL exists in the configuration
	Import []string `yaml:"import"`
	// docgen:nodoc
	ImportFilterChain filter.Chain
	// description: |
	//   List of export filters. Syntax is the same as with import
	Export []string `yaml:"export"`
	// docgen:nodoc
	ExportFilterChain filter.Chain
	// description: |
	//   Configures the daemon as a route server client
	RouteServerClient *bool `yaml:"route_server_client"`
	// description: |
	//   Configures the daemon as a route reflector client
	RouteReflectorClient *bool `yaml:"route_reflector_client"`
	// description: |
	//   Cluster ID for route reflection
	ClusterID string `yaml:"cluster_id"`
	// docgen:nodoc
	ClusterIDIP *bnet.IP
	// description: |
	//   Configures the client in passive mode
	Passive *bool `yaml:"passive"`
	// description: |
	//   Neighbors that belong to this group. See bgpneighbors.md for details.
	Neighbors []*BGPNeighbor `yaml:"neighbors"`
	// description: |
	//   Configuration values for the IPv4 AFI family
	IPv4 *AddressFamilyConfig `yaml:"ipv4"`
	// description: |
	//   Configuration values for the IPv6 AFI family
	IPv6 *AddressFamilyConfig `yaml:"ipv6"`
	// description: |
	//   Name of the routing instance this groups belongs to
	RoutingInstance string `yaml:"routing_instance"`
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
	// description: |
	//   Enable multipath
	Enable bool `yaml:"enable"`
	// description: |
	//   Enable learning multiple paths for routes coming from different AS
	MulipleAS bool `yaml:"multiple_as"`
}

type BGPNeighbor struct {
	// description: |
	//   Address for the peer
	PeerAddress string `yaml:"peer_address"`
	// docgen:nodoc
	PeerAddressIP *bnet.IP
	// description: |
	//   Local address for the session
	LocalAddress string `yaml:"local_address"`
	// docgen:nodoc
	LocalAddressIP *bnet.IP
	// description: |
	//   Disable the session with this peer
	Disabled bool `yaml:"disabled"`
	// description: |
	//   Maximum allowed TTL for routes from peers belonging to this group
	TTL uint8 `yaml:"ttl"`
	// description: |
	//   MD5 secret for the session
	AuthenticationKey string `yaml:"authentication_key"`
	// description: |
	//   Peer AS number
	PeerAS uint32 `yaml:"peer_as"`
	// description: |
	//   Local AS number
	LocalAS uint32 `yaml:"local_as"`
	// description: |
	//   Hold timer
	HoldTime uint16 `yaml:"hold_time"`
	// docgen:nodoc
	HoldTimeDuration time.Duration
	// description: |
	//   Enables multipath routes
	//   Allowed values:
	//   - enable: enables the feature
	//   - multiple_as: enables learning multiple paths for routes coming from different AS
	Multipath *Multipath `yaml:"multipath"`
	// description: |
	//   List of import filters.
	//   Example:
	//   import: ["ACCEPT_ALL"]
	//   # this example assumes that the a policy named ACCEPT_ALL exists in the configuration
	Import []string `yaml:"import"`
	// docgen:nodoc
	ImportFilterChain filter.Chain
	// description: |
	//   List of export filters. Syntax is the same as with import
	Export []string `yaml:"export"`
	// docgen:nodoc
	ExportFilterChain filter.Chain
	// description: |
	//   Configures the daemon as a route server client
	RouteServerClient *bool `yaml:"route_server_client"`
	// description: |
	//   Configures the daemon as a route reflector client
	RouteReflectorClient *bool `yaml:"route_reflector_client"`
	// description: |
	//   Configures the client in passive mode
	Passive *bool `yaml:"passive"`
	// description: |
	//   Cluster ID for route reflection
	ClusterID string `yaml:"cluster_id"`
	// docgen:nodoc
	ClusterIDIP *bnet.IP
	// description: |
	//   Configuration values for the IPv4 AFI family
	IPv4 *AddressFamilyConfig `yaml:"ipv4"`
	// description: |
	//   Configuration values for the IPv6 AFI family
	IPv6 *AddressFamilyConfig `yaml:"ipv6"`
	// description: |
	//   Advertise the multiprotocol capability for the IPv4 AFI
	AdvertiseIPv4MultiProtocol bool `yaml:"advertise_ipv4_multiprotocol"`
	// description: |
	//   Name of the routing instance this groups belongs to
	RoutingInstance string `yaml:"routing_instance"`
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
	// description: |
	//   Enable add_path for send and receive
	AddPath *AddPathConfig `yaml:"add_path"`
	// description: |
	//   Enable extended next hop for the address family
	NextHopExtended bool `yaml:"next_hop_extended"`
}

type AddPathConfig struct {
	// description: |
	//   Enable receive add_path
	Receive bool `yaml:"receive"`
	// description: |
	//   Enable send add_path
	Send *AddPathSendConfig `yaml:"send"`
}

type AddPathSendConfig struct {
	// description: |
	//   Enable multipath by add_path
	Multipath bool `yaml:"multipath"`
	// description: |
	//   Maximum allowed path count for add_path
	PathCount uint8 `yaml:"path_count"`
}
