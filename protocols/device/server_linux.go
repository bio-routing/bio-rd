package device

import (
	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

func (ds *Server) loadAdapter() error {
	a, err := newOSAdapterLinux(ds)
	if err != nil {
		return errors.Wrap(err, "Unable to create linux adapter")
	}

	ds.osAdapter = a
	return nil
}

type osAdapterLinux struct {
	srv    *Server
	handle *netlink.Handle
	done   chan struct{}
}

func newOSAdapterLinux(srv *Server) (*osAdapterLinux, error) {
	o := &osAdapterLinux{
		srv: srv,
	}

	h, err := netlink.NewHandle()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create netlink handle")
	}

	o.handle = h
	return o, nil
}

func (o *osAdapterLinux) start() error {
	chLU := make(chan netlink.LinkUpdate)
	err := netlink.LinkSubscribe(chLU, o.done)
	if err != nil {
		return errors.Wrap(err, "Unable to subscribe for link updates")
	}

	chAU := make(chan netlink.AddrUpdate)
	err = netlink.AddrSubscribe(chAU, o.done)
	if err != nil {
		return errors.Wrap(err, "Unable to subscribe for address updates")
	}

	err = o.init()
	if err != nil {
		return errors.Wrap(err, "Init failed")
	}

	go o.monitorLinks(chLU)
	go o.monitorAddrs(chAU)

	return nil
}

func (o *osAdapterLinux) init() error {
	links, err := o.handle.LinkList()
	if err != nil {
		return errors.Wrap(err, "Unable to get links")
	}

	for _, l := range links {
		d := linkUpdateToDevice(l.Attrs())

		for _, f := range []int{4, 6} {
			addrs, err := o.handle.AddrList(l, f)
			if err != nil {
				return errors.Wrapf(err, "Unable to get addresses for interface %s", d.Name)
			}

			for _, addr := range addrs {
				d.Addrs = append(d.Addrs, bnet.NewPfxFromIPNet(addr.IPNet))
			}
		}

		o.srv.addDevice(d)
	}

	return nil
}

func (o *osAdapterLinux) monitorAddrs(chAU chan netlink.AddrUpdate) {
	for {
		select {
		case <-o.done:
			return
		case au := <-chAU:
			o.processAddrUpdate(&au)
		}
	}
}

func (o *osAdapterLinux) monitorLinks(chLU chan netlink.LinkUpdate) {
	for {
		select {
		case <-o.done:
			return
		case lu := <-chLU:
			o.processLinkUpdate(&lu)
		}
	}
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

func (o *osAdapterLinux) processAddrUpdate(au *netlink.AddrUpdate) {
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

func (o *osAdapterLinux) processLinkUpdate(lu *netlink.LinkUpdate) {
	attrs := lu.Attrs()

	o.srv.devicesMu.Lock()
	defer o.srv.devicesMu.Unlock()

	if _, ok := o.srv.devices[uint64(attrs.Index)]; !ok {
		d := newDevice()
		d.Index = uint64(attrs.Index)
		o.srv.addDevice(d)
	}

	o.srv.devices[uint64(attrs.Index)].updateLink(attrs)
	o.srv.notify(uint64(attrs.Index))
	if attrs.OperState == netlink.OperNotPresent {
		o.srv.delDevice(uint64(attrs.Index))
		return
	}
}
