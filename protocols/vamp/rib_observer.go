package vamp

import (
	"github.com/bio-routing/bio-rd/net"
	vapi "github.com/bio-routing/bio-rd/protocols/vamp/api"
	"github.com/bio-routing/bio-rd/route"
	rapi "github.com/bio-routing/bio-rd/route/api"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
)

type RIBObserver struct {
	vrfID    uint32
	locRIB   *locRIB.LocRIB
	updateCh chan *vapi.VAMPMessage
}

func newRIBOserver(locRIB *locRIB.LocRIB, vrfID uint32, updateCh chan *vapi.VAMPMessage) *RIBObserver {
	return &RIBObserver{
		vrfID:    vrfID,
		locRIB:   locRIB,
		updateCh: updateCh,
	}
}

func (ro *RIBObserver) Start() {
	ro.locRIB.Register(ro)
}

func (ro *RIBObserver) Stop() {
	ro.locRIB.Unregister(ro)
}

func (ro *RIBObserver) sendMsg(msg *vapi.VAMPMessage) {
	ro.updateCh <- msg
}

func (ro *RIBObserver) sendRouteInfoMsg(pfx *net.Prefix, path *route.Path, announcement bool) {
	msg := &vapi.VAMPMessage{
		MessageType: vapi.VAMPMessage_ROUTE_INFO,
		VrfId:       ro.vrfID,
		RouteInfo: &vapi.RouteInfo{
			Route: &rapi.Route{
				Pfx: pfx.ToProto(),
				Paths: []*rapi.Path{
					path.ToProto(),
				},
			},
			Announcement: announcement,
		},
	}

	ro.sendMsg(msg)
}

/*
 * RoutingTableClient interface
 */
func (ro *RIBObserver) AddPath(pfx *net.Prefix, path *route.Path) error {
	ro.sendRouteInfoMsg(pfx, path, true)
	return nil
}

func (ro *RIBObserver) AddPathInitialDump(pfx *net.Prefix, path *route.Path) error {
	ro.sendRouteInfoMsg(pfx, path, true)
	return nil
}

func (ro *RIBObserver) EndOfRIB() {
	ro.sendMsg(&vapi.VAMPMessage{
		MessageType: vapi.VAMPMessage_END_OF_RIB,
		VrfId:       ro.vrfID,
	})
}

func (ro *RIBObserver) RemovePath(pfx *net.Prefix, path *route.Path) bool {
	ro.sendRouteInfoMsg(pfx, path, false)
	return true
}

func (ro *RIBObserver) ReplacePath(pfx *net.Prefix, oldPath *route.Path, newPath *route.Path) {
	ro.sendRouteInfoMsg(pfx, oldPath, true)
	ro.sendRouteInfoMsg(pfx, newPath, false)
}

func (ro *RIBObserver) RefreshRoute(*net.Prefix, []*route.Path) {
	// TODO
}

// A call to Dispose() signals that no more updates are to be expected from the RIB the client is registered to.
func (ro *RIBObserver) Dispose() {}
