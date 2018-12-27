package device

import (
	"net"
	"sync"

	bnet "github.com/bio-routing/bio-rd/net"
)

type LinkUpdate struct {
	Index        uint64
	MTU          uint16
	Name         string
	HardwareAddr net.HardwareAddr
	Flags        net.Flags
	OperState    uint8
}

type Device struct {
	Name         string
	Index        uint64
	MTU          uint16
	HardwareAddr net.HardwareAddr
	Flags        net.Flags
	OperState    uint8
	Addrs        []bnet.Prefix
	l            sync.RWMutex
}

func newDevice() *Device {
	return &Device{
		Addrs: make([]bnet.Prefix, 0),
	}
}

func (d *Device) addAddr(pfx bnet.Prefix) {
	d.l.Lock()
	defer d.l.Unlock()

	d.Addrs = append(d.Addrs, pfx)
}

func (d *Device) delAddr(del bnet.Prefix) {
	d.l.Lock()
	defer d.l.Unlock()

	for i, pfx := range d.Addrs {
		if !pfx.Equal(del) {
			continue
		}

		d.Addrs = append(d.Addrs[:i], d.Addrs[i+1:]...)
	}
}
