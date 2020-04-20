package server

import (
	"bytes"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

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
	HoldingTimer uint16
	Level1       *InterfaceLevelConfig
	Level2       *InterfaceLevelConfig
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
	name        string
	srv         *Server
	cfg         *InterfaceConfig
	active      uint64
	p2pAdjState uint8
	devStatus   device.DeviceInterface
	done        chan struct{}
	wg          sync.WaitGroup
	helloTicker btime.Ticker
	ethHandler  *ethernet.Handler
	conn        net.Conn
	ISP2PHELLO  net.Conn
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
	fmt.Printf("DeviceUpdate!\n")
	oldState := uint8(device.IfOperUnknown)
	if nifa.devStatus != nil {
		oldState = nifa.devStatus.GetOperState()
	}

	if oldState != device.IfOperUp && dev.GetOperState() == device.IfOperUp {
		log.Infof("ISIS: Interface %s came up (phy). Enabling ISIS", nifa.name)
		nifa.devStatus = dev
		fmt.Printf("Calling nifa.start()\n")
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

	ethHandler, err := ethernet.NewHandler(nifa.name)
	if err != nil {
		return errors.Wrapf(err, "Unable to create ethernet handler (%s)", nifa.name)
	}

	nifa.ethHandler = ethHandler
	nifa.ISP2PHELLO = nifa.ethHandler.NewConn(ethernet.ISP2PHELLO)

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
	fmt.Printf("This is the p2pHelloSender!\n")
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

			_, err := nifa.ISP2PHELLO.Write(hdrBuf.Bytes())
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
		HoldingTimer:   nifa.cfg.HoldingTimer,
		PDULength:      packet.P2PHelloMinLen,
		LocalCircuitID: 1,
		TLVs:           make([]packet.TLV, 4),
	}

	h.TLVs[0] = packet.NewP2PAdjacencyStateTLV(nifa.p2pAdjState, uint32(nifa.devStatus.GetIndex()))
	h.TLVs[1] = packet.NewProtocolsSupportedTLV([]uint8{
		packet.NLPIDIPv4,
		packet.NLPIDIPv6,
	})
	h.TLVs[2] = packet.NewIPInterfaceAddressesTLV([]uint32{
		10*256 ^ 3,
	})
	areas := make([]types.AreaID, 0)
	for _, net := range nifa.srv.nets {
		areas = append(areas, net.AreaID)
	}
	h.TLVs[3] = packet.NewAreaAddressesTLV(areas)

	return h
}
