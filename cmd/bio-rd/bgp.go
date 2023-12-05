package main

import (
	"fmt"
	"time"

	"github.com/bio-routing/bio-rd/cmd/bio-rd/config"
	bgpserver "github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/routingtable"
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
	for _, g := range cfg.Groups {
		for _, n := range g.Neighbors {
			newCfg := c.toPeerConfig(n, c.srv.GetDefaultVRF())
			oldCfg := c.srv.GetPeerConfig(c.srv.GetDefaultVRF(), n.PeerAddressIP)
			if oldCfg == nil {
				continue
			}

			if !oldCfg.NeedsRestart(newCfg) {
				c.srv.ReplaceImportFilterChain(c.srv.GetDefaultVRF(), n.PeerAddressIP, newCfg.IPv4.ImportFilterChain)
				c.srv.ReplaceExportFilterChain(c.srv.GetDefaultVRF(), n.PeerAddressIP, newCfg.IPv4.ExportFilterChain)
				continue
			}

			c.srv.DisposePeer(c.srv.GetDefaultVRF(), oldCfg.PeerAddress)
			return c.srv.AddPeer(*newCfg)
		}

	}

	return nil
}

func (c *bgpConfigurator) configureNewSessions(cfg *config.BGP) error {
	for _, g := range cfg.Groups {
		for _, n := range g.Neighbors {
			if c.srv.GetPeerConfig(c.srv.GetDefaultVRF(), n.PeerAddressIP) != nil {
				continue
			}

			newCfg := c.toPeerConfig(n, vrfReg.GetVRFByName(vrf.DefaultVRFName))
			err := c.srv.AddPeer(*newCfg)
			if err != nil {
				return fmt.Errorf("unable to add BGP peer: %w", err)
			}
		}
	}

	return nil
}

func (c *bgpConfigurator) toPeerConfig(n *config.BGPNeighbor, vrf *vrf.VRF) *bgpserver.PeerConfig {
	r := &bgpserver.PeerConfig{
		AdminEnabled:      !n.Disabled,
		AuthenticationKey: n.AuthenticationKey,
		LocalAS:           n.LocalAS,
		PeerAS:            n.PeerAS,
		PeerAddress:       n.PeerAddressIP,
		LocalAddress:      n.LocalAddressIP,
		TTL:               n.TTL,
		ReconnectInterval: time.Second * 15,
		HoldTime:          n.HoldTimeDuration,
		KeepAlive:         n.HoldTimeDuration / 3,
		RouterID:          c.srv.RouterID(),
		IPv4: &bgpserver.AddressFamilyConfig{
			ImportFilterChain: n.ImportFilterChain,
			ExportFilterChain: n.ExportFilterChain,
			AddPathSend: routingtable.ClientOptions{
				MaxPaths: 10,
			},
		},
		VRF: vrf,
	}

	// TODO: configureAFIsForBGPPeer(n, r)

	if n.Passive != nil {
		r.Passive = *n.Passive
	}

	if n.RouteServerClient != nil {
		r.RouteServerClient = *n.RouteServerClient
	}

	return r
}
