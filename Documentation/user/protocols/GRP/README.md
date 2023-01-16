# gRPC Routing Protocol (GRP)

This protocol is designed to allow leveraging Bio-Routing as a proxy between classic routing protocols like BGP, IS-IS, and OSPF and an gRPC based controller.
Bio-Routing can be used to learn routes via gRPC and export them into the desired protocols, learn routes from classic protocols and export them via gRPC, or both.

The gRPC Routing Protocol follows the design principle for BGP as much as is practical and borrows the concept of AdjRIBIn and AdjRIBOut including the possibility to define import and export filters.

To be most flexible the gRPC Routing Protocol allows learning and/or exporting routes for multiple VRFs over one gRPC connection.

Bio-Routing, at least for now, only implements a gRPC Routing Protocol client componenent, meaning that the user has to run a GRP server by themselves. An example can be found in the examples/grp/ folder inside this repository.

The following diagram shows the architectural overview:

![gRPC Routing Protocol architecural overview](GRP Architecture.drawio.svg)