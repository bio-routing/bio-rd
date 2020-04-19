package device

import "github.com/vishvananda/netlink"

func (d *Device) updateLink(attrs *netlink.LinkAttrs) {
	d.l.Lock()
	defer d.l.Unlock()

	d.mtu = uint16(attrs.MTU)
	d.name = attrs.Name
	copy(d.HardwareAddr, attrs.HardwareAddr)
	d.flags = attrs.Flags
	d.operState = uint8(attrs.OperState)
}

func (d *Device) notify(clients []Client) {
	n := d.copy()

	for _, c := range clients {
		c.DeviceUpdate(n)
	}
}
