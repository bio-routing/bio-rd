package main

import (
	"fmt"
	"time"

	"github.com/bio-routing/bio-rd/cmd/bio-rd/config"
	bgpserver "github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
)

type bgpConfigurator struct {
	srv bgpserver.BGPServer
}

func (c *bgpConfigurator) configure(cfg *config.BGP) error {
	c.deconfigureRemovedSessions(cfg)

	err := c.reconfigureModifiedSessions(cfg)
	if err != nil {
		return fmt.Errorf("could not reconfigure session: %w", err)
	}

	err = c.configureNewSessions(cfg)
	if err != nil {
		return fmt.Errorf("could not configure session: %w", err)
	}

	return nil
}

func (c *bgpConfigurator) deconfigureRemovedSessions(cfg *config.BGP) {
	for _, p := range c.srv.GetPeers() {
		found := false
		for _, g := range cfg.Groups {
			for _, n := range g.Neighbors {
				if n.PeerAddressIP == p.Addr() && p.VRF() == c.srv.GetDefaultVRF() {
					found = true
					break
				}
			}
		}

		if !found {
			c.srv.DisposePeer(c.srv.GetDefaultVRF(), p.Addr())
		}
	}
}

func (c *bgpConfigurator) reconfigureModifiedSessions(cfg *config.BGP) error {
	for _, bg := range cfg.Groups {
		for _, bn := range bg.Neighbors {
			newCfg := c.newPeerConfig(bn, bg, c.srv.GetDefaultVRF())
			oldCfg := c.srv.GetPeerConfig(c.srv.GetDefaultVRF(), bn.PeerAddressIP)
			if oldCfg == nil {
				continue
			}

			if !oldCfg.NeedsRestart(newCfg) {
				c.srv.ReplaceImportFilterChain(c.srv.GetDefaultVRF(),
					bn.PeerAddressIP,
					c.determineFilterChain(bg.ImportFilterChain, bn.ImportFilterChain))
				c.srv.ReplaceExportFilterChain(c.srv.GetDefaultVRF(),
					bn.PeerAddressIP,
					c.determineFilterChain(bg.ExportFilterChain, bn.ExportFilterChain))
				continue
			}

			c.srv.DisposePeer(c.srv.GetDefaultVRF(), oldCfg.PeerAddress)
			return c.srv.AddPeer(*newCfg)
		}
	}

	return nil
}

func (c *bgpConfigurator) configureNewSessions(cfg *config.BGP) error {
	for _, bg := range cfg.Groups {
		for _, bn := range bg.Neighbors {
			if c.srv.GetPeerConfig(c.srv.GetDefaultVRF(), bn.PeerAddressIP) != nil {
				continue
			}

			newCfg := c.newPeerConfig(bn, bg, vrfReg.GetVRFByName(vrf.DefaultVRFName))
			err := c.srv.AddPeer(*newCfg)
			if err != nil {
				return fmt.Errorf("unable to add BGP peer: %w", err)
			}
		}
	}

	return nil
}

func (c *bgpConfigurator) newPeerConfig(bn *config.BGPNeighbor, bg *config.BGPGroup, vrf *vrf.VRF) *bgpserver.PeerConfig {
	p := &bgpserver.PeerConfig{
		AdminEnabled:      !bn.Disabled,
		AuthenticationKey: bn.AuthenticationKey,
		LocalAS:           bn.LocalAS,
		PeerAS:            bn.PeerAS,
		PeerAddress:       bn.PeerAddressIP,
		LocalAddress:      bn.LocalAddressIP,
		TTL:               bn.TTL,
		ReconnectInterval: time.Second * 15,
		HoldTime:          bn.HoldTimeDuration,
		KeepAlive:         bn.HoldTimeDuration / 3,
		RouterID:          c.srv.RouterID(),
		VRF:               vrf,
	}

	c.configureIPv4(bn, bg, p)
	c.configureIPv6(bn, bg, p)

	if bn.Passive != nil {
		p.Passive = *bn.Passive
	}

	if bn.RouteServerClient != nil {
		p.RouteServerClient = *bn.RouteServerClient
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
		ImportFilterChain: c.determineFilterChain(bg.ImportFilterChain, bn.ImportFilterChain),
		ExportFilterChain: c.determineFilterChain(bg.ExportFilterChain, bn.ExportFilterChain),
		AddPathSend: routingtable.ClientOptions{
			BestOnly: true,
		},
		AddPathRecv: false,
	}
}

func (c *bgpConfigurator) determineFilterChain(groupChain filter.Chain, neighChain filter.Chain) filter.Chain {
	if len(neighChain) > 0 {
		return neighChain
	}

	return groupChain
}

func (c *bgpConfigurator) configureAddressFamily(baf *config.AddressFamilyConfig, af *bgpserver.AddressFamilyConfig) {
	if baf.AddPath != nil {
		c.configureAddPath(baf.AddPath, af)
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
