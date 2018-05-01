package rt

import (
	"fmt"
	"sync"

	"github.com/bio-routing/bio-rd/net"
)

type RouteTableClient interface {
	AddPath(*Route)
	ReplaceRoute(*Route)
	RemovePath(*Route)
	RemoveRoute(*net.Prefix)
}

// RT represents a routing table
type RT struct {
	root            *node
	selectBestPaths bool
	mu              sync.RWMutex
	ClientManager
}

// node is a node in the compressed trie that is used to implement a routing table
type node struct {
	skip  uint8
	dummy bool
	route *Route
	l     *node
	h     *node
}

// New creates a new empty LPM
func New(selectBestPaths bool) *RT {
	rt := &RT{
		selectBestPaths: selectBestPaths,
	}

	rt.ClientManager.routingTable = rt
	return rt
}

func (rt *RT) updateNewClient(client RouteTableClient) {
	rt.root.updateNewClient(client)
}

func (n *node) updateNewClient(client RouteTableClient) {
	if n == nil {
		return
	}

	if !n.dummy {
		client.AddPath(n.route)
	}

	n.l.updateNewClient(client)
	n.h.updateNewClient(client)
}

func (rt *RT) updatePathSelection(route *Route) {
	fmt.Printf("updatePathSelection\n")
	formerActivePaths := rt.Get(route.pfx, false)[0].activePaths
	activePaths := rt.selectBestPath(route).route.activePaths

	fmt.Printf("formerActivePaths: %v\n", formerActivePaths)
	fmt.Printf("activePaths: %v\n", activePaths)
	x := missingPaths(activePaths, formerActivePaths)
	fmt.Printf("x: %v\n", x)

	for _, advertisePath := range x {
		adervtiseRoute := &Route{
			pfx:   route.pfx,
			paths: []*Path{advertisePath},
		}
		for client := range rt.clients {
			fmt.Printf("Propagating an Update via AddPath()")
			client.AddPath(adervtiseRoute)
		}
	}

	for _, withdrawPath := range missingPaths(formerActivePaths, activePaths) {
		withdrawRoute := &Route{
			pfx:   route.pfx,
			paths: []*Path{withdrawPath},
		}
		for client := range rt.clients {
			fmt.Printf("Propagating an Update via RemovePath()")
			client.RemovePath(withdrawRoute)
		}
	}
}

func (rt *RT) AddPath(route *Route) {
	rt.addPath(route)
	if rt.selectBestPaths {
		rt.updatePathSelection(route)
		return
	}

	for client := range rt.clients {
		client.AddPath(route)
	}
}

// RemovePath removes a path from the trie
func (rt *RT) RemovePath(route *Route) {
	rt.root.removePath(route)

	for client := range rt.clients {
		client.RemovePath(route)
	}
}

// RemoveRoute removes a prefix from the rt including all it's paths
func (rt *RT) RemoveRoute(pfx *net.Prefix) {
	if rt.selectBestPaths {
		return
	}

	r := rt.Get(pfx, false)
	if len(r) == 0 {
		return
	}

	for client := range rt.clients {
		for _, path := range r[0].paths {
			withdrawRoute := NewRoute(pfx, []*Path{path})
			client.RemovePath(withdrawRoute)
		}
	}

	rt.root.removeRoute(pfx)
}

// ReplaceRoute replaces all paths of a route. Route is added if it doesn't exist yet.
func (rt *RT) ReplaceRoute(route *Route) {
	fmt.Printf("Replacing a route!\n")
	if rt.selectBestPaths {
		fmt.Printf("Ignoring because rt.selectBestPaths is false\n")
		return
	}

	r := rt.Get(route.pfx, false)
	if len(r) > 0 {
		rt.RemoveRoute(route.pfx)
	}
	rt.addPath(route)
	rt.updatePathSelection(route)
}

// LPM performs a longest prefix match for pfx on lpm
func (rt *RT) LPM(pfx *net.Prefix) (res []*Route) {
	if rt.root == nil {
		return nil
	}

	rt.root.lpm(pfx, &res)
	return res
}

// Get get's prefix pfx from the LPM
func (rt *RT) Get(pfx *net.Prefix, moreSpecifics bool) (res []*Route) {
	if rt.root == nil {
		return nil
	}

	node := rt.root.get(pfx)
	if moreSpecifics {
		return node.dumpPfxs(res)
	}

	if node == nil {
		return nil
	}

	return []*Route{
		node.route,
	}
}

func (rt *RT) addPath(route *Route) {
	if rt.root == nil {
		rt.root = newNode(route, route.Pfxlen(), false)
		return
	}

	rt.root = rt.root.insert(route)
}

func (rt *RT) selectBestPath(route *Route) *node {
	if rt.root == nil {
		return nil
	}

	node := rt.root.get(route.pfx)
	if !rt.selectBestPaths {
		// If we don't select best path(s) evey path is best path
		node.route.activePaths = make([]*Path, len(node.route.paths))
		copy(node.route.activePaths, node.route.paths)
		return node
	}

	node.route.bestPaths()
	return node
}

func newNode(route *Route, skip uint8, dummy bool) *node {
	n := &node{
		route: route,
		skip:  skip,
		dummy: dummy,
	}
	return n
}

func (n *node) removePath(route *Route) {
	if n == nil {
		return
	}

	if *n.route.Prefix() == *route.Prefix() {
		if n.dummy {
			return
		}

		if n.route.Remove(route) {
			// FIXME: Can this node actually be removed from the trie entirely?
			n.dummy = true
		}

		return
	}

	b := getBitUint32(route.Prefix().Addr(), n.route.Pfxlen()+1)
	if !b {
		n.l.removePath(route)
		return
	}
	n.h.removePath(route)
	return
}

