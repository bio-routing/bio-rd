package server

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
)

// UpdateSenderAddPath converts table changes into BGP update messages with add path
type UpdateSenderAddPath struct {
	routingtable.ClientManager
	fsm *FSM
}

func newUpdateSenderAddPath(fsm *FSM) *UpdateSenderAddPath {
	return &UpdateSenderAddPath{
		fsm: fsm,
	}
}

// AddPath serializes a new path and sends out a BGP update message
func (u *UpdateSenderAddPath) AddPath(pfx net.Prefix, p *route.Path) error {
	update, err := updateMessageForPath(pfx, p, u.fsm)
	if err != nil {
		log.Errorf("Unable to create BGP Update: %v", err)
		return nil
	}

	updateBytes, err := update.SerializeUpdate()
	if err != nil {
		log.Errorf("Unable to serialize BGP Update: %v", err)
		return nil
	}

	_, err = u.fsm.con.Write(updateBytes)
	if err != nil {
		return fmt.Errorf("Failed sending Update: %v", err)
	}
	return nil
}

// RemovePath withdraws prefix `pfx` from a peer
func (u *UpdateSenderAddPath) RemovePath(pfx net.Prefix, p *route.Path) bool {
	err := withDrawPrefixesAddPath(u.fsm.con, pfx, p)
	return err == nil
}

// UpdateNewClient does nothing
func (u *UpdateSenderAddPath) UpdateNewClient(client routingtable.RouteTableClient) error {
	log.Warningf("BGP Update Sender: RemovePath not implemented")
	return nil
}
