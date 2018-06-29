package server

import (
	"sync"
	"time"

	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"

	bnet "github.com/bio-routing/bio-rd/net"
	log "github.com/sirupsen/logrus"
)

// UpdateSender converts table changes into BGP update messages
type UpdateSender struct {
	routingtable.ClientManager
	fsm       *FSM
	iBGP      bool
	toSendMu  sync.Mutex
	toSend    map[string]*pathPfxs
	destroyCh chan struct{}
}

type pathPfxs struct {
	path *route.Path
	pfxs []bnet.Prefix
}

func newUpdateSender(fsm *FSM) *UpdateSender {
	return &UpdateSender{
		fsm:       fsm,
		iBGP:      fsm.peer.localASN == fsm.peer.peerASN,
		destroyCh: make(chan struct{}),
		toSend:    make(map[string]*pathPfxs),
	}
}

// Start starts the update sender
func (u *UpdateSender) Start(aggrTime time.Duration) {
	go u.sender(aggrTime)
}

// Destroy destroys everything (with greetings to Hatebreed)
func (u *UpdateSender) Destroy() {
	u.destroyCh <- struct{}{}
}

// AddPath adds path p for pfx to toSend queue
func (u *UpdateSender) AddPath(pfx bnet.Prefix, p *route.Path) error {
	u.toSendMu.Lock()

	hash := p.BGPPath.ComputeHash()
	if _, exists := u.toSend[hash]; exists {
		u.toSend[hash].pfxs = append(u.toSend[hash].pfxs, pfx)
		u.toSendMu.Unlock()
		return nil
	}

	u.toSend[p.BGPPath.ComputeHash()] = &pathPfxs{
		path: p,
		pfxs: []bnet.Prefix{
			pfx,
		},
	}

	u.toSendMu.Unlock()
	return nil
}

// sender serializes BGP update messages
func (u *UpdateSender) sender(aggrTime time.Duration) {
	ticker := time.NewTicker(aggrTime)
	var err error
	var pathAttrs *packet.PathAttribute
	var budget int

	for {
		select {
		case <-u.destroyCh:
			return
		case <-ticker.C:
		}

		u.toSendMu.Lock()

		for key, pathNLRIs := range u.toSend {
			budget = packet.MaxLen - packet.HeaderLen - packet.MinUpdateLen - int(pathNLRIs.path.BGPPath.Length())
			pathAttrs, err = packet.PathAttributes(pathNLRIs.path, u.iBGP)
			if err != nil {
				log.Errorf("Unable to get path attributes: %v", err)
				continue
			}

			updatesPrefixes := make([][]bnet.Prefix, 0, 1)
			prefixes := make([]bnet.Prefix, 0, 1)
			for _, pfx := range pathNLRIs.pfxs {
				budget -= int(packet.BytesInAddr(pfx.Pfxlen())) + 1
				if budget < 0 {
					updatesPrefixes = append(updatesPrefixes, prefixes)
					prefixes = make([]bnet.Prefix, 0, 1)
					budget = packet.MaxLen - int(pathNLRIs.path.BGPPath.Length())
				}

				prefixes = append(prefixes, pfx)
			}
			if len(prefixes) > 0 {
				updatesPrefixes = append(updatesPrefixes, prefixes)
			}

			delete(u.toSend, key)
			u.toSendMu.Unlock()

			u.sendUpdates(pathAttrs, updatesPrefixes, pathNLRIs.path.BGPPath.PathIdentifier)
			u.toSendMu.Lock()
		}
		u.toSendMu.Unlock()
	}
}

func (u *UpdateSender) sendUpdates(pathAttrs *packet.PathAttribute, updatePrefixes [][]bnet.Prefix, pathID uint32) {
	var nlri *packet.NLRI
	var err error

	for _, updatePrefix := range updatePrefixes {
		update := &packet.BGPUpdate{
			PathAttributes: pathAttrs,
		}

		for _, pfx := range updatePrefix {
			nlri = &packet.NLRI{
				PathIdentifier: pathID,
				IP:             pfx.Addr(),
				Pfxlen:         pfx.Pfxlen(),
				Next:           update.NLRI,
			}
			update.NLRI = nlri
		}

		err = serializeAndSendUpdate(u.fsm.con, update, u.fsm.options)
		if err != nil {
			log.Errorf("Failed to serialize and send: %v", err)
		}
	}
}

// RemovePath withdraws prefix `pfx` from a peer
func (u *UpdateSender) RemovePath(pfx bnet.Prefix, p *route.Path) bool {
	err := withDrawPrefixesAddPath(u.fsm.con, u.fsm.options, pfx, p)
	if err != nil {
		log.Errorf("Unable to withdraw prefix: %v", err)
		return false
	}
	return true
}

// UpdateNewClient does nothing
func (u *UpdateSender) UpdateNewClient(client routingtable.RouteTableClient) error {
	log.Warningf("BGP Update Sender: UpdateNewClient not implemented")
	return nil
}

// RouteCount returns the number of stored routes
func (u *UpdateSender) RouteCount() int64 {
	log.Warningf("BGP Update Sender: RouteCount not implemented")
	return 0
}
