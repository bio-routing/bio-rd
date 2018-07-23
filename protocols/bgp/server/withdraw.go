package server

import (
	"errors"
	"io"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/route"
)

// withdrawPrefixes generates a BGPUpdate message and write it to the given
// io.Writer.
func withdrawPrefixes(out io.Writer, opt *types.Options, prefixes ...net.Prefix) error {
	if len(prefixes) < 1 {
		return nil
	}
	var rootNLRI *packet.NLRI
	var currentNLRI *packet.NLRI
	for _, pfx := range prefixes {
		if rootNLRI == nil {
			rootNLRI = &packet.NLRI{
				IP:     pfx.Addr().ToUint32(),
				Pfxlen: pfx.Pfxlen(),
			}
			currentNLRI = rootNLRI
		} else {
			currentNLRI.Next = &packet.NLRI{
				IP:     pfx.Addr().ToUint32(),
				Pfxlen: pfx.Pfxlen(),
			}
			currentNLRI = currentNLRI.Next
		}
	}
	update := &packet.BGPUpdate{
		WithdrawnRoutes: rootNLRI,
	}
	return serializeAndSendUpdate(out, update, opt)

}

// withdrawPrefixesAddPath generates a BGPUpdateAddPath message and write it to the given
// io.Writer.
func withdrawPrefixesAddPath(out io.Writer, opt *types.Options, pfx net.Prefix, p *route.Path) error {
	if p.Type != route.BGPPathType {
		return errors.New("wrong path type, expected BGPPathType")
	}
	if p.BGPPath == nil {
		return errors.New("got nil BGPPath")
	}
	update := &packet.BGPUpdate{
		WithdrawnRoutes: &packet.NLRI{
			PathIdentifier: p.BGPPath.PathIdentifier,
			IP:             pfx.Addr().ToUint32(),
			Pfxlen:         pfx.Pfxlen(),
		},
	}
	return serializeAndSendUpdate(out, update, opt)
}

func withdrawPrefixesMultiProtocol(out io.Writer, opt *types.Options, pfx net.Prefix, afi uint16, safi uint8) error {
	update := &packet.BGPUpdate{
		PathAttributes: &packet.PathAttribute{
			TypeCode: packet.MultiProtocolUnreachNLRICode,
			Value: packet.MultiProtocolUnreachNLRI{
				AFI:      afi,
				SAFI:     safi,
				Prefixes: []net.Prefix{pfx},
			},
		},
	}

	return serializeAndSendUpdate(out, update, opt)
}
