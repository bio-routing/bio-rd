package server

import (
	log "github.com/sirupsen/logrus"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
)

// UpdateSenderAddPath converts table changes into BGP update messages with add path
type UpdateSenderAddPath struct {
	routingtable.ClientManager
	fsm  *FSM
	iBGP bool
}

func newUpdateSenderAddPath(fsm *FSM) *UpdateSenderAddPath {
	return &UpdateSenderAddPath{
		fsm:  fsm,
		iBGP: fsm.peer.localASN == fsm.peer.peerASN,
	}
}

// AddPath serializes a new path and sends out a BGP update message
func (u *UpdateSenderAddPath) AddPath(pfx net.Prefix, p *route.Path) error {
	pathAttrs, err := pathAttribues(p)

	if err != nil {
		log.Errorf("Unable to create BGP Update: %v", err)
		return nil
	}
	update := &packet.BGPUpdateAddPath{
		PathAttributes: pathAttrs,
		NLRI: &packet.NLRIAddPath{
			PathIdentifier: p.BGPPath.PathIdentifier,
			IP:             pfx.Addr(),
			Pfxlen:         pfx.Pfxlen(),
		},
	}
	return serializeAndSendUpdate(u.fsm.con, update, u.fsm.options)
}

// RemovePath withdraws prefix `pfx` from a peer
func (u *UpdateSenderAddPath) RemovePath(pfx net.Prefix, p *route.Path) bool {
	err := withDrawPrefixesAddPath(u.fsm.con, u.fsm.options, pfx, p)
	return err == nil
}

// UpdateNewClient does nothing
func (u *UpdateSenderAddPath) UpdateNewClient(client routingtable.RouteTableClient) error {
	log.Warningf("BGP Update Sender: UpdateNewClient not implemented")
	return nil
}

// RouteCount returns the number of stored routes
func (u *UpdateSenderAddPath) RouteCount() int64 {
	log.Warningf("BGP Update Sender: RouteCount not implemented")
	return 0
}
