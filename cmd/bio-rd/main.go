package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bio-routing/bio-rd/cmd/bio-rd/config"
	bgpapi "github.com/bio-routing/bio-rd/protocols/bgp/api"
	bgpserver "github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
	"github.com/bio-routing/bio-rd/util/servicewrapper"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var (
	configFilePath       = flag.String("config.file", "bio-rd.yml", "bio-rd config file")
	grpcPort             = flag.Uint("grpc_port", 5566, "GRPC API server port")
	grpcKeepaliveMinTime = flag.Uint("grpc_keepalive_min_time", 1, "Minimum time (seconds) for a client to wait between GRPC keepalive pings")
	metricsPort          = flag.Uint("metrics_port", 55667, "Metrics HTTP server port")
	sigHUP               = make(chan os.Signal)
	vrfReg               = vrf.NewVRFRegistry()
	bgpSrv               bgpserver.BGPServer
	runCfg               *config.Config
)

func main() {
	flag.Parse()

	startCfg, err := config.GetConfig(*configFilePath)
	if err != nil {
		log.Errorf("Unable to get config: %v", err)
		os.Exit(1)
	}

	bgpSrv = bgpserver.NewBGPServer(
		startCfg.RoutingOptions.RouterIDUint32,
		[]string{
			":179",
		},
	)

	err = bgpSrv.Start()
	if err != nil {
		log.Fatalf("Unable to start BGP server: %v", err)
		os.Exit(1)
	}

	vrfReg.CreateVRFIfNotExists("master", 0)

	go configReloader()
	sigHUP <- syscall.SIGHUP
	installSignalHandler()

	s := bgpserver.NewBGPAPIServer(bgpSrv)
	unaryInterceptors := []grpc.UnaryServerInterceptor{}
	streamInterceptors := []grpc.StreamServerInterceptor{}
	srv, err := servicewrapper.New(
		uint16(*grpcPort),
		servicewrapper.HTTP(uint16(*metricsPort)),
		unaryInterceptors,
		streamInterceptors,
		keepalive.EnforcementPolicy{
			MinTime:             time.Duration(*grpcKeepaliveMinTime) * time.Second,
			PermitWithoutStream: true,
		},
	)
	if err != nil {
		log.Errorf("failed to listen: %v", err)
		os.Exit(1)
	}

	bgpapi.RegisterBgpServiceServer(srv.GRPC(), s)
	if err := srv.Serve(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

	select {}
}

func installSignalHandler() {
	signal.Notify(sigHUP, syscall.SIGHUP)
}

func configReloader() {
	for {
		<-sigHUP
		log.Infof("Reloading configuration")
		newCfg, err := config.GetConfig(*configFilePath)
		if err != nil {
			log.Errorf("Failed to get config: %v", err)
			continue
		}

		err = loadConfig(newCfg)
		if err != nil {
			log.Errorf("Unable to load config: %v", err)
			continue
		}

		log.Infof("Configuration reloaded")
	}
}

func loadConfig(cfg *config.Config) error {

	for _, ri := range cfg.RoutingInstances {
		err := configureRoutingInstance(ri)
		_ = err

	}

	if cfg.Protocols != nil {
		if cfg.Protocols.BGP != nil {
			err := configureProtocolsBGP(cfg.Protocols.BGP)
			if err != nil {
				return errors.Wrap(err, "Unable to configure BGP")
			}
		}
	}

	return nil
}

func configureProtocolsBGP(bgp *config.BGP) error {
	// Tear down peers that are to be removed
	for _, p := range bgpSrv.GetPeers() {
		found := false
		for _, g := range bgp.Groups {
			for _, n := range g.Neighbors {
				if n.PeerAddressIP == p {
					found = true
					break
				}
			}
		}

		if !found {
			bgpSrv.DisposePeer(p)
		}
	}

	// Tear down peers that need new sessions as they changed too significantly
	for _, g := range bgp.Groups {
		for _, n := range g.Neighbors {
			newCfg := BGPPeerConfig(n, vrfReg.GetVRFByRD(0))
			oldCfg := bgpSrv.GetPeerConfig(n.PeerAddressIP)
			if oldCfg == nil {
				continue
			}

			if !oldCfg.NeedsRestart(newCfg) {
				bgpSrv.ReplaceImportFilterChain(n.PeerAddressIP, newCfg.IPv4.ImportFilterChain)
				bgpSrv.ReplaceExportFilterChain(n.PeerAddressIP, newCfg.IPv4.ExportFilterChain)
				continue
			}

			bgpSrv.DisposePeer(oldCfg.PeerAddress)
		}
	}

	// Turn up all sessions that are missing
	for _, g := range bgp.Groups {
		for _, n := range g.Neighbors {
			if bgpSrv.GetPeerConfig(n.PeerAddressIP) != nil {
				continue
			}

			newCfg := BGPPeerConfig(n, vrfReg.GetVRFByRD(0))
			err := bgpSrv.AddPeer(*newCfg)
			if err != nil {
				return errors.Wrap(err, "Unable to add BGP peer")
			}
		}
	}

	return nil
}

func BGPPeerConfig(n *config.BGPNeighbor, vrf *vrf.VRF) *bgpserver.PeerConfig {
	r := &bgpserver.PeerConfig{
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

	if n.Passive != nil {
		r.Passive = *n.Passive
	}

	if n.RouteServerClient != nil {
		r.RouteServerClient = *n.RouteServerClient
	}

	return r
}

func configureRoutingInstance(ri *config.RoutingInstance) error {
	vrf := vrfReg.GetVRFByName(ri.Name)

	// RD Change
	if vrf.RD() != ri.InternalRouteDistinguisher {
		// TODO: Drop all routing adjacencies
		vrf.Dispose()
		vrfReg.UnregisterVRF(vrf)

		vrf = vrfReg.CreateVRFIfNotExists(ri.Name, ri.InternalRouteDistinguisher)
		// TODO: Add all routing adjacencies
	}

	return nil
}
