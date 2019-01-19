package server

import (
	"fmt"
	"sync"

	"github.com/bio-routing/bio-rd/config"
	"github.com/pkg/errors"
)

type devicesManager struct {
	srv  *Server
	db   []*dev
	dbMu sync.RWMutex
}

func newDevicesManager(srv *Server) *devicesManager {
	return &devicesManager{
		srv: srv,
		db:  make([]*dev, 0),
	}
}

func (dm *devicesManager) addDevice(ifcfg *config.ISISInterfaceConfig) error {
	dm.dbMu.Lock()
	defer dm.dbMu.Unlock()

	for i := range dm.db {
		if dm.db[i].name == ifcfg.Name {
			return fmt.Errorf("Interface exists already")
		}
	}

	d := newDev(dm.srv, ifcfg)
	dm.db = append(dm.db, d)
	dm.srv.ds.Subscribe(d, d.name)
	return nil
}

func (dm *devicesManager) removeDevice(name string) error {
	dm.dbMu.Lock()
	defer dm.dbMu.Unlock()

	for i := range dm.db {
		if dm.db[i].name != name {
			continue
		}

		dm.srv.ds.Unsubscribe(dm.db[i], name)
		err := dm.db[i].disable()
		if err != nil {
			return errors.Wrap(err, "Unable to disable interface")
		}

		dm.db = append(dm.db[:i], dm.db[i+1:]...)
		return nil
	}

	return fmt.Errorf("Device %q not found", name)
}
