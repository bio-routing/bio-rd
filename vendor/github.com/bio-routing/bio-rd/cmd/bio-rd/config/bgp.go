package config

import (
	"fmt"
	"time"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/pkg/errors"
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
	AuthenticationKey string         `yaml:"authentication_key"`
	PeerAS            uint32         `yaml:"peer_as"`
	LocalAS           uint32         `yaml:"local_as"`
	HoldTime          uint16         `yaml:"hold_time"`
	Multipath         *Multipath     `yaml:"multipath"`
	Import            []string       `yaml:"import"`
	Export            []string       `yaml:"export"`
	RouteServerClient bool           `yaml:"route_server_client"`
	Passive           bool           `yaml:"passive"`
	Neighbors         []*BGPNeighbor `yaml:"neighbors"`
	AFIs              []*AFI         `yaml:"afi"`
}

func (bg *BGPGroup) load(localAS uint32, policyOptions *PolicyOptions) error {
	if bg.LocalAS == 0 {
		bg.LocalAS = localAS
	}

	if bg.LocalAddress != "" {
		a, err := bnet.IPFromString(bg.LocalAddress)
		if err != nil {
			return errors.Wrap(err, "Unable to parse BGP local address")
		}

		bg.LocalAddressIP = a.Dedup()
	}

	if bg.HoldTime == 0 {
		bg.HoldTime = 90
	}

	for _, n := range bg.Neighbors {
		if n.RouteServerClient == nil {
			n.RouteServerClient = &bg.RouteServerClient
		}

		if n.Passive == nil {
			n.Passive = &bg.Passive
		}

		if n.LocalAddress == "" {
			n.LocalAddress = bg.LocalAddress
		}

		if n.AuthenticationKey == "" {
			n.AuthenticationKey = bg.AuthenticationKey
		}

		if n.LocalAS == 0 {
			n.LocalAS = localAS
		}

		if n.LocalAS == 0 {
			return fmt.Errorf("local_as 0 is invalid")
		}

		if n.PeerAS == 0 {
			n.PeerAS = bg.PeerAS
		}

		if n.PeerAS == 0 {
			return fmt.Errorf("peer_as 0 is invalid")
		}

		if n.HoldTime == 0 {
			n.HoldTime = bg.HoldTime
		}

		err := n.load(policyOptions)
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
	PeerAddress       string `yaml:"peer_address"`
	PeerAddressIP     *bnet.IP
	LocalAddress      string `yaml:"local_address"`
	LocalAddressIP    *bnet.IP
	AuthenticationKey string `yaml:"authentication_key"`
	PeerAS            uint32 `yaml:"peer_as"`
	LocalAS           uint32 `yaml:"local_as"`
	HoldTime          uint16 `yaml:"hold_time"`
	HoldTimeDuration  time.Duration
	Multipath         *Multipath `yaml:"multipath"`
	Import            []string   `yaml:"import"`
	ImportFilterChain filter.Chain
	Export            []string `yaml:"export"`
	ExportFilterChain filter.Chain
	RouteServerClient *bool  `yaml:"route_server_client"`
	Passive           *bool  `yaml:"passive"`
	ClusterID         string `yaml:"cluster_id"`
	ClusterIDIP       *bnet.IP
	AFIs              []*AFI `yaml:"afi"`
}

func (bn *BGPNeighbor) load(po *PolicyOptions) error {
	if bn.PeerAS == 0 {
		return fmt.Errorf("Peer %q is lacking peer as number", bn.PeerAddress)
	}

	if bn.PeerAddress == "" {
		return fmt.Errorf("Mandatory parameter BGP peer address is empty")
	}

	a, err := bnet.IPFromString(bn.LocalAddress)
	if err != nil {
		return errors.Wrap(err, "Unable to parse BGP local address")
	}

	bn.LocalAddressIP = a.Dedup()

	b, err := bnet.IPFromString(bn.PeerAddress)
	if err != nil {
		return errors.Wrap(err, "Unable to parse BGP peer address")
	}

	bn.PeerAddressIP = b.Dedup()
	bn.HoldTimeDuration = time.Second * time.Duration(bn.HoldTime)

	for i := range bn.Import {
		f := po.getPolicyStatementFilter(bn.Import[i])
		if f == nil {
			return fmt.Errorf("policy statement %q undefined", bn.Import[i])
		}

		bn.ImportFilterChain = append(bn.ImportFilterChain, f)
	}

	for i := range bn.Export {
		f := po.getPolicyStatementFilter(bn.Export[i])
		if f == nil {
			return fmt.Errorf("policy statement %q undefined", bn.Export[i])
		}

		bn.ExportFilterChain = append(bn.ExportFilterChain, f)
	}
	return nil
}

type AFI struct {
	Name string `yaml:"name"`
	SAFI SAFI   `yaml:"safi"`
}

type SAFI struct {
	Name    string   `yaml:"name"`
	AddPath *AddPath `yaml:"add_path"`
}

type AddPath struct {
	Receive bool         `yaml:"receive"`
	Send    *AddPathSend `yaml:"send"`
}

type AddPathSend struct {
	Multipath bool  `yaml:"multipath"`
	PathCount uint8 `yaml:"path_count"`
}
