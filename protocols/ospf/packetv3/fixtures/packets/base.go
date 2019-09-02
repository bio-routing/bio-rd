package packets

import "github.com/bio-routing/bio-rd/protocols/ospf/packetv3"

// Packets by source file
var Packets = make(map[string][]*packetv3.OSPFv3Message)
