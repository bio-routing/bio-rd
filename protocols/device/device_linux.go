package device

import "github.com/vishvananda/netlink"

func (d *Device) updateLink(attrs *netlink.LinkAttrs) {
	d.l.Lock()
	defer d.l.Unlock()

	d.MTU = attrs.MTU
	d.Name = attrs.Name
	copy(d.HardwareAddr, attrs.HardwareAddr)
	d.Flags = attrs.Flags
	d.OperState = attrs.OperState
}

func (d *Device) notify(clients []Client) {
	n := d.copy()

	for _, c := range clients {
		c.LinkUpdate(n)
	}
}
