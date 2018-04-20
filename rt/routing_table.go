package rt

import (
	"github.com/bio-routing/bio-rd/net"
)

type RT struct {
	root  *node
	nodes uint64
}

type node struct {
	skip  uint8
	dummy bool
	route *Route
	l     *node
	h     *node
}

// New creates a new empty LPM
func New() *RT {
	return &RT{}
}

func newNode(route *Route, skip uint8, dummy bool) *node {
	n := &node{
		route: route,
		skip:  skip,
		dummy: dummy,
	}
	return n
}

// LPM performs a longest prefix match for pfx on lpm
func (lpm *RT) LPM(pfx *net.Prefix) (res []*Route) {
	if lpm.root == nil {
		return nil
	}

	lpm.root.lpm(pfx, &res)
	return res
}

// RemovePath removes a path from the trie
func (lpm *RT) RemovePath(route *Route) {
	lpm.root.removePath(route)
}

func (lpm *RT) RemovePfx(pfx *net.Prefix) {
	lpm.root.removePfx(pfx)
}

// Get get's prefix pfx from the LPM
func (lpm *RT) Get(pfx *net.Prefix, moreSpecifics bool) (res []*Route) {
	if lpm.root == nil {
		return nil
	}

	node := lpm.root.get(pfx)
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

// Insert inserts a route into the LPM
func (lpm *RT) Insert(route *Route) {
	if lpm.root == nil {
		lpm.root = newNode(route, route.Pfxlen(), false)
		return
	}

	lpm.root = lpm.root.insert(route)
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

func (n *node) removePfx(pfx *net.Prefix) {
	if n == nil {
		return
	}

	if *n.route.Prefix() == *pfx {
		if n.dummy {
			return
		}

		n.dummy = true

		return
	}

	b := getBitUint32(pfx.Addr(), n.route.Pfxlen()+1)
	if !b {
		n.l.removePfx(pfx)
		return
	}
	n.h.removePfx(pfx)
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
		n.l = newNode(route, route.Pfxlen()-parentPfxLen-1, false)
		return n
	}
	n.l = n.l.insert(route)
	return n
}

func (n *node) insertHigh(route *Route, parentPfxLen uint8) *node {
	if n.h == nil {
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

func (lpm *RT) Dump() []*Route {
	res := make([]*Route, 0)
	return lpm.root.dump(res)
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
