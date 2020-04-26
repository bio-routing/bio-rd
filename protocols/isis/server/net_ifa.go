package server

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bio-routing/bio-rd/net/ethernet"
	"github.com/bio-routing/bio-rd/protocols/device"
	btime "github.com/bio-routing/bio-rd/util/time"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// InterfaceConfig represents a network interfaces ISIS config
type InterfaceConfig struct {
	Name         string
	Passive      bool
	PointToPoint bool
	Level1       *InterfaceLevelConfig
	Level2       *InterfaceLevelConfig
	mock         bool
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
	active            uint64
	devStatus         device.DeviceInterface
	done              chan struct{}
	wg                sync.WaitGroup
	helloTicker       btime.Ticker
	ethHandler        ethernet.HandlerInterface
	isP2PHelloCon     net.Conn
	neighborManagerL1 *neighborManager
	neighborManagerL2 *neighborManager
}

func newNetIfa(srv *Server, cfg *InterfaceConfig) *netIfa {
	ret := &netIfa{
		name: cfg.Name,
		srv:  srv,
		cfg:  cfg,
		done: make(chan struct{}),
	}

	if cfg.Level1 != nil {
		ret.neighborManagerL1 = newNeighborManager(ret, 1)
	}

	if cfg.Level2 != nil {
		ret.neighborManagerL2 = newNeighborManager(ret, 2)
	}

	if srv.netIfaManager.useMockTicker {
		ret.helloTicker = btime.NewMockTicker()
	} else {
		ret.helloTicker = btime.NewBIOTicker(time.Duration(cfg.getMinHelloInterval()) * time.Second)
	}

	srv.ds.Subscribe(ret, cfg.Name)
	return ret
}

// DeviceUpdate receives device up/down events and other (net)device changes
func (nifa *netIfa) DeviceUpdate(dev device.DeviceInterface) {
	oldState := uint8(device.IfOperUnknown)
	if nifa.devStatus != nil {
		oldState = nifa.devStatus.GetOperState()
	}

	if oldState != device.IfOperUp && dev.GetOperState() == device.IfOperUp {
		log.WithFields(nifa.fields()).Info("Interface changed state to operational. Enabling IS-IS")
		nifa.devStatus = dev
		err := nifa.start()
		if err != nil {
			log.Errorf("Unable to start ISIS on interface %s: %v", nifa.name, err)
		}
	}
}

func (nifa *netIfa) fields() log.Fields {
	return log.Fields{
		"protocol":  "IS-IS",
		"component": "Interface",
		"interface": nifa.name,
	}
}

func (nifa *netIfa) start() error {
	log.WithFields(nifa.fields()).Info("Starting ISIS")

	if !atomic.CompareAndSwapUint64(&nifa.active, 0, 1) {
		return fmt.Errorf("already active")
	}

	nifa.active = 1

	if nifa.cfg.mock {
		nifa.ethHandler = ethernet.NewMockHandler()
	} else {
		ethHandler, err := ethernet.NewHandler(nifa.name)
		if err != nil {
			return errors.Wrapf(err, "Unable to create ethernet handler (%s)", nifa.name)
		}
		nifa.ethHandler = ethHandler
	}

	nifa.isP2PHelloCon = nifa.ethHandler.NewConn(ethernet.ISp2pHello)

	nifa.wg.Add(1)
	go nifa.p2pHelloSender()

	err := nifa.ethHandler.MCastJoin(ethernet.ISp2pHello)
	if err != nil {
		return errors.Wrap(err, "Unable to join IS p2p hello multicast group")
	}

	nifa.wg.Add(2)
	go nifa.receiver()

	return nil
}

func (nifa *netIfa) stop() {
	close(nifa.done)
	nifa.wg.Wait()
}

func (nifa *netIfa) broadCastL1() error {
	return fmt.Errorf("L1 Hello not supported yet")
}

func (nifa *netIfa) broadCastL2() error {
	return fmt.Errorf("broadcast networks not supported yet")
}
