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
			return nil
		case lu := <-chLU:
			ds.processLinkUpdate(&lu)
		}
	}

	return nil
}

func (ds *Server) processLinkUpdate(lu *netlink.LinkUpdate) {
	attrs := lu.Attrs()

	ds.clientsByDeviceMu.RLock()
	defer ds.clientsByDeviceMu.RUnlock()

	if _, ok := ds.clientsByDevice[attrs.Name]; !ok {
		return
	}

	lu := LinkUpdate{
		Index:        uint64(attrs.Index),
		MTU:          uint16(attrs.MTU),
		Name:         attrs.Name,
		HardwareAddr: attrs.HardwareAddr,
		Flags:        attrs.Flags,
		OperState:    uint8(attrs.OperState),
	}

	for _, c := range ds.clientsByDevice[attrs.Name] {
		c.LinkUpdate(lu)
	}
}
