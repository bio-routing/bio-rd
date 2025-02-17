package main

import (
	"fmt"
	"time"

	"github.com/bio-routing/bio-rd/cmd/bio-rd/config"
	bgpserver "github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
)

const (
	DefaultReconnectInterval = time.Second * 15
)

type bgpConfigurator struct {
	srv    bgpserver.BGPServer
	vrfReg *vrf.VRFRegistry
}

func (c *bgpConfigurator) configure(cfg *config.BGP) error {
	for _, bg := range cfg.Groups {
		for _, bn := range bg.Neighbors {
			err := c.configureSession(bn, bg)
			if err != nil {
				return fmt.Errorf("could not configure session for neighbor %s: %w", bn.PeerAddress, err)
			}
		}
	}

	c.deconfigureRemovedSessions(cfg)

	return nil
}

func (c *bgpConfigurator) configureSession(bn *config.BGPNeighbor, bg *config.BGPGroup) error {
	v, err := c.determineVRF(bn, bg)
	if err != nil {
		return fmt.Errorf("could not determine VRF: %w", err)
	}

	newCfg := c.newPeerConfig(bn, bg, v)
	oldCfg := c.srv.GetPeerConfig(v, bn.PeerAddressIP)
	if oldCfg != nil {
		return c.reconfigureModifiedSession(bn, bg, newCfg, oldCfg)
	}

	err = c.srv.AddPeer(*newCfg)
	if err != nil {
		return fmt.Errorf("unable to add BGP peer: %w", err)
	}

	return nil
}

func (c *bgpConfigurator) deconfigureRemovedSessions(cfg *config.BGP) {
	for _, p := range c.srv.GetPeers() {
		if !c.peerExistsInConfig(cfg, p) {
			c.srv.DisposePeer(p.VRF(), p.Addr())
		}
	}
}

func (c *bgpConfigurator) peerExistsInConfig(cfg *config.BGP, p bgpserver.PeerKey) bool {
	for _, bg := range cfg.Groups {
		for _, bn := range bg.Neighbors {
			v, _ := c.determineVRF(bn, bg)
			if bn.PeerAddressIP == p.Addr() && p.VRF() == v {
				return true
			}
		}
	}

	return false
}

func (c *bgpConfigurator) reconfigureModifiedSession(bn *config.BGPNeighbor, bg *config.BGPGroup, newCfg, oldCfg *bgpserver.PeerConfig) error {
	if oldCfg.NeedsRestart(newCfg) {
		return c.replaceSession(newCfg, oldCfg)
	}

	err := c.srv.ReplaceImportFilterChain(
		newCfg.VRF,
		bn.PeerAddressIP,
		bn.ImportFilterChain)
	if err != nil {
		return fmt.Errorf("could not replace import filter: %w", err)
	}

	err = c.srv.ReplaceExportFilterChain(
		newCfg.VRF,
		bn.PeerAddressIP,
		bn.ExportFilterChain)
	if err != nil {
		return fmt.Errorf("could not replace export filter: %w", err)
	}

	return nil
}

func (c *bgpConfigurator) replaceSession(newCfg, oldCfg *bgpserver.PeerConfig) error {
	c.srv.DisposePeer(oldCfg.VRF, oldCfg.PeerAddress)
	err := c.srv.AddPeer(*newCfg)
	if err != nil {
		return fmt.Errorf("unable to reconfigure BGP peer: %w", err)
	}

	return nil
}

func (c *bgpConfigurator) determineVRF(bn *config.BGPNeighbor, bg *config.BGPGroup) (*vrf.VRF, error) {
	if len(bn.RoutingInstance) > 0 {
		return c.vrfByName(bn.RoutingInstance)
	}

	if len(bg.RoutingInstance) > 0 {
		return c.vrfByName(bg.RoutingInstance)
	}

	return bgpSrv.GetDefaultVRF(), nil
}

func (c *bgpConfigurator) vrfByName(name string) (*vrf.VRF, error) {
	v := c.vrfReg.GetVRFByName(name)
	if v == nil {
		return nil, fmt.Errorf("could not find VRF for name %s", name)
	}

	return v, nil
}

func (c *bgpConfigurator) newPeerConfig(bn *config.BGPNeighbor, bg *config.BGPGroup, vrf *vrf.VRF) *bgpserver.PeerConfig {
	p := &bgpserver.PeerConfig{
		AdminEnabled:               !bn.Disabled,
		AuthenticationKey:          bn.AuthenticationKey,
		LocalAS:                    bn.LocalAS,
		PeerAS:                     bn.PeerAS,
		PeerAddress:                bn.PeerAddressIP,
		LocalAddress:               bn.LocalAddressIP,
		TTL:                        bn.TTL,
		ReconnectInterval:          DefaultReconnectInterval,
		HoldTime:                   bn.HoldTimeDuration,
		KeepAlive:                  bn.HoldTimeDuration / 3,
		RouterID:                   c.srv.RouterID(),
		VRF:                        vrf,
		AdvertiseIPv4MultiProtocol: bn.AdvertiseIPv4MultiProtocol,
	}

	c.configureIPv4(bn, bg, p)
	c.configureIPv6(bn, bg, p)

	if bn.Passive != nil {
		p.Passive = *bn.Passive
	}

	if bn.RouteServerClient != nil {
		p.RouteServerClient = *bn.RouteServerClient
	}

	if bn.RouteReflectorClient != nil {
		p.RouteReflectorClient = *bn.RouteReflectorClient
	}

	if bn.ClusterIDIP != nil {
		p.RouteReflectorClusterID = bn.ClusterIDIP.ToUint32()
	}

	return p
}

func (c *bgpConfigurator) configureIPv4(bn *config.BGPNeighbor, bg *config.BGPGroup, p *bgpserver.PeerConfig) {
	if !bn.PeerAddressIP.IsIPv4() && bn.IPv4 == nil {
		return
	}

	p.IPv4 = c.newAFIConfig(bn, bg)

	if bn.IPv4 != nil {
		c.configureAddressFamily(bn.IPv4, p.IPv4)
	}
}

func (c *bgpConfigurator) configureIPv6(bn *config.BGPNeighbor, bg *config.BGPGroup, p *bgpserver.PeerConfig) {
	if bn.PeerAddressIP.IsIPv4() && bn.IPv6 == nil {
		return
	}

	p.IPv6 = c.newAFIConfig(bn, bg)

	if bn.IPv6 != nil {
		c.configureAddressFamily(bn.IPv6, p.IPv6)
	}
}

func (c *bgpConfigurator) newAFIConfig(bn *config.BGPNeighbor, bg *config.BGPGroup) *bgpserver.AddressFamilyConfig {
	return &bgpserver.AddressFamilyConfig{
		ImportFilterChain: bn.ImportFilterChain,
		ExportFilterChain: bn.ExportFilterChain,
		AddPathSend: routingtable.ClientOptions{
			BestOnly: true,
		},
		AddPathRecv: false,
	}
}

func (c *bgpConfigurator) configureAddressFamily(baf *config.AddressFamilyConfig, af *bgpserver.AddressFamilyConfig) {
	if baf.AddPath != nil {
		c.configureAddPath(baf.AddPath, af)
	}

	if baf.NextHopExtended {
		af.NextHopExtended = true
	}
}

func (c *bgpConfigurator) configureAddPath(bac *config.AddPathConfig, af *bgpserver.AddressFamilyConfig) {
	af.AddPathRecv = bac.Receive

	if bac.Send == nil {
		return
	}

	af.AddPathSend.BestOnly = !bac.Send.Multipath
	af.AddPathSend.MaxPaths = uint(bac.Send.PathCount)
}
