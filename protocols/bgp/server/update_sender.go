package server

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
)

// UpdateSender converts table changes into BGP update messages
type UpdateSender struct {
	routingtable.ClientManager
	fsm  *FSM
	iBGP bool
}

func newUpdateSender(fsm *FSM) *UpdateSender {
	return &UpdateSender{
		fsm:  fsm,
		iBGP: fsm.localASN == fsm.remoteASN,
	}
}

// AddPath serializes a new path and sends out a BGP update message
func (u *UpdateSender) AddPath(pfx net.Prefix, p *route.Path) error {
	pathAttrs, err := pathAttribues(p, u.fsm)
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
func (u *UpdateSender) RemovePath(pfx net.Prefix, p *route.Path) bool {
	err := withDrawPrefixes(u.fsm.con, pfx)
	return err == nil
}

// UpdateNewClient does nothing
func (u *UpdateSender) UpdateNewClient(client routingtable.RouteTableClient) error {
	log.Warningf("BGP Update Sender: UpdateNewClient() not supported")
	return nil
}

func asPathString(iBGP bool, localASN uint16, asPath string) string {
	ret := ""
	if iBGP {
		ret = ret + fmt.Sprintf("%d ", localASN)
	}
	ret = ret + asPath
	return strings.TrimRight(ret, " ")
}
