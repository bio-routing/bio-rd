package device

import (
	"fmt"

	"github.com/vishvananda/netlink"
)

func (ds *Server) monitorDevices() error {
	chLU := make(chan netlink.LinkUpdate)
	chDone := make(chan struct{})

	err := netlink.LinkSubscribe(chLU, chDone)
	if err != nil {
		return fmt.Errorf("Unable to subscribe to link changes: %v", err)
	}

	for {
		select {
		case <-ds.done:
			chDone <- struct{}{}
			return
		case lu := <-chLU:
			ds.processLinkUpdate(&lu)
		}
	}

	return nil
}

func (ds *Server) processLinkUpdate(lu *netlink.LinkUpdate) {
	attrs := lu.Attrs()

	ds.devicesMu.RLock()
	defer ds.devicesMu.RUnlock()

	for _, d := range ds.devices {
		if d.Name != lu.Name {
			continue
		}

		d.clientsMu.RLock()
		defer d.clientsMu.RUnlock()
		for _, c := range d.clients {
			c.LinkUpdate(LinkUpdate{
				Index:        uint64(attrs.Index),
				MTU:          uint16(attrs.MTU),
				Name:         attrs.Name,
				HardwareAddr: attrs.HardwareAddr,
				Flags:        attrs.Flags,
				OperState:    attrs.OperState,
			})
		}
	}
}