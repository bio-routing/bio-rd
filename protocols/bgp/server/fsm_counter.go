package server

type fsmConters struct {
	updatesReceived uint64
	updatesSent     uint64
}

func (c *fsmConters) reset() {
	c.updatesReceived = 0
	c.updatesSent = 0
}
