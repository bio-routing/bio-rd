package device

import (
	"fmt"

	bnet "github.com/bio-routing/bio-rd/net"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

type osAdapter struct {
	srv    *Server
	handle *netlink.Handle
	done   chan struct{}
}

func newOSAdapter(srv *Server) (*osAdapter, error) {
	o := &osAdapter{
		srv: srv,
	}

	h, err := netlink.NewHandle()
	if err != nil {
		return nil, fmt.Errorf("Failed to create netlink handle: %v", err)
	}

	o.handle = h
	return o, nil
}

func (o *osAdapter) start() error {
	chLU := make(chan netlink.LinkUpdate)
	err := netlink.LinkSubscribe(chLU, o.done)
	if err != nil {
		return fmt.Errorf("Unable to subscribe for link updates: %v", err)
	}

	chAU := make(chan netlink.AddrUpdate)
	err = netlink.AddrSubscribe(chAU, o.done)
	if err != nil {
		return fmt.Errorf("Unable to subscribe for address updates: %v", err)
	}

	err = o.init()
	if err != nil {
		return fmt.Errorf("Init failed: %v", err)
	}

	go o.monitorLinks(chLU)
	go o.monitorAddrs(chAU)

	return nil
}

func (o *osAdapter) init() error {
	links, err := o.handle.LinkList()
	if err != nil {
		return fmt.Errorf("Unable to get links: %v", err)
	}

	for _, l := range links {
		d := linkUpdateToDevice(l.Attrs())

		for _, f := range []int{4, 6} {
			addrs, err := o.handle.AddrList(l, f)
			if err != nil {
				return fmt.Errorf("Unable to get addresses for interface %s: %v", d.Name, err)
			}

			for _, addr := range addrs {
				d.Addrs = append(d.Addrs, bnet.NewPfxFromIPNet(addr.IPNet))
			}
		}

		o.srv.addDevice(d)
	}

	return nil
}

func (o *osAdapter) monitorAddrs(chAU chan netlink.AddrUpdate) {
	for {
		select {
		case <-o.done:
			return
		case au := <-chAU:
			o.processAddrUpdate(&au)
		}
	}
}

func (o *osAdapter) monitorLinks(chLU chan netlink.LinkUpdate) {
	for {
		select {
		case <-o.done:
			return
		case lu := <-chLU:
			o.processLinkUpdate(&lu)
		}
	}

	return
}

func linkUpdateToDevice(attrs *netlink.LinkAttrs) *Device {
	return &Device{
		Index:        uint64(attrs.Index),
		MTU:          uint16(attrs.MTU),
		Name:         attrs.Name,
		HardwareAddr: attrs.HardwareAddr,
		Flags:        attrs.Flags,
		OperState:    uint8(attrs.OperState),
	}
}

func (o *osAdapter) processAddrUpdate(au *netlink.AddrUpdate) {
	o.srv.devicesMu.RLock()
	defer o.srv.devicesMu.RUnlock()

	if _, ok := o.srv.devices[uint64(au.LinkIndex)]; !ok {
		log.Warningf("Received address update for non existent device index %d", au.LinkIndex)
		return
	}

	d := o.srv.devices[uint64(au.LinkIndex)]
	if au.NewAddr {
		d.addAddr(bnet.NewPfxFromIPNet(&au.LinkAddress))
		return
	}

	d.delAddr(bnet.NewPfxFromIPNet(&au.LinkAddress))
}

func (o *osAdapter) processLinkUpdate(lu *netlink.LinkUpdate) {
	attrs := lu.Attrs()

	o.srv.devicesMu.Lock()
	defer o.srv.devicesMu.Unlock()

	if _, ok := o.srv.devices[uint64(attrs.Index)]; !ok {
		o.srv.devices[uint64(attrs.Index)] = newDevice()
		o.srv.devices[uint64(attrs.Index)].updateLink(attrs)
		o.notify(uint64(attrs.Index))

		if attrs.OperState == netlink.OperNotPresent {
			delete(o.srv.devices, uint64(attrs.Index))
			return
		}

		return
	}

	d := linkUpdateToDevice(attrs)
	o.srv.devices[d.Index] = d
	o.notify(uint64(d.Index))
}

func (o *osAdapter) notify(index uint64) {
	o.srv.clientsByDeviceMu.RLock()
	defer o.srv.clientsByDeviceMu.RUnlock()

	for i, d := range o.srv.devices {
		if i != index {
			continue
		}

		d.notify(o.srv.clientsByDevice[d.Name])
	}

}
