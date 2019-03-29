package server

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
)

type prefixLimitDecorator struct {
	client routingtable.RouteTableClient
	fsm    *FSM
}

func (d *prefixLimitDecorator) AddPath(pfx net.Prefix, p *route.Path) error {
	err := d.client.AddPath(pfx, p)
	if err == nil {
		return nil
	}

	switch err.(type) {
	case *routingtable.PrefixLimitError:
		/* TODO: sub code AdministrativeShutdown should be replaced by the corresponding sub code
		   as soon the IANA issues some for https://ftp.fau.de/ripe.net/internet-drafts/draft-sa-grow-maxprefix-02.txt */
		d.fsm.sendNotification(packet.Cease, packet.AdministrativeShutdown)
		d.fsm.cease()
	}

	return err
}

func (d *prefixLimitDecorator) RemovePath(pfx net.Prefix, p *route.Path) bool {
	return d.client.RemovePath(pfx, p)
}

func (d *prefixLimitDecorator) UpdateNewClient(c routingtable.RouteTableClient) error {
	return d.client.UpdateNewClient(c)
}

func (d *prefixLimitDecorator) Register(c routingtable.RouteTableClient) {
	d.client.Register(c)
}

func (d *prefixLimitDecorator) Unregister(c routingtable.RouteTableClient) {
	d.client.Unregister(c)
}

func (d *prefixLimitDecorator) RouteCount() int64 {
	return d.client.RouteCount()
}

func (d *prefixLimitDecorator) ClientCount() uint64 {
	return d.client.ClientCount()
}

func (d *prefixLimitDecorator) Dump() []*route.Route {
	return d.client.Dump()
}
