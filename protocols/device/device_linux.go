package device

import "github.com/vishvananda/netlink"

func (d *Device) updateLink(lu *netlink.LinkUpdate) {
	d.l.Lock()
	defer d.l.Unlock()

	d.MTU = lu.MTU
	d.Name = lu.Name
	d.HardwareAddr = lu.HardwareAddr
	d.Flags = lu.Flags
	d.OperState = lu.OperState
}

func (d *Device) notify(clients []Client) {
	n := d.copy()

	for _, c := range clients {
		c.LinkUpdate(n)
	}
}
