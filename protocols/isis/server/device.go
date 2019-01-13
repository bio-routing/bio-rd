package server

import (
	"fmt"
	"sync"

	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/protocols/device"
	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
)

type dev struct {
	name               string
	srv                *Server
	up                 bool
	passive            bool
	p2p                bool
	level2             *level
	supportedProtocols []uint8
	phy                *device.Device
	//phyMu              sync.RWMutex
	done chan struct{}
	wg   sync.WaitGroup

	helloMethod    func()
	receiverMethod func()
}

type level struct {
	HelloInterval uint16
	HoldTime      uint16
	Metric        uint32
	neighbors     *neighbors
}

func newDev(srv *Server, ifcfg *config.ISISInterfaceConfig) *dev {
	d := &dev{
		name:               ifcfg.Name,
		srv:                srv,
		passive:            ifcfg.Passive,
		p2p:                ifcfg.P2P,
		supportedProtocols: []uint8{packet.NLPIDIPv4, packet.NLPIDIPv6},
		done:               make(chan struct{}),
	}

	d.helloMethod = d.receiverRoutine
	d.receiverMethod = d.receiverRoutine

	if ifcfg.ISISLevel2Config != nil {
		d.level2 = &level{}
		d.level2.HelloInterval = ifcfg.ISISLevel2Config.HelloInterval
		d.level2.HoldTime = ifcfg.ISISLevel2Config.HoldTime
		d.level2.Metric = ifcfg.ISISLevel2Config.Metric
		d.level2.neighbors = newNeighbors()
	}

	//srv.ds.Subscribe(d, ifcfg.Name)
	return d
}

// DeviceUpdate receives interface status information and manages ISIS interface state
func (d *dev) DeviceUpdate(phy *device.Device) {
	d.phy = phy
	if d.phy.OperState == device.IfOperUp {
		err := d.enable()
		if err != nil {
			log.Errorf("Unable to enable ISIS on %q: %v", d.name, err)
		}
		return
	}

	err := d.disable()
	if err != nil {
		log.Errorf("Unable to disable ISIS on %q: %v", d.name, err)
		return
	}
}

func (d *dev) enable() error {
	err := d.srv.sys.openPacketSocket()
	if err != nil {
		return fmt.Errorf("Failed to open packet socket: %v", err)
	}

	err = d.srv.sys.mcastJoin(packet.AllP2PISS)
	if err != nil {
		return fmt.Errorf("Failed to join multicast group: %v", err)
	}

	d.done = make(chan struct{})

	d.wg.Add(1)
	go d.receiverMethod()

	d.wg.Add(1)
	go d.helloMethod()

	log.Infof("ISIS: Interface %q is now up", d.name)
	d.up = true
	return nil
}

func (d *dev) disable() error {
	close(d.done)

	err := d.srv.sys.closePacketSocket()
	if err != nil {
		return errors.Wrap(err, "Unable to close socket")
	}

	d.wg.Wait()
	d.up = false
	return nil
}

func (d *dev) receiverRoutine() {

}

/*func (d *dev) receiverRoutine() {
	for {
		select {
		case <-d.done:
			return
		default:
			rawPkt, src, err := d.sys.recvPacket()
			if err != nil {
				log.Errorf("recvPacket() failed: %v", err)
				return
			}

			d.processIngressPacket(rawPkt, src)
		}
	}
}
*/
