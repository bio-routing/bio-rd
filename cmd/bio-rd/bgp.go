package main

import (
	"fmt"
	"time"

	"github.com/bio-routing/bio-rd/cmd/bio-rd/config"
	bgpserver "github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
)

func configureProtocolsBGP(bgp *config.BGP) error {
	// Tear down peers that are to be removed
	for _, p := range bgpSrv.GetPeers() {
		found := false
		for _, g := range bgp.Groups {
			for _, n := range g.Neighbors {
				if n.PeerAddressIP == p.Addr() && p.VRF() == bgpSrv.GetDefaultVRF() {
					found = true
					break
				}
			}
		}

		if !found {
			bgpSrv.DisposePeer(bgpSrv.GetDefaultVRF(), p.Addr())
		}
	}

	// Tear down peers that need new sessions as they changed too significantly
	for _, g := range bgp.Groups {
		for _, n := range g.Neighbors {
			newCfg := toBGPPeerConfig(n, bgpSrv.GetDefaultVRF())
			oldCfg := bgpSrv.GetPeerConfig(bgpSrv.GetDefaultVRF(), n.PeerAddressIP)
			if oldCfg == nil {
				continue
			}

			if !oldCfg.NeedsRestart(newCfg) {
				bgpSrv.ReplaceImportFilterChain(bgpSrv.GetDefaultVRF(), n.PeerAddressIP, newCfg.IPv4.ImportFilterChain)
				bgpSrv.ReplaceExportFilterChain(bgpSrv.GetDefaultVRF(), n.PeerAddressIP, newCfg.IPv4.ExportFilterChain)
				continue
			}

			bgpSrv.DisposePeer(bgpSrv.GetDefaultVRF(), oldCfg.PeerAddress)
		}
	}

	// Turn up all sessions that are missing
	for _, g := range bgp.Groups {
		for _, n := range g.Neighbors {
			if bgpSrv.GetPeerConfig(bgpSrv.GetDefaultVRF(), n.PeerAddressIP) != nil {
				continue
			}

			newCfg := toBGPPeerConfig(n, vrfReg.GetVRFByName(vrf.DefaultVRFName))
			err := bgpSrv.AddPeer(*newCfg)
			if err != nil {
				return fmt.Errorf("unable to add BGP peer: %w", err)
			}
		}
	}

	return nil
}

// BGPPeerConfig converts a BGPNeighbor config into a PeerConfig
func toBGPPeerConfig(n *config.BGPNeighbor, vrf *vrf.VRF) *bgpserver.PeerConfig {
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
		RouterID:          bgpSrv.RouterID(),
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
