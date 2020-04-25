package server

import (
	"bytes"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bio-routing/tflow2/convert"

	"github.com/bio-routing/bio-rd/net/ethernet"
	"github.com/bio-routing/bio-rd/protocols/device"
	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
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
	name          string
	srv           *Server
	cfg           *InterfaceConfig
	active        uint64
	p2pAdjState   uint8
	devStatus     device.DeviceInterface
	done          chan struct{}
	wg            sync.WaitGroup
	helloTicker   btime.Ticker
	ethHandler    ethernet.HandlerInterface
	conn          net.Conn
	isP2PHelloCon net.Conn
}

func newNetIfa(srv *Server, cfg *InterfaceConfig) *netIfa {
	ret := &netIfa{
		name:        cfg.Name,
		srv:         srv,
		cfg:         cfg,
		p2pAdjState: packet.P2PAdjStateDown,
		done:        make(chan struct{}),
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
		log.Infof("ISIS: Interface %s came up (phy). Enabling ISIS", nifa.name)
		nifa.devStatus = dev
		err := nifa.start()
		if err != nil {
			log.Errorf("Unable to start ISIS on interface %s", nifa.name)
		}
	}
}

func (nifa *netIfa) start() error {
	log.Infof("Starting ISIS on %s", nifa.name)

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

	fmt.Printf("TODO: start hello sender and packet receiver\n")
	// TODO: start hello sender and packet receiver

	nifa.wg.Add(1)
	go nifa.p2pHelloSender()

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

func (nifa *netIfa) p2pHelloSender() {
	for {
		select {
		case <-nifa.done:
			nifa.helloTicker.Stop()
			nifa.wg.Done()
			return
		case <-nifa.helloTicker.C():
			fmt.Printf("Sending Hello!\n")

			hello := nifa.p2pHello()
			helloBuf := bytes.NewBuffer(nil)
			hello.Serialize(helloBuf)

			hdr := packet.ISISHeader{
				ProtoDiscriminator:  0x83,
				LengthIndicator:     packet.P2PHelloMinLen,
				ProtocolIDExtension: 1,
				IDLength:            0,
				PDUType:             packet.P2P_HELLO,
				Version:             1,
				MaxAreaAddresses:    0,
			}

			hdrBuf := bytes.NewBuffer(nil)
			hdr.Serialize(hdrBuf)
			hdrBuf.Write(helloBuf.Bytes())

			_, err := nifa.isP2PHelloCon.Write(hdrBuf.Bytes())
			if err != nil {
				panic(err)
			}
		}

	}
}

func (nifa *netIfa) p2pHello() *packet.P2PHello {
	circuitType := uint8(0)
	if nifa.cfg.Level1 != nil {
		circuitType++
	}
	if nifa.cfg.Level2 != nil {
		circuitType += 2
	}

	h := &packet.P2PHello{
		CircuitType:    circuitType,
		SystemID:       nifa.srv.nets[0].SystemID,
		HoldingTimer:   nifa.cfg.holdingTimer(),
		PDULength:      packet.P2PHelloMinLen,
		LocalCircuitID: 1,
		TLVs:           make([]packet.TLV, 0, 5),
	}

	h.TLVs = append(h.TLVs, packet.NewP2PAdjacencyStateTLV(nifa.p2pAdjState, uint32(nifa.devStatus.GetIndex())))
	h.TLVs = append(h.TLVs, packet.NewProtocolsSupportedTLV([]uint8{
		packet.NLPIDIPv4,
		packet.NLPIDIPv6,
	}))

	ipv4Addrs := make([]uint32, 0)
	for _, a := range nifa.devStatus.GetAddrs() {
		if !a.Addr().IsIPv4() {
			continue
		}

		ipv4Addrs = append(ipv4Addrs, convert.Uint32(convert.Reverse(a.Addr().Bytes())))
	}
	h.TLVs = append(h.TLVs, packet.NewIPInterfaceAddressesTLV(ipv4Addrs))

	areas := make([]types.AreaID, 0)
	for _, net := range nifa.srv.nets {
		areas = append(areas, append([]byte{net.AFI}, net.AreaID...))
	}
	h.TLVs = append(h.TLVs, packet.NewAreaAddressesTLV(areas))

	return h
}
