package server

import (
	"github.com/bio-routing/bio-rd/protocols/ospfv3/packet"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// handles Area 0 only
type backboneOSPFServer struct {
	routerID packet.ID
	backbone *areaManager

	log logrus.FieldLogger
}

func NewBackboneOSPFServer(routerID packet.ID) (*backboneOSPFServer, error) {
	log := logrus.StandardLogger().WithField("server", "OSPFv3")
	area, err := newAreaManager(
		log.WithField("component", "areaManager"),
		AreaConfig{routerID: routerID},
		routingtable.NewRoutingTable(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create backbone area")
	}

	srv := &backboneOSPFServer{
		routerID: routerID,
		backbone: area,
		log:      log.WithField("component", "server"),
	}

	return srv, nil
}
