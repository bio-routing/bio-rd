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
	fsm *FSM
}

func newUpdateSender(fsm *FSM) *UpdateSender {
	return &UpdateSender{
		fsm: fsm,
	}
}

// AddPath serializes a new path and sends out a BGP update message
func (u *UpdateSender) AddPath(pfx net.Prefix, p *route.Path) error {
	asPathPA, err := packet.ParseASPathStr(strings.TrimRight(fmt.Sprintf("%d %s", u.fsm.localASN, p.BGPPath.ASPath), " "))
	if err != nil {
		return fmt.Errorf("Unable to parse AS path: %v", err)
	}

	update := &packet.BGPUpdate{
		PathAttributes: &packet.PathAttribute{
			TypeCode: packet.OriginAttr,
			Value:    p.BGPPath.Origin,
			Next: &packet.PathAttribute{
				TypeCode: packet.ASPathAttr,
				Value:    asPathPA.Value,
				Next: &packet.PathAttribute{
					TypeCode: packet.NextHopAttr,
					Value:    p.BGPPath.NextHop,
					Next: &packet.PathAttribute{
						TypeCode: packet.LocalPrefAttr,
						Value:    p.BGPPath.LocalPref,
					},
				},
			},
		},
		NLRI: &packet.NLRI{
			IP:     pfx.Addr(),
			Pfxlen: pfx.Pfxlen(),
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
