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

func newNode(pfx *net.Prefix, path *route.Path, skip uint8, dummy bool) *node {
	n := &node{
		route: route.NewRoute(pfx, path),
		skip:  skip,
		dummy: dummy,
	}
	return n
}

func (n *node) removePath(pfx *net.Prefix, p *route.Path) (final bool) {
	if n == nil {
		return false
	}

	if n.route.Prefix().Equal(pfx) {
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

	b := pfx.Addr().BitAtPosition(n.route.Pfxlen() + 1)
	if !b {
		return n.l.removePath(pfx, p)
	}
	return n.h.removePath(pfx, p)
}

func (n *node) lpm(needle *net.Prefix, res *[]*route.Route) {
	if n == nil {
		return
	}

	currentPfx := n.route.Prefix()
	if currentPfx.Equal(needle) && !n.dummy {
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

func (n *node) get(pfx *net.Prefix) *node {
	if n == nil {
		return nil
	}

	if n.route.Prefix().Equal(pfx) {
		if n.dummy {
			return nil
		}
		return n
	}

	if n.route.Pfxlen() > pfx.Len() {
		return nil
	}

	b := pfx.Addr().BitAtPosition(n.route.Pfxlen() + 1)
	if !b {
		return n.l.get(pfx)
	}
	return n.h.get(pfx)
}

func (n *node) addPath(pfx *net.Prefix, p *route.Path) (*node, bool) {
	currentPfx := n.route.Prefix()
	if currentPfx.Equal(pfx) {
		n.route.AddPath(p)
		// Store previous dummy-ness to check if this node became new
		dummy := n.dummy
		n.dummy = false
		return n, dummy
	}

	// is pfx NOT a subnet of this node?
	if !currentPfx.Contains(pfx) {
		if pfx.Contains(currentPfx) {
			return n.insertBefore(pfx, p), true
		}

		return n.newSuperNode(pfx, p), true
	}

	// pfx is a subnet of this node
	b := pfx.Addr().BitAtPosition(n.route.Pfxlen() + 1)

	if !b {
		return n.insertLow(pfx, p, currentPfx.Len())
	}
	return n.insertHigh(pfx, p, n.route.Pfxlen())
}

func (n *node) insertLow(pfx *net.Prefix, p *route.Path, parentPfxLen uint8) (*node, bool) {
	if n.l == nil {
		n.l = newNode(pfx, p, pfx.Len()-parentPfxLen-1, false)
		return n, true
	}

	newRoot, isNew := n.l.addPath(pfx, p)
	n.l = newRoot
	return n, isNew
}

func (n *node) insertHigh(pfx *net.Prefix, p *route.Path, parentPfxLen uint8) (*node, bool) {
	if n.h == nil {
		n.h = newNode(pfx, p, pfx.Len()-parentPfxLen-1, false)
		return n, true
	}
	newRoot, isNew := n.h.addPath(pfx, p)
	n.h = newRoot
	return n, isNew
}

func (n *node) newSuperNode(pfx *net.Prefix, p *route.Path) *node {
	superNet := pfx.GetSupernet(n.route.Prefix())

	pfxLenDiff := n.route.Pfxlen() - superNet.Len()
	skip := n.skip - pfxLenDiff

	pseudoNode := newNode(&superNet, nil, skip, true)
	pseudoNode.insertChildren(n, pfx, p)
	return pseudoNode
}

func (n *node) insertChildren(old *node, newPfx *net.Prefix, newPath *route.Path) {
	// Place the old node
	b := old.route.Prefix().Addr().BitAtPosition(n.route.Pfxlen() + 1)
	if !b {
		n.l = old
		n.l.skip = old.route.Pfxlen() - n.route.Pfxlen() - 1
	} else {
		n.h = old
		n.h.skip = old.route.Pfxlen() - n.route.Pfxlen() - 1
	}

	// Place the new Prefix
	newNode := newNode(newPfx, newPath, newPfx.Len()-n.route.Pfxlen()-1, false)
	b = newPfx.Addr().BitAtPosition(n.route.Pfxlen() + 1)
	if !b {
		n.l = newNode
	} else {
		n.h = newNode
	}
}

func (n *node) insertBefore(pfx *net.Prefix, p *route.Path) *node {
	tmp := n

	pfxLenDiff := n.route.Pfxlen() - pfx.Len()
	skip := n.skip - pfxLenDiff
	new := newNode(pfx, p, skip, false)

	b := n.route.Prefix().Addr().BitAtPosition(pfx.Len() + 1)
	if !b {
		new.l = tmp
		new.l.skip = tmp.route.Pfxlen() - pfx.Len() - 1
	} else {
		new.h = tmp
		new.h.skip = tmp.route.Pfxlen() - pfx.Len() - 1
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
