package rt

type ClientManager struct {
	clients      map[RouteTableClient]struct{} // Ensures a client registers at most once
	routingTable *RT
}

func (c *ClientManager) Register(client RouteTableClient) {
	if c.clients == nil {
		c.clients = make(map[RouteTableClient]struct{}, 0)
	}
	c.clients[client] = struct{}{}
	c.routingTable.updateNewClient(client)
}

func (c *ClientManager) Unregister(client RouteTableClient) {
	if _, ok := c.clients[client]; !ok {
		return
	}
	delete(c.clients, client)
}
