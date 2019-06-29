package kernel

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
)

type Kernel struct {
	osKernel osKernel
}

type osKernel interface {
	AddPath(pfx net.Prefix, path *route.Path) error
	RemovePath(pfx net.Prefix, path *route.Path) bool
	uninit() error
}

func New() (*Kernel, error) {
	k := &Kernel{}
	err := k.init()
	if err != nil {
		return nil, err
	}

	return k, nil
}

func (k *Kernel) AddPath(pfx net.Prefix, path *route.Path) error {
	return k.osKernel.AddPath(pfx, path)
}

func (k *Kernel) RemovePath(pfx net.Prefix, path *route.Path) bool {
	return k.osKernel.RemovePath(pfx, path)
}

func (k *Kernel) UpdateNewClient(routingtable.RouteTableClient) error {
	return nil
}

func (k *Kernel) Register(routingtable.RouteTableClient) {
}

func (k *Kernel) RegisterWithOptions(routingtable.RouteTableClient, routingtable.ClientOptions) {
}

func (k *Kernel) Unregister(routingtable.RouteTableClient) {
}

func (k *Kernel) RouteCount() int64 {
	return -1
}

func (k *Kernel) ClientCount() uint64 {
	return 0
}

func (k *Kernel) Dump() []*route.Route {
	return nil
}

func (k *Kernel) Dispose() {
	k.osKernel.uninit()
}
