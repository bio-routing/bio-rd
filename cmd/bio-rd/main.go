package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/protocols/kernel"
	"github.com/bio-routing/bio-rd/route"

	"github.com/bio-routing/bio-rd/cmd/bio-rd/config"
	bnet "github.com/bio-routing/bio-rd/net"
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
	configFilePath       *string
	sigCh                = make(chan os.Signal, 1)
	grpcPort             = flag.Uint("grpc_port", 5566, "GRPC API server port")
	grpcKeepaliveMinTime = flag.Uint("grpc_keepalive_min_time", 1, "Minimum time (seconds) for a client to wait between GRPC keepalive pings")
	metricsPort          = flag.Uint("metrics_port", 55667, "Metrics HTTP server port")
	vrfReg               = vrf.NewVRFRegistry()
	bgpSrv               bgpserver.BGPServer
)

func init() {
	hostname, _ := os.Hostname()
	if hostname == "A" {
		configFilePath = flag.String("config.file", "bio-rd-A.yml", "bio-rd config file")
	} else if hostname == "B" {
		configFilePath = flag.String("config.file", "bio-rd-B.yml", "bio-rd config file")
	} else {
		configFilePath = flag.String("config.file", "bio-rd.yml", "bio-rd config file")
	}
}

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

	go signalChecker()
	sigCh <- syscall.SIGHUP
	installSignalHandler()

	rib, _ := vrfReg.GetVRFByName("master").RIBByName("inet.0")
	k, err := kernel.New()
	if err != nil {
		log.Errorf("Unable to create protocol kernel: %v", err)
		os.Exit(1)
	}
	defer k.Dispose()

	rib.Register(k)

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
	signal.Notify(sigCh, syscall.SIGHUP)
	signal.Notify(sigCh, syscall.SIGINT)
}

func signalChecker() {
	for sig := range sigCh {
		if sig != syscall.SIGHUP {
			log.Infof("Received signal to STOP")
			// TO DO: send grpc message to the peer to shut it down
			os.Exit(1)
		}

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

	go configureRoutingOptions(cfg.RoutingOptions)

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

func createBGPPath(bgpConf *bgpserver.PeerConfig) *route.Path {
	return &route.Path{
		Type: route.BGPPathType,
		BGPPath: &route.BGPPath{
			BGPPathA: &route.BGPPathA{
				NextHop: bgpConf.PeerAddress,
				Source:  bgpConf.LocalAddress,
			},
			ASPath: &types.ASPath{
				types.ASPathSegment{
					Type: types.ASSequence,
					ASNs: []uint32{bgpConf.LocalAS, bgpConf.PeerAS},
				},
			},
		},
	}
}

func configureRoutingOptions(ro *config.RoutingOptions) {
	rib, _ := vrfReg.GetVRFByName("master").RIBByName("inet.0")
	var paths []*route.Path
	addressPrefixPairs := make(map[string]string)
	var err error

	ticker := time.NewTicker(15 * time.Second)
	for range ticker.C {
		for _, p := range bgpSrv.GetPeers() {
			bgpConf := bgpSrv.GetPeerConfig(p)
			paths = append(paths, createBGPPath(bgpConf))
		}

		for _, sr := range ro.StaticRoutes {
			addressPrefix := strings.Split(sr.Prefix, "/")
			addressPrefixPairs[addressPrefix[0]] = addressPrefix[1]
		}

		for _, path := range paths {
			for a, p := range addressPrefixPairs {

				var address bnet.IP
				address, err = bnet.IPFromString(a)
				if err != nil {
					log.Errorf("error getting IP from a string", err)
				}
				var pref uint64
				pref, err = strconv.ParseUint(p, 10, 8)
				if err != nil {
					log.Errorf("error parsing prefix string to uint", err)
				}

				prefix := bnet.NewPfx(address, uint8(pref))
				if !rib.ContainsPfxPath(&prefix, path) {
					err = rib.AddPath(&prefix, path)
					if err != nil {
						log.Errorf("error adding path to the rib", err)
					}
				}
			}

			in := bgpSrv.GetRIBIn(path.NextHop(), 1, 1)
			if in == nil {
				fmt.Println("RIB-In is nil, continuing")
				continue
			}

			routes := in.Dump()
			for _, r := range routes {
				if !rib.ContainsPfxPath(r.Prefix(), path) {
					err = rib.AddPath(r.Prefix(), path)
					if err != nil {
						log.Errorf("error adding path to the rib", err)
					}
				}
			}
		}
	}
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
