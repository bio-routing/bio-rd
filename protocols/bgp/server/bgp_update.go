package server

import (
	"fmt"
	"strings"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/route"
)

func updateMessageForPath(pfx net.Prefix, p *route.Path, fsm *FSM) (*packet.BGPUpdate, error) {
	pathAttrs, err := pathAttribues(p, fsm)
	if err != nil {
		return nil, err
	}

	return &packet.BGPUpdate{
		PathAttributes: pathAttrs,
		NLRI: &packet.NLRI{
			IP:     pfx.Addr(),
			Pfxlen: pfx.Pfxlen(),
		},
	}, nil
}

func pathAttribues(p *route.Path, fsm *FSM) (*packet.PathAttribute, error) {
	asPathPA, err := packet.ParseASPathStr(strings.TrimRight(fmt.Sprintf("%d %s", fsm.localASN, p.BGPPath.ASPath), " "))
	if err != nil {
		return nil, fmt.Errorf("Unable to parse AS path: %v", err)
	}

	origin := &packet.PathAttribute{
		TypeCode: packet.OriginAttr,
		Value:    p.BGPPath.Origin,
		Next:     asPathPA,
	}

	nextHop := &packet.PathAttribute{
		TypeCode: packet.NextHopAttr,
		Value:    p.BGPPath.NextHop,
	}
	asPathPA.Next = nextHop

	if p.BGPPath != nil {
		err := addOptionalPathAttribues(p, nextHop)

		if err != nil {
			return nil, err
		}
	}

	return origin, nil
}

func addOptionalPathAttribues(p *route.Path, parent *packet.PathAttribute) error {
	current := parent

	if len(p.BGPPath.LargeCommunities) > 0 {
		largeCommunities, err := packet.LargeCommunityAttributeForString(p.BGPPath.LargeCommunities)
		if err != nil {
			return fmt.Errorf("Could not create large community attribute: %v", err)
		}

		current.Next = largeCommunities
		current = largeCommunities
	}

	return nil
}
