package server

/*
type devices struct {
	srv  *Server
	db   map[string]*dev
	dbMu sync.RWMutex
}

func newDevices(srv *Server) *devices {
	return &devices{
		srv: srv,
		db:  make(map[string]*dev),
	}
}

func (db *devices) addDevice(ifcfg *config.ISISInterfaceConfig) error {
	db.dbMu.Lock()
	defer db.dbMu.Unlock()

	if _, ok := db.db[ifcfg.Name]; ok {
		return fmt.Errorf("Interface exists already")
	}

	d := newDev(db.srv, ifcfg)
	db.db[ifcfg.Name] = d

	db.srv.ds.Subscribe(d, d.name)
	return nil
}

func (db *devices) removeDevice(name string) error {
	db.dbMu.Lock()
	defer db.dbMu.Unlock()

	if _, ok := db.db[name]; !ok {
		return fmt.Errorf("Interface not found")
	}

	db.srv.ds.Unsubscribe(db.db[name], name)
	err := db.db[name].disable()
	if err != nil {
		return errors.Wrap(err, "Unable to disable interface")
	}

	delete(db.db, name)
	return nil
}
*/
