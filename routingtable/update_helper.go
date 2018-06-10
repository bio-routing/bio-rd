package routingtable

import (
	"strings"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/route"
	log "github.com/sirupsen/logrus"
)

// ShouldPropagateUpdate performs some default checks and returns if an route update should be propagated to a neighbor
func ShouldPropagateUpdate(pfx net.Prefix, p *route.Path, n *Neighbor) bool {
	return !isOwnPath(p, n) && !isDisallowedByCommunity(p, n)
}

func isOwnPath(p *route.Path, n *Neighbor) bool {
	if p.Type != n.Type {
		return false
	}

	switch p.Type {
	case route.BGPPathType:
		return p.BGPPath.Source == n.Address
	}

	return false
}

func isDisallowedByCommunity(p *route.Path, n *Neighbor) bool {
	if p.BGPPath == nil || len(p.BGPPath.Communities) == 0 {
		return false
	}

	strs := strings.Split(p.BGPPath.Communities, " ")
	for _, str := range strs {
		com, err := packet.ParseCommunityString(str)
		if err != nil {
			log.WithField("Sender", "routingtable.ShouldAnnounce()").
				WithField("community", str).
				WithError(err).
				Error("Could not parse community")
			continue
		}

		if (com == packet.WellKnownCommunityNoExport && !n.IBGP) || com == packet.WellKnownCommunityNoAdvertise {
			return true
		}
	}

	return false
}