func (n *node) replaceRoute(route *Route) {
	if n == nil {
		return
	}

	if *n.route.Prefix() == *route.Prefix() {
		n.route = route
		n.dummy = false
		return
	}

	b := getBitUint32(route.Prefix().Addr(), n.route.Pfxlen()+1)
	if !b {
		n.l.replaceRoute(route)
		return
	}
	n.h.replaceRoute(route)
	return
}

func (n *node) removeRoute(pfx *net.Prefix) {
	if n == nil {
		return
	}

	if *n.route.Prefix() == *pfx {
		if n.dummy {
			return
		}

		// TODO: Remove node if possible
		n.dummy = true
		return
	}

	b := getBitUint32(pfx.Addr(), n.route.Pfxlen()+1)
	if !b {
		n.l.removeRoute(pfx)
		return
	}
	n.h.removeRoute(pfx)
	return
}

func (n *node) lpm(needle *net.Prefix, res *[]*Route) {
	if n == nil {
		return
	}

	if *n.route.Prefix() == *needle && !n.dummy {
		*res = append(*res, n.route)
		return
	}

	if !n.route.Prefix().Contains(needle) {
		return
	}

	if !n.dummy {
		*res = append(*res, n.route)
	}
	n.l.lpm(needle, res)
	n.h.lpm(needle, res)
}

func (n *node) dumpPfxs(res []*Route) []*Route {
	if n == nil {
		return nil
	}

	if !n.dummy {
		res = append(res, n.route)
	}

	if n.l != nil {
		res = n.l.dumpPfxs(res)
	}

	if n.h != nil {
		res = n.h.dumpPfxs(res)
	}

	return res
}

func (n *node) get(pfx *net.Prefix) *node {
	if n == nil {
		return nil
	}

	if *n.route.Prefix() == *pfx {
		if n.dummy {
			return nil
		}
		return n
	}

	if n.route.Pfxlen() > pfx.Pfxlen() {
		return nil
	}

	b := getBitUint32(pfx.Addr(), n.route.Pfxlen()+1)
	if !b {
		return n.l.get(pfx)
	}
	return n.h.get(pfx)
}

func (n *node) insert(route *Route) *node {
	if *n.route.Prefix() == *route.Prefix() {
		n.route.AddPaths(route.paths)
		n.dummy = false
		return n
	}

	// is pfx NOT a subnet of this node?
	if !n.route.Prefix().Contains(route.Prefix()) {
		route.bestPaths()
		if route.Prefix().Contains(n.route.Prefix()) {
			return n.insertBefore(route, n.route.Pfxlen()-n.skip-1)
		}

		return n.newSuperNode(route)
	}

	// pfx is a subnet of this node
	b := getBitUint32(route.Prefix().Addr(), n.route.Pfxlen()+1)
	if !b {
		return n.insertLow(route, n.route.Prefix().Pfxlen())
	}
	return n.insertHigh(route, n.route.Pfxlen())
}

func (n *node) insertLow(route *Route, parentPfxLen uint8) *node {
	if n.l == nil {
		route.bestPaths()
		n.l = newNode(route, route.Pfxlen()-parentPfxLen-1, false)
		return n
	}
	n.l = n.l.insert(route)
	return n
}

func (n *node) insertHigh(route *Route, parentPfxLen uint8) *node {
	if n.h == nil {
		route.bestPaths()
		n.h = newNode(route, route.Pfxlen()-parentPfxLen-1, false)
		return n
	}
	n.h = n.h.insert(route)
	return n
}

func (n *node) newSuperNode(route *Route) *node {
	superNet := route.Prefix().GetSupernet(n.route.Prefix())

	pfxLenDiff := n.route.Pfxlen() - superNet.Pfxlen()
	skip := n.skip - pfxLenDiff

	pseudoNode := newNode(NewRoute(superNet, nil), skip, true)
	pseudoNode.insertChildren(n, route)
	return pseudoNode
}

func (n *node) insertChildren(old *node, new *Route) {
	// Place the old node
	b := getBitUint32(old.route.Prefix().Addr(), n.route.Pfxlen()+1)
	if !b {
		n.l = old
		n.l.skip = old.route.Pfxlen() - n.route.Pfxlen() - 1
	} else {
		n.h = old
		n.h.skip = old.route.Pfxlen() - n.route.Pfxlen() - 1
	}

	// Place the new Prefix
	newNode := newNode(new, new.Pfxlen()-n.route.Pfxlen()-1, false)
	b = getBitUint32(new.Prefix().Addr(), n.route.Pfxlen()+1)
	if !b {
		n.l = newNode
	} else {
		n.h = newNode
	}
}

func (n *node) insertBefore(route *Route, parentPfxLen uint8) *node {
	tmp := n

	pfxLenDiff := n.route.Pfxlen() - route.Pfxlen()
	skip := n.skip - pfxLenDiff
	new := newNode(route, skip, false)

	b := getBitUint32(route.Prefix().Addr(), parentPfxLen)
	if !b {
		new.l = tmp
		new.l.skip = tmp.route.Pfxlen() - route.Pfxlen() - 1
	} else {
		new.h = tmp
		new.h.skip = tmp.route.Pfxlen() - route.Pfxlen() - 1
	}

	return new
}

// Dump dumps all routes in table rt into a slice
func (rt *RT) Dump() []*Route {
	res := make([]*Route, 0)
	return rt.root.dump(res)
}

func (n *node) dump(res []*Route) []*Route {
	if n == nil {
		return res
	}

	if !n.dummy {
		res = append(res, n.route)
	}

	res = n.l.dump(res)
	res = n.h.dump(res)
	return res
}

func getBitUint32(x uint32, pos uint8) bool {
	return ((x) & (1 << (32 - pos))) != 0
}
