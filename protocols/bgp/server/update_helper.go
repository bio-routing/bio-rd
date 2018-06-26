package server

import (
	"fmt"
	"io"

	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/route"
	log "github.com/sirupsen/logrus"
)

func pathAttribues(p *route.Path) (*packet.PathAttribute, error) {
	asPath := &packet.PathAttribute{
		TypeCode: packet.ASPathAttr,
		Value:    p.BGPPath.ASPath,
	}

	origin := &packet.PathAttribute{
		TypeCode: packet.OriginAttr,
		Value:    p.BGPPath.Origin,
	}
	asPath.Next = origin

	nextHop := &packet.PathAttribute{
		TypeCode: packet.NextHopAttr,
		Value:    p.BGPPath.NextHop,
	}
	origin.Next = nextHop

	localPref := &packet.PathAttribute{
		TypeCode: packet.LocalPrefAttr,
		Value:    p.BGPPath.LocalPref,
	}
	nextHop.Next = localPref

	if p.BGPPath != nil {
		optionals := addOptionalPathAttribues(p, localPref)
		optionals.Next = p.BGPPath.UnknownAttributes
	}

	return asPath, nil
}

func addOptionalPathAttribues(p *route.Path, parent *packet.PathAttribute) *packet.PathAttribute {
	current := parent

	if len(p.BGPPath.Communities) > 0 {
		communities := &packet.PathAttribute{
			TypeCode: packet.CommunitiesAttr,
			Value:    p.BGPPath.Communities,
		}
		current.Next = communities
		current = communities
	}

	if len(p.BGPPath.LargeCommunities) > 0 {
		largeCommunities := &packet.PathAttribute{
			TypeCode: packet.LargeCommunitiesAttr,
			Value:    p.BGPPath.LargeCommunities,
		}
		current.Next = largeCommunities
		current = largeCommunities
	}

	return current
}

func serializeAndSendUpdate(out io.Writer, update *packet.BGPUpdate, opt *packet.Options) error {
	updateBytes, err := update.SerializeUpdate(opt)
	if err != nil {
		log.Errorf("Unable to serialize BGP Update: %v", err)
		return nil
	}

	_, err = out.Write(updateBytes)
	if err != nil {
		return fmt.Errorf("Failed sending Update: %v", err)
	}
	return nil
}
