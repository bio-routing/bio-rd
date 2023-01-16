# Redistribution of routes between routing tables

See the [Bio-Routing Routing table design](../user/routintable/overview.md) page for an overview of the internal architecture.

This document will shed some light on the implementation details of route redistribution between protocols inside Bio-Routing.

The central code point for propagating routes between routing tables and protocols, and therefore redistribution, is the `func (p *Path) CheckRedistribute(newPathType uint8) (*Path, bool)` method. It's task it to check if a path needs to be redistributed or is a protocol-native path with respect to the protocol (e.g. `Path.Type`) it should be propagated to. It makes sure consumers are operating of a copy of the path and have the freedom to manipulate path attributes as they wish (by means of import filters or protocol specific code paths), and will also return if this path has to be redistributed or is protocol-native.

The actual protocol-specific redistribution logic needs to be implemented in each protocol consuming prefixes via the `RouteTableClient` interface, namely the `AddPath()` and `AddPathInitialDump()` methods. The consumer needs to make sure to set up the protocol specific path attributes and dedup the path, if available for the given path type.

This should look like the following snippet:


    func (a *AdjRIBOut) AddPath(pfx *bnet.Prefix, p *route.Path) error {
  	    p, redist := p.CheckRedistribute(route.BGPPathType)
  	    if redist {
  	  	    err := a.redistributePath(p)
  	  	    if err != nil {
  	  	  	  return err
  	  	    }
  	    }

        ... any protocol specific checks, bells and whistels

        p, reject := a.exportFilterChain.Process(pfx, p)
	    if reject {
		     return nil
	    }

	    p.BGPPath = p.BGPPath.Dedup()