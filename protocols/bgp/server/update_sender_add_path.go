package server

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
)

// UpdateSenderAddPath converts table changes into BGP update messages with add path
type UpdateSenderAddPath struct {
	routingtable.ClientManager
	fsm      *FSM
	iBGP     bool
	buffer   chan *route.Route
	toSendMu sync.Mutex
	toSend   map[string]struct {
		path *route.Path
		pfxs []bnet.Prefix
	}
}

func newUpdateSenderAddPath(fsm *FSM) *UpdateSenderAddPath {
	return &UpdateSenderAddPath{
		fsm:    fsm,
		iBGP:   fsm.peer.localASN == fsm.peer.peerASN,
		buffer: make(chan *route.Route, 1000),
	}
}

func (u *UpdateSenderAddPath) AddPath(pfx bnet.Prefix, p *route.Path) error {
	u.buffer <- route.NewRoute(pfx, p)
	return nil
}

// sender serializes BGP update messages
func (u *UpdateSenderAddPath) sender() {
	ticker := time.NewTicker(time.Millisecond * 5)
	var r *route.Route
	var p *route.Path
	var pfx bnet.Prefix
	var err error
	var pathAttrs *packet.PathAttribute
	var update *packet.BGPUpdateAddPath
	var lastNLRI *packet.NLRIAddPath
	var budget int64

	for {
		<-ticker.C
		u.toSendMu.Lock()

		for _, pathNLRIs := range u.toSend {
			budget = packet.MaxLen

			pathAttrs, err = pathAttribues(pathNLRIs.path)
			if err != nil {
				log.Errorf("Unable to get path attributes: %v", err)
				continue
			}

			update = &packet.BGPUpdateAddPath{
				PathAttributes: pathAttrs,
				/*NLRI: &packet.NLRIAddPath{
					PathIdentifier: p.BGPPath.PathIdentifier,
					IP:             pfx.Addr(),
					Pfxlen:         pfx.Pfxlen(),
				},*/
			}

			for _, pfx := range pathNLRIs.pfxs {
				nlri = &packet.NLRIAddPath{
					PathIdentifier: pathNLRIs.path.BGPPath.PathIdentifier,
					IP:             pfx.Addr(),
					Pfxlen:         pfx.Pfxlen(),
				}
			}
		}

		u.toSendMu.Unlock()

		p = r.Paths()[0]
		pfx = r.Prefix()
		pathAttrs, err = pathAttribues(p)

		if err != nil {
			log.Errorf("Unable to create BGP Update: %v", err)
			continue
		}
		update = &packet.BGPUpdateAddPath{
			PathAttributes: pathAttrs,
			NLRI: &packet.NLRIAddPath{
				PathIdentifier: p.BGPPath.PathIdentifier,
				IP:             pfx.Addr(),
				Pfxlen:         pfx.Pfxlen(),
			},
		}

		err = serializeAndSendUpdate(u.fsm.con, update, u.fsm.options)
		if err != nil {
			log.Errorf("Failed to serialize and send: %v", err)
		}
	}

}

// RemovePath withdraws prefix `pfx` from a peer
func (u *UpdateSenderAddPath) RemovePath(pfx bnet.Prefix, p *route.Path) bool {
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
