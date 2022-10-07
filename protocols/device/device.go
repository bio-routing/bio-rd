package device

import (
	"net"
	"sync"

	bnet "github.com/bio-routing/bio-rd/net"
)

const (
	IfOperUnknown        = 0
	IfOperNotPresent     = 1
	IfOperDown           = 2
	IfOperLowerLayerDown = 3
	IfOperTesting        = 4
	IfOperDormant        = 5
	IfOperUp             = 6
)

type DeviceInterface interface {
	GetIndex() uint64
	GetOperState() uint8
	GetAddrs() []*bnet.Prefix
}

// Device represents a network device
type Device struct {
	name         string
	index        uint64
	mtu          uint16
	HardwareAddr net.HardwareAddr
	flags        net.Flags
	operState    uint8
	addrs        []*bnet.Prefix
	l            sync.RWMutex
}

func newDevice() *Device {
	return &Device{
		addrs: make([]*bnet.Prefix, 0),
	}
}

// GetName gets the devices name
func (d *Device) GetName() string {
	return d.name
}

// GetIndex gets the interface ifIndex
func (d *Device) GetIndex() uint64 {
	return d.index
}

// GetOperState gets the operational state
func (d *Device) GetOperState() uint8 {
	return d.operState
}

// GetAddrs gets the IP addresses on the interface
func (d *Device) GetAddrs() []*bnet.Prefix {
	return d.addrs
}

func (d *Device) addAddr(pfx *bnet.Prefix) {
	d.l.Lock()
	defer d.l.Unlock()

	d.addrs = append(d.addrs, pfx)
}

func (d *Device) delAddr(del *bnet.Prefix) {
	d.l.Lock()
	defer d.l.Unlock()

	for i, pfx := range d.addrs {
		if !pfx.Equal(del) {
			continue
		}

		d.addrs = append(d.addrs[:i], d.addrs[i+1:]...)
	}
}

func (d *Device) copy() *Device {
	d.l.RLock()
	defer d.l.RUnlock()

	n := &Device{
		name:      d.name,
		index:     d.index,
		mtu:       d.mtu,
		flags:     d.flags,
		operState: d.operState,
		addrs:     make([]*bnet.Prefix, len(d.addrs)),
	}

	copy(n.HardwareAddr, d.HardwareAddr)
	copy(n.addrs, d.addrs)

	return n
}
