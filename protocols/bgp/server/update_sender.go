package server

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
)

type UpdateSender struct {
	fsm *FSM
}

func newUpdateSender(fsm *FSM) *UpdateSender {
	return &UpdateSender{
		fsm: fsm,
	}
}

func (u *UpdateSender) AddPath(pfx net.Prefix, p *route.Path) error {
	fmt.Printf("SENDING AN BGP UPDATE\n")
	asPathPA, err := packet.ParseASPathStr(fmt.Sprintf("%d %s", u.fsm.localASN, p.BGPPath.ASPath))
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
	fmt.Printf("Sending Update: %v\n", updateBytes)
	_, err = u.fsm.con.Write(updateBytes)
	if err != nil {
		return fmt.Errorf("Failed sending Update: %v", err)
	}
	return nil
}

func (u *UpdateSender) RemovePath(pfx net.Prefix, p *route.Path) bool {
	log.Warningf("BGP Update Sender: RemovePath not implemented")
	return false
}

func (u *UpdateSender) UpdateNewClient(client routingtable.RouteTableClient) error {
	log.Warningf("BGP Update Sender: RemovePath not implemented")
	return nil
}
