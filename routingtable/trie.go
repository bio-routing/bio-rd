package routingtable

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
)

// node is a node in the compressed trie that is used to implement a routing table
type node struct {
	skip  uint8
	dummy bool
	route *route.Route
	l     *node
	h     *node
}

func getBitUint32(x uint32, pos uint8) bool {
	return ((x) & (1 << (32 - pos))) != 0
}

func newNode(pfx net.Prefix, path *route.Path, skip uint8, dummy bool) *node {
	n := &node{
		route: route.NewRoute(pfx, path),
		skip:  skip,
		dummy: dummy,
	}
	return n
}

func (n *node) removePath(pfx net.Prefix, p *route.Path) (final bool) {
	if n == nil {
		return false
	}

	if n.route.Prefix() == pfx {
		if n.dummy {
			return
		}

		nPathsAfterDel := n.route.RemovePath(p)
		if len(n.route.Paths()) == 0 {
			// FIXME: Can this node actually be removed from the trie entirely?
			n.dummy = true
		}

		return nPathsAfterDel == 0
	}

	b := getBitUint32(pfx.Addr(), n.route.Pfxlen()+1)
	if !b {
		return n.l.removePath(pfx, p)
	}
	return n.h.removePath(pfx, p)
}

func (n *node) lpm(needle net.Prefix, res *[]*route.Route) {
	if n == nil {
		return
	}

	currentPfx := n.route.Prefix()
	if currentPfx == needle && !n.dummy {
		*res = append(*res, n.route)
		return
	}

	if !currentPfx.Contains(needle) {
		return
	}

	if !n.dummy {
		*res = append(*res, n.route)
	}
	n.l.lpm(needle, res)
	n.h.lpm(needle, res)
}

func (n *node) dumpPfxs(res []*route.Route) []*route.Route {
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

func (n *node) get(pfx net.Prefix) *node {
	if n == nil {
		return nil
	}

	if n.route.Prefix() == pfx {
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

func (n *node) addPath(pfx net.Prefix, p *route.Path) *node {
	currentPfx := n.route.Prefix()
	if currentPfx == pfx {
		n.route.AddPath(p)
		n.dummy = false
		return n
	}

	// is pfx NOT a subnet of this node?
	if !currentPfx.Contains(pfx) {
		if pfx.Contains(currentPfx) {
			return n.insertBefore(pfx, p, n.route.Pfxlen()-n.skip-1)
		}

		return n.newSuperNode(pfx, p)
	}

	// pfx is a subnet of this node
	b := getBitUint32(pfx.Addr(), n.route.Pfxlen()+1)
	if !b {
		return n.insertLow(pfx, p, currentPfx.Pfxlen())
	}
	return n.insertHigh(pfx, p, n.route.Pfxlen())
}

func (n *node) insertLow(pfx net.Prefix, p *route.Path, parentPfxLen uint8) *node {
	if n.l == nil {
		n.l = newNode(pfx, p, pfx.Pfxlen()-parentPfxLen-1, false)
		return n
	}
	n.l = n.l.addPath(pfx, p)
	return n
}

func (n *node) insertHigh(pfx net.Prefix, p *route.Path, parentPfxLen uint8) *node {
	if n.h == nil {
		n.h = newNode(pfx, p, pfx.Pfxlen()-parentPfxLen-1, false)
		return n
	}
	n.h = n.h.addPath(pfx, p)
	return n
}

func (n *node) newSuperNode(pfx net.Prefix, p *route.Path) *node {
	superNet := pfx.GetSupernet(n.route.Prefix())

	pfxLenDiff := n.route.Pfxlen() - superNet.Pfxlen()
	skip := n.skip - pfxLenDiff

	pseudoNode := newNode(superNet, nil, skip, true)
	pseudoNode.insertChildren(n, pfx, p)
	return pseudoNode
}

func (n *node) insertChildren(old *node, newPfx net.Prefix, newPath *route.Path) {
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
	newNode := newNode(newPfx, newPath, newPfx.Pfxlen()-n.route.Pfxlen()-1, false)
	b = getBitUint32(newPfx.Addr(), n.route.Pfxlen()+1)
	if !b {
		n.l = newNode
	} else {
		n.h = newNode
	}
}

func (n *node) insertBefore(pfx net.Prefix, p *route.Path, parentPfxLen uint8) *node {
	tmp := n

	pfxLenDiff := n.route.Pfxlen() - pfx.Pfxlen()
	skip := n.skip - pfxLenDiff
	new := newNode(pfx, p, skip, false)

	b := getBitUint32(pfx.Addr(), parentPfxLen)
	if !b {
		new.l = tmp
		new.l.skip = tmp.route.Pfxlen() - pfx.Pfxlen() - 1
	} else {
		new.h = tmp
		new.h.skip = tmp.route.Pfxlen() - pfx.Pfxlen() - 1
	}

	return new
}

func (n *node) dump(res []*route.Route) []*route.Route {
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
