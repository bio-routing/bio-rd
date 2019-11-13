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

// Device represents a network device
type Device struct {
	Name         string
	Index        uint64
	MTU          uint16
	HardwareAddr net.HardwareAddr
	Flags        net.Flags
	OperState    uint8
	Addrs        []*bnet.Prefix
	l            sync.RWMutex
}

func newDevice() *Device {
	return &Device{
		Addrs: make([]*bnet.Prefix, 0),
	}
}

func (d *Device) addAddr(pfx *bnet.Prefix) {
	d.l.Lock()
	defer d.l.Unlock()

	d.Addrs = append(d.Addrs, pfx)
}

func (d *Device) delAddr(del *bnet.Prefix) {
	d.l.Lock()
	defer d.l.Unlock()

	for i, pfx := range d.Addrs {
		if !pfx.Equal(del) {
			continue
		}

		d.Addrs = append(d.Addrs[:i], d.Addrs[i+1:]...)
	}
}

func (d *Device) copy() *Device {
	d.l.RLock()
	defer d.l.RUnlock()

	n := &Device{
		Name:      d.Name,
		Index:     d.Index,
		MTU:       d.MTU,
		Flags:     d.Flags,
		OperState: d.OperState,
		Addrs:     make([]*bnet.Prefix, len(d.Addrs)),
	}

	copy(n.HardwareAddr, d.HardwareAddr)
	for i, a := range d.Addrs {
		n.Addrs[i] = a
	}

	return n
}
