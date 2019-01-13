package server

type devices struct {
	db map[string]*dev
}

func newDevices() *devices {
	return &devices{}
}
