package server

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/rt"
)

type UpdateSender struct {
	fsm *FSM
}

func newUpdateSender(fsm *FSM) *UpdateSender {
	return &UpdateSender{
		fsm: fsm,
	}
}

func (u *UpdateSender) AddPath(route *rt.Route) {
	log.Warningf("BGP Update Sender: AddPath not implemented")

	update := &packet.BGPUpdate{
		PathAttributes: &packet.PathAttribute{
			TypeCode: packet.OriginAttr,
			Value:    uint8(packet.IGP),
			Next: &packet.PathAttribute{
				TypeCode: packet.ASPathAttr,
				Value: packet.ASPath{
					{
						Type: 2,
						ASNs: []uint32{15169, 3329},
					},
				},
				Next: &packet.PathAttribute{
					TypeCode: packet.NextHopAttr,
					Value:    [4]byte{100, 110, 120, 130},
				},
			},
		},
		NLRI: &packet.NLRI{
			IP:     route.Prefix().Addr(),
			Pfxlen: route.Pfxlen(),
		},
	}

	updateBytes, err := update.SerializeUpdate()
	if err != nil {
		log.Errorf("Unable to serialize BGP Update: %v", err)
		return
	}
	fmt.Printf("Sending Update: %v\n", updateBytes)
	u.fsm.con.Write(updateBytes)
}

func (u *UpdateSender) ReplaceRoute(*rt.Route) {
	log.Warningf("BGP Update Sender: ReplaceRoute not implemented")
}

func (u *UpdateSender) RemovePath(*rt.Route) {
	log.Warningf("BGP Update Sender: RemovePath not implemented")
}

func (u *UpdateSender) RemoveRoute(*net.Prefix) {
	log.Warningf("BGP Update Sender: RemoveRoute not implemented")
}
