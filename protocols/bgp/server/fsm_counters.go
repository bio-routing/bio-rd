package server

type fsmCounters struct {
	updatesReceived uint64
	updatesSent     uint64
}

func (c *fsmCounters) reset() {
	c.updatesReceived = 0
	c.updatesSent = 0
}
