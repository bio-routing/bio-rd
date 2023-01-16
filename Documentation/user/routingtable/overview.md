# Bio-Routing routing table design

Bio-Routing is following the "one RIB model" like Junos and bird (compared to the protocol specific RIBs in the Cisco world).

This means that routes/paths from all protocols will end up in the same LocRIB (per VRF) and will be propagated to all protocols registered with the LocRIB, if the protocol's import filter does not reject the prefix.

The following diagram visualizes this architecture and the interactions between the different routing tables:

![Routing table interaction diagram](RIB-FIB.drawio.svg "Routing table interaction diagram")

The import/export filters are represented by a filter chain (`filter.Chain`), which could look like this to prevent static routes from being redistributed into BGP for example:

    filter.Chain{
            filter.NewFilter(
                    "Accept all but static routes",
                    []*filter.Term{
                            filter.NewTerm(
                                    "Ignore static routes",
                                    []*filter.TermCondition{
                                            filter.NewTermConditionWithProtocols(route.StaticPathType),
                                    },
                                    []actions.Action{
                                            actions.NewRejectAction(),
                                    },
                            ),
                            filter.NewTerm(
                                    "accept",
                                    []*filter.TermCondition{},
                                    []actions.Action{
                                            actions.NewAcceptAction(),
                                    }),
                    }),
    },


