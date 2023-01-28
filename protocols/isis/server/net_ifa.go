package server

import (
	"fmt"
	"sync"
	"time"

	"github.com/bio-routing/bio-rd/net/ethernet"
	"github.com/bio-routing/bio-rd/protocols/device"
	"github.com/bio-routing/bio-rd/util/log"

	bbclock "github.com/benbjohnson/clock"
)

var (
	allISNetworkEntitiesAddr = ethernet.MACAddr{0x09, 0x00, 0x2B, 0x00, 0x00, 0x05}
)

// InterfaceConfig represents a network interfaces ISIS config
type InterfaceConfig struct {
	Name         string
	Passive      bool
	PointToPoint bool
	Level1       *InterfaceLevelConfig
	Level2       *InterfaceLevelConfig
}

// holdingTimer() picks the maximum holding timer from Level1 and Level2 config
func (ifCfg *InterfaceConfig) holdingTimer() uint16 {
	if ifCfg.Level1 != nil && ifCfg.Level2 == nil {
		return ifCfg.Level1.HoldingTimer
	}

	if ifCfg.Level2 != nil && ifCfg.Level1 == nil {
		return ifCfg.Level2.HoldingTimer
	}

	if ifCfg.Level1 == nil && ifCfg.Level2 == nil {
		return 0
	}

	if ifCfg.Level1.HoldingTimer > ifCfg.Level2.HoldingTimer {
		return ifCfg.Level1.HoldingTimer
	}

	return ifCfg.Level2.HoldingTimer
}

func (ifCfg *InterfaceConfig) getMinHelloInterval() uint16 {
	if ifCfg.Level1 == nil && ifCfg.Level2 == nil {
		return 0
	}

	if ifCfg.Level1 == nil {
		return ifCfg.Level2.HelloInterval
	}

	if ifCfg.Level2 == nil {
		return ifCfg.Level1.HelloInterval
	}

	if ifCfg.Level1.HelloInterval < ifCfg.Level2.HelloInterval {
		return ifCfg.Level1.HelloInterval
	}

	return ifCfg.Level2.HelloInterval
}

// InterfaceLevelConfig is the ISIS level config of an interface
type InterfaceLevelConfig struct {
	HelloInterval uint16
	HoldingTimer  uint16
	Metric        uint32
	Passive       bool
	Priority      uint8
}

type netIfaInterface interface {
	start() error
}

type netIfa struct {
	name              string
	srv               *Server
	cfg               *InterfaceConfig
	done              chan struct{}
	wg                sync.WaitGroup
	helloTicker       *bbclock.Ticker
	neighborManagerL1 *neighborManager
	neighborManagerL2 *neighborManager
	mu                sync.RWMutex
	initialized       bool
	devStatus         device.DeviceInterface
	ethernetInterface ethernet.EthernetInterfaceI
}

func newNetIfa(srv *Server, cfg *InterfaceConfig) *netIfa {
	ret := &netIfa{
		name: cfg.Name,
		srv:  srv,
		cfg:  cfg,
		done: make(chan struct{}),
	}

	if cfg.Level1 != nil {
		ret.neighborManagerL1 = newNeighborManager(srv, ret, 1)
	}

	if cfg.Level2 != nil {
		ret.neighborManagerL2 = newNeighborManager(srv, ret, 2)
	}

	ret.helloTicker = clock.Ticker(time.Duration(cfg.getMinHelloInterval()) * time.Second)

	srv.ds.Subscribe(ret, cfg.Name)
	return ret
}

func (nifa *netIfa) getName() string {
	nifa.mu.RLock()
	defer nifa.mu.RUnlock()

	return nifa.name
}

func (nifa *netIfa) isInitialized() bool {
	nifa.mu.Lock()
	defer nifa.mu.Unlock()

	return nifa.initialized
}

// DeviceUpdate receives device up/down events and other (net)device changes
func (nifa *netIfa) DeviceUpdate(dev device.DeviceInterface) {
	nifa.mu.Lock()
	defer nifa.mu.Unlock()

	oldState := uint8(device.IfOperUnknown)
	if nifa.devStatus != nil {
		oldState = nifa.devStatus.GetOperState()
	}

	nifa.devStatus = dev
	if oldState != device.IfOperUp && dev.GetOperState() == device.IfOperUp {
		log.WithFields(nifa.fields()).Info("Interface changed state to operational. Enabling IS-IS")

		err := nifa._start()
		if err != nil {
			log.Errorf("unable to start ISIS on interface %s: %v", nifa.name, err)
		}

		return
	}

	if oldState == device.IfOperUp && dev.GetOperState() != device.IfOperUp {
		log.WithFields(nifa.fields()).Info("Interface changed state to down. Disabling IS-IS")

		nifa._stop()
		return
	}
}

func (nifa *netIfa) fields() log.Fields {
	return log.Fields{
		"protocol":  "IS-IS",
		"component": "Interface",
		"interface": nifa.name,
	}
}

func (nifa *netIfa) _start() error {
	log.WithFields(nifa.fields()).Info("Starting ISIS")

	if nifa.initialized {
		return fmt.Errorf("already running")
	}

	if nifa.cfg.Passive {
		return nil
	}

	ethIfa, err := nifa.srv.ethernetInterfaceFactory.New(nifa.name, getISISBPF(), getISISLLC())
	if err != nil {
		return fmt.Errorf("unable to create ethernet handler (%s): %w", nifa.name, err)
	}
	nifa.ethernetInterface = ethIfa

	nifa.wg.Add(1)
	go nifa.p2pHelloSender()

	err = nifa.ethernetInterface.MCastJoin(allISNetworkEntitiesAddr)
	if err != nil {
		nifa._stop()
		return fmt.Errorf("unable to join IS p2p hello multicast group: %w", err)
	}

	nifa.wg.Add(1)
	go nifa.receiver()

	nifa.initialized = true
	return nil
}

func (nifa *netIfa) stop() {
	nifa.mu.Lock()
	defer nifa.mu.Unlock()

	nifa._stop()
}

func (nifa *netIfa) _stop() {
	if nifa.neighborManagerL1 != nil {
		nifa.neighborManagerL1.netDown()
	}

	if nifa.neighborManagerL2 != nil {
		nifa.neighborManagerL2.netDown()
	}

	close(nifa.done)
	nifa.ethernetInterface.Close()
	nifa.wg.Wait()
}

func getISISLLC() ethernet.LLC {
	return ethernet.LLC{
		DSAP:         0xfe,
		SSAP:         0xfe,
		ControlField: 0x03,
	}
}

func getISISBPF() *ethernet.BPF {
	b := ethernet.NewBPF()
	b.AddTerm(ethernet.BPFTerm{
		Code: 0x28,
		K:    0x0000000e - 14,
	})
	b.AddTerm(ethernet.BPFTerm{
		Code: 0x15,
		Jf:   3,
		K:    0x0000fefe,
	})
	b.AddTerm(ethernet.BPFTerm{
		Code: 0x30,
		K:    0x00000011 - 14,
	})
	b.AddTerm(ethernet.BPFTerm{
		Code: 0x15,
		Jf:   1,
		K:    0x00000083,
	})
	b.AddTerm(ethernet.BPFTerm{
		Code: 0x6,
		K:    0x00040000,
	})
	b.AddTerm(ethernet.BPFTerm{
		Code: 0x6,
	})

	return b
}
