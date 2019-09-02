
// GENERATED FILE - do not edit!
// to regenerate this, run "go run ./protocols/ospf/packetv3/fixtures/packets/gen/"

package packets

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/ospf/packetv3"
)

func init() {
	filePkts := make([]*packetv3.OSPFv3Message, 38)
	filePkts[0] = packet_OSPFv3_broadcast_adjacency_001()
	filePkts[1] = packet_OSPFv3_broadcast_adjacency_002()
	filePkts[2] = packet_OSPFv3_broadcast_adjacency_003()
	filePkts[3] = packet_OSPFv3_broadcast_adjacency_004()
	filePkts[4] = packet_OSPFv3_broadcast_adjacency_005()
	filePkts[5] = packet_OSPFv3_broadcast_adjacency_006()
	filePkts[6] = packet_OSPFv3_broadcast_adjacency_007()
	filePkts[7] = packet_OSPFv3_broadcast_adjacency_008()
	filePkts[8] = packet_OSPFv3_broadcast_adjacency_009()
	filePkts[9] = packet_OSPFv3_broadcast_adjacency_010()
	filePkts[10] = packet_OSPFv3_broadcast_adjacency_011()
	filePkts[11] = packet_OSPFv3_broadcast_adjacency_012()
	filePkts[12] = packet_OSPFv3_broadcast_adjacency_013()
	filePkts[13] = packet_OSPFv3_broadcast_adjacency_014()
	filePkts[14] = packet_OSPFv3_broadcast_adjacency_015()
	filePkts[15] = packet_OSPFv3_broadcast_adjacency_016()
	filePkts[16] = packet_OSPFv3_broadcast_adjacency_017()
	filePkts[17] = packet_OSPFv3_broadcast_adjacency_018()
	filePkts[18] = packet_OSPFv3_broadcast_adjacency_019()
	filePkts[19] = packet_OSPFv3_broadcast_adjacency_020()
	filePkts[20] = packet_OSPFv3_broadcast_adjacency_021()
	filePkts[21] = packet_OSPFv3_broadcast_adjacency_022()
	filePkts[22] = packet_OSPFv3_broadcast_adjacency_023()
	filePkts[23] = packet_OSPFv3_broadcast_adjacency_024()
	filePkts[24] = packet_OSPFv3_broadcast_adjacency_025()
	filePkts[25] = packet_OSPFv3_broadcast_adjacency_026()
	filePkts[26] = packet_OSPFv3_broadcast_adjacency_027()
	filePkts[27] = packet_OSPFv3_broadcast_adjacency_028()
	filePkts[28] = packet_OSPFv3_broadcast_adjacency_029()
	filePkts[29] = packet_OSPFv3_broadcast_adjacency_030()
	filePkts[30] = packet_OSPFv3_broadcast_adjacency_031()
	filePkts[31] = packet_OSPFv3_broadcast_adjacency_032()
	filePkts[32] = packet_OSPFv3_broadcast_adjacency_033()
	filePkts[33] = packet_OSPFv3_broadcast_adjacency_034()
	filePkts[34] = packet_OSPFv3_broadcast_adjacency_035()
	filePkts[35] = packet_OSPFv3_broadcast_adjacency_036()
	filePkts[36] = packet_OSPFv3_broadcast_adjacency_037()
	filePkts[37] = packet_OSPFv3_broadcast_adjacency_038()
	Packets["OSPFv3_broadcast_adjacency.cap"] = filePkts
}

func packet_OSPFv3_broadcast_adjacency_001() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x1, PacketLength:0x24, RouterID:0x1010101, AreaID:0x1, Checksum:0xfb86, InstanceID:0x0, Body:nil}
	body := &packetv3.Hello{InterfaceID:0x5, RouterPriority:0x1, Options:packetv3.RouterOptions{Flags:0x13}, HelloInterval:0xa, RouterDeadInterval:0x28, DesignatedRouterID:0x0, BackupDesignatedRouterID:0x0, Neighbors:[]packetv3.ID(nil)}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_002() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x1, PacketLength:0x24, RouterID:0x1010101, AreaID:0x1, Checksum:0xfb86, InstanceID:0x0, Body:nil}
	body := &packetv3.Hello{InterfaceID:0x5, RouterPriority:0x1, Options:packetv3.RouterOptions{Flags:0x13}, HelloInterval:0xa, RouterDeadInterval:0x28, DesignatedRouterID:0x0, BackupDesignatedRouterID:0x0, Neighbors:[]packetv3.ID(nil)}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_003() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x1, PacketLength:0x24, RouterID:0x1010101, AreaID:0x1, Checksum:0xfb86, InstanceID:0x0, Body:nil}
	body := &packetv3.Hello{InterfaceID:0x5, RouterPriority:0x1, Options:packetv3.RouterOptions{Flags:0x13}, HelloInterval:0xa, RouterDeadInterval:0x28, DesignatedRouterID:0x0, BackupDesignatedRouterID:0x0, Neighbors:[]packetv3.ID(nil)}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_004() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x1, PacketLength:0x24, RouterID:0x1010101, AreaID:0x1, Checksum:0xfb86, InstanceID:0x0, Body:nil}
	body := &packetv3.Hello{InterfaceID:0x5, RouterPriority:0x1, Options:packetv3.RouterOptions{Flags:0x13}, HelloInterval:0xa, RouterDeadInterval:0x28, DesignatedRouterID:0x0, BackupDesignatedRouterID:0x0, Neighbors:[]packetv3.ID(nil)}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_005() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x1, PacketLength:0x24, RouterID:0x2020202, AreaID:0x1, Checksum:0xf983, InstanceID:0x0, Body:nil}
	body := &packetv3.Hello{InterfaceID:0x5, RouterPriority:0x1, Options:packetv3.RouterOptions{Flags:0x13}, HelloInterval:0xa, RouterDeadInterval:0x28, DesignatedRouterID:0x0, BackupDesignatedRouterID:0x0, Neighbors:[]packetv3.ID(nil)}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_006() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x1, PacketLength:0x28, RouterID:0x1010101, AreaID:0x1, Checksum:0xf578, InstanceID:0x0, Body:nil}
	body := &packetv3.Hello{InterfaceID:0x5, RouterPriority:0x1, Options:packetv3.RouterOptions{Flags:0x13}, HelloInterval:0xa, RouterDeadInterval:0x28, DesignatedRouterID:0x1010101, BackupDesignatedRouterID:0x0, Neighbors:[]packetv3.ID{0x2020202}}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_007() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x2, PacketLength:0x1c, RouterID:0x2020202, AreaID:0x1, Checksum:0xd826, InstanceID:0x0, Body:nil}
	body := &packetv3.DatabaseDescription{Options:packetv3.RouterOptions{Flags:0x13}, InterfaceMTU:0x5dc, DBFlags:0x7, DDSequenceNumber:0x1d46, LSAHeaders:[]*packetv3.LSA(nil)}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_008() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x2, PacketLength:0x1c, RouterID:0x1010101, AreaID:0x1, Checksum:0xd342, InstanceID:0x0, Body:nil}
	body := &packetv3.DatabaseDescription{Options:packetv3.RouterOptions{Flags:0x13}, InterfaceMTU:0x5dc, DBFlags:0x7, DDSequenceNumber:0x242c, LSAHeaders:[]*packetv3.LSA(nil)}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_009() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x2, PacketLength:0xa8, RouterID:0x1010101, AreaID:0x1, Checksum:0x79f6, InstanceID:0x0, Body:nil}
	body := &packetv3.DatabaseDescription{Options:packetv3.RouterOptions{Flags:0x13}, InterfaceMTU:0x5dc, DBFlags:0x2, DDSequenceNumber:0x1d46, LSAHeaders:[]*packetv3.LSA{nil, nil, nil, nil, nil, nil, nil}}
	body.LSAHeaders = make([]*packetv3.LSA, 7)
	body.LSAHeaders[0] = &packetv3.LSA{Age:0x27, Type:0x2001, ID:0x0, AdvertisingRouter:0x1010101, SequenceNumber:0x80000002, Checksum:0xd13a, Length:0x18, Body:packetv3.Serializable(nil)}
	body.LSAHeaders[1] = &packetv3.LSA{Age:0x28, Type:0x2003, ID:0x0, AdvertisingRouter:0x1010101, SequenceNumber:0x80000001, Checksum:0xebd, Length:0x24, Body:packetv3.Serializable(nil)}
	body.LSAHeaders[2] = &packetv3.LSA{Age:0x28, Type:0x2003, ID:0x1, AdvertisingRouter:0x1010101, SequenceNumber:0x80000001, Checksum:0xeba0, Length:0x24, Body:packetv3.Serializable(nil)}
	body.LSAHeaders[3] = &packetv3.LSA{Age:0x28, Type:0x2003, ID:0x2, AdvertisingRouter:0x1010101, SequenceNumber:0x80000001, Checksum:0xbaf6, Length:0x24, Body:packetv3.Serializable(nil)}
	body.LSAHeaders[4] = &packetv3.LSA{Age:0x28, Type:0x2003, ID:0x3, AdvertisingRouter:0x1010101, SequenceNumber:0x80000001, Checksum:0x6259, Length:0x24, Body:packetv3.Serializable(nil)}
	body.LSAHeaders[5] = &packetv3.LSA{Age:0x22, Type:0x8, ID:0x5, AdvertisingRouter:0x1010101, SequenceNumber:0x80000002, Checksum:0x3d08, Length:0x38, Body:packetv3.Serializable(nil)}
	body.LSAHeaders[6] = &packetv3.LSA{Age:0x22, Type:0x2009, ID:0x0, AdvertisingRouter:0x1010101, SequenceNumber:0x80000001, Checksum:0xe8d2, Length:0x2c, Body:packetv3.Serializable(nil)}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_010() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x2, PacketLength:0x94, RouterID:0x2020202, AreaID:0x1, Checksum:0x7867, InstanceID:0x0, Body:nil}
	body := &packetv3.DatabaseDescription{Options:packetv3.RouterOptions{Flags:0x13}, InterfaceMTU:0x5dc, DBFlags:0x3, DDSequenceNumber:0x1d47, LSAHeaders:[]*packetv3.LSA{nil, nil, nil, nil, nil, nil}}
	body.LSAHeaders = make([]*packetv3.LSA, 6)
	body.LSAHeaders[0] = &packetv3.LSA{Age:0x4, Type:0x2001, ID:0x0, AdvertisingRouter:0x2020202, SequenceNumber:0x80000002, Checksum:0xb354, Length:0x18, Body:packetv3.Serializable(nil)}
	body.LSAHeaders[1] = &packetv3.LSA{Age:0x5, Type:0x2003, ID:0x0, AdvertisingRouter:0x2020202, SequenceNumber:0x80000001, Checksum:0xefd7, Length:0x24, Body:packetv3.Serializable(nil)}
	body.LSAHeaders[2] = &packetv3.LSA{Age:0x5, Type:0x2003, ID:0x1, AdvertisingRouter:0x2020202, SequenceNumber:0x80000001, Checksum:0xcdba, Length:0x24, Body:packetv3.Serializable(nil)}
	body.LSAHeaders[3] = &packetv3.LSA{Age:0x5, Type:0x2003, ID:0x2, AdvertisingRouter:0x2020202, SequenceNumber:0x80000001, Checksum:0x9c11, Length:0x24, Body:packetv3.Serializable(nil)}
	body.LSAHeaders[4] = &packetv3.LSA{Age:0x5, Type:0x2003, ID:0x3, AdvertisingRouter:0x2020202, SequenceNumber:0x80000001, Checksum:0x4473, Length:0x24, Body:packetv3.Serializable(nil)}
	body.LSAHeaders[5] = &packetv3.LSA{Age:0x4, Type:0x8, ID:0x5, AdvertisingRouter:0x2020202, SequenceNumber:0x80000001, Checksum:0x5433, Length:0x2c, Body:packetv3.Serializable(nil)}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_011() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x2, PacketLength:0x1c, RouterID:0x1010101, AreaID:0x1, Checksum:0xda2e, InstanceID:0x0, Body:nil}
	body := &packetv3.DatabaseDescription{Options:packetv3.RouterOptions{Flags:0x13}, InterfaceMTU:0x5dc, DBFlags:0x0, DDSequenceNumber:0x1d47, LSAHeaders:[]*packetv3.LSA(nil)}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_012() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x3, PacketLength:0x64, RouterID:0x2020202, AreaID:0x1, Checksum:0x2c9a, InstanceID:0x0, Body:nil}
	body := &packetv3.LinkStateRequestMsg{Requests:[]packetv3.LinkStateRequest{packetv3.LinkStateRequest{LSType:0x2001, LinkStateID:0x0, AdvertisingRouter:0x1010101}, packetv3.LinkStateRequest{LSType:0x2003, LinkStateID:0x3, AdvertisingRouter:0x1010101}, packetv3.LinkStateRequest{LSType:0x2003, LinkStateID:0x2, AdvertisingRouter:0x1010101}, packetv3.LinkStateRequest{LSType:0x2003, LinkStateID:0x1, AdvertisingRouter:0x1010101}, packetv3.LinkStateRequest{LSType:0x2003, LinkStateID:0x0, AdvertisingRouter:0x1010101}, packetv3.LinkStateRequest{LSType:0x8, LinkStateID:0x5, AdvertisingRouter:0x1010101}, packetv3.LinkStateRequest{LSType:0x2009, LinkStateID:0x0, AdvertisingRouter:0x1010101}}}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_013() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x3, PacketLength:0x58, RouterID:0x1010101, AreaID:0x1, Checksum:0x44b3, InstanceID:0x0, Body:nil}
	body := &packetv3.LinkStateRequestMsg{Requests:[]packetv3.LinkStateRequest{packetv3.LinkStateRequest{LSType:0x2001, LinkStateID:0x0, AdvertisingRouter:0x2020202}, packetv3.LinkStateRequest{LSType:0x2003, LinkStateID:0x3, AdvertisingRouter:0x2020202}, packetv3.LinkStateRequest{LSType:0x2003, LinkStateID:0x2, AdvertisingRouter:0x2020202}, packetv3.LinkStateRequest{LSType:0x2003, LinkStateID:0x1, AdvertisingRouter:0x2020202}, packetv3.LinkStateRequest{LSType:0x2003, LinkStateID:0x0, AdvertisingRouter:0x2020202}, packetv3.LinkStateRequest{LSType:0x8, LinkStateID:0x5, AdvertisingRouter:0x2020202}}}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_014() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x2, PacketLength:0x1c, RouterID:0x2020202, AreaID:0x1, Checksum:0xd82a, InstanceID:0x0, Body:nil}
	body := &packetv3.DatabaseDescription{Options:packetv3.RouterOptions{Flags:0x13}, InterfaceMTU:0x5dc, DBFlags:0x1, DDSequenceNumber:0x1d48, LSAHeaders:[]*packetv3.LSA(nil)}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_015() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x4, PacketLength:0x120, RouterID:0x1010101, AreaID:0x1, Checksum:0xe556, InstanceID:0x0, Body:nil}
	body := &packetv3.LinkStateUpdate{LSAs:[]*packetv3.LSA{nil, nil, nil, nil, nil, nil, nil}}
	body.LSAs = make([]*packetv3.LSA, 7)
	body.LSAs[0] = &packetv3.LSA{Age:0x28, Type:0x2001, ID:0x0, AdvertisingRouter:0x1010101, SequenceNumber:0x80000002, Checksum:0xd13a, Length:0x18, Body:&packetv3.RouterLSA{Flags:0x1, Options:packetv3.RouterOptions{Flags:0x33}, LinkDescriptions:[]packetv3.AreaLinkDescription(nil)}}
	body.LSAs[1] = &packetv3.LSA{Age:0x29, Type:0x2003, ID:0x3, AdvertisingRouter:0x1010101, SequenceNumber:0x80000001, Checksum:0x6259, Length:0x24, Body:&packetv3.InterAreaPrefixLSA{Metric:packetv3.InterfaceMetric{High:0x0, Low:0x4a}, Prefix:packetv3.LSAPrefix{PrefixLength:0x40, Options:packetv3.PrefixOptions{NoUnicast:false, LocalAddress:false, Propagate:false, DN:false}, Special:0x0, Address:net.IPv6(0x20010db800000003, 0x0)}}}
	body.LSAs[2] = &packetv3.LSA{Age:0x29, Type:0x2003, ID:0x2, AdvertisingRouter:0x1010101, SequenceNumber:0x80000001, Checksum:0xbaf6, Length:0x24, Body:&packetv3.InterAreaPrefixLSA{Metric:packetv3.InterfaceMetric{High:0x0, Low:0x54}, Prefix:packetv3.LSAPrefix{PrefixLength:0x40, Options:packetv3.PrefixOptions{NoUnicast:false, LocalAddress:false, Propagate:false, DN:false}, Special:0x0, Address:net.IPv6(0x20010db800000004, 0x0)}}}
	body.LSAs[3] = &packetv3.LSA{Age:0x29, Type:0x2003, ID:0x1, AdvertisingRouter:0x1010101, SequenceNumber:0x80000001, Checksum:0xeba0, Length:0x24, Body:&packetv3.InterAreaPrefixLSA{Metric:packetv3.InterfaceMetric{High:0x0, Low:0x4a}, Prefix:packetv3.LSAPrefix{PrefixLength:0x40, Options:packetv3.PrefixOptions{NoUnicast:false, LocalAddress:false, Propagate:false, DN:false}, Special:0x0, Address:net.IPv6(0x20010db800000034, 0x0)}}}
	body.LSAs[4] = &packetv3.LSA{Age:0x29, Type:0x2003, ID:0x0, AdvertisingRouter:0x1010101, SequenceNumber:0x80000001, Checksum:0xebd, Length:0x24, Body:&packetv3.InterAreaPrefixLSA{Metric:packetv3.InterfaceMetric{High:0x0, Low:0x40}, Prefix:packetv3.LSAPrefix{PrefixLength:0x40, Options:packetv3.PrefixOptions{NoUnicast:false, LocalAddress:false, Propagate:false, DN:false}, Special:0x0, Address:net.IPv6(0x20010db800000000, 0x0)}}}
	body.LSAs[5] = &packetv3.LSA{Age:0x23, Type:0x8, ID:0x5, AdvertisingRouter:0x1010101, SequenceNumber:0x80000002, Checksum:0x3d08, Length:0x38, Body:&packetv3.LinkLSA{RouterPriority:0x1, Options:packetv3.RouterOptions{Flags:0x33}, LinkLocalInterfaceAddress:net.IPv6(0xfe80000000000000, 0x1), PrefixNum:0x1, Prefixes:[]packetv3.LSAPrefix{packetv3.LSAPrefix{PrefixLength:0x40, Options:packetv3.PrefixOptions{NoUnicast:false, LocalAddress:false, Propagate:false, DN:false}, Special:0x0, Address:net.IPv6(0x20010db800000012, 0x0)}}}}
	body.LSAs[6] = &packetv3.LSA{Age:0x23, Type:0x2009, ID:0x0, AdvertisingRouter:0x1010101, SequenceNumber:0x80000001, Checksum:0xe8d2, Length:0x2c, Body:&packetv3.IntraAreaPrefixLSA{ReferencedLSType:0x2001, ReferencedLinkStateID:0x0, ReferencedAdvertisingRouter:0x1010101, Prefixes:[]packetv3.LSAPrefix{packetv3.LSAPrefix{PrefixLength:0x40, Options:packetv3.PrefixOptions{NoUnicast:false, LocalAddress:false, Propagate:false, DN:false}, Special:0xa, Address:net.IPv6(0x20010db800000012, 0x0)}}}}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_016() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x4, PacketLength:0xe8, RouterID:0x2020202, AreaID:0x1, Checksum:0xe1b8, InstanceID:0x0, Body:nil}
	body := &packetv3.LinkStateUpdate{LSAs:[]*packetv3.LSA{nil, nil, nil, nil, nil, nil}}
	body.LSAs = make([]*packetv3.LSA, 6)
	body.LSAs[0] = &packetv3.LSA{Age:0x5, Type:0x2001, ID:0x0, AdvertisingRouter:0x2020202, SequenceNumber:0x80000002, Checksum:0xb354, Length:0x18, Body:&packetv3.RouterLSA{Flags:0x1, Options:packetv3.RouterOptions{Flags:0x33}, LinkDescriptions:[]packetv3.AreaLinkDescription(nil)}}
	body.LSAs[1] = &packetv3.LSA{Age:0x6, Type:0x2003, ID:0x3, AdvertisingRouter:0x2020202, SequenceNumber:0x80000001, Checksum:0x4473, Length:0x24, Body:&packetv3.InterAreaPrefixLSA{Metric:packetv3.InterfaceMetric{High:0x0, Low:0x4a}, Prefix:packetv3.LSAPrefix{PrefixLength:0x40, Options:packetv3.PrefixOptions{NoUnicast:false, LocalAddress:false, Propagate:false, DN:false}, Special:0x0, Address:net.IPv6(0x20010db800000003, 0x0)}}}
	body.LSAs[2] = &packetv3.LSA{Age:0x6, Type:0x2003, ID:0x2, AdvertisingRouter:0x2020202, SequenceNumber:0x80000001, Checksum:0x9c11, Length:0x24, Body:&packetv3.InterAreaPrefixLSA{Metric:packetv3.InterfaceMetric{High:0x0, Low:0x54}, Prefix:packetv3.LSAPrefix{PrefixLength:0x40, Options:packetv3.PrefixOptions{NoUnicast:false, LocalAddress:false, Propagate:false, DN:false}, Special:0x0, Address:net.IPv6(0x20010db800000004, 0x0)}}}
	body.LSAs[3] = &packetv3.LSA{Age:0x6, Type:0x2003, ID:0x1, AdvertisingRouter:0x2020202, SequenceNumber:0x80000001, Checksum:0xcdba, Length:0x24, Body:&packetv3.InterAreaPrefixLSA{Metric:packetv3.InterfaceMetric{High:0x0, Low:0x4a}, Prefix:packetv3.LSAPrefix{PrefixLength:0x40, Options:packetv3.PrefixOptions{NoUnicast:false, LocalAddress:false, Propagate:false, DN:false}, Special:0x0, Address:net.IPv6(0x20010db800000034, 0x0)}}}
	body.LSAs[4] = &packetv3.LSA{Age:0x6, Type:0x2003, ID:0x0, AdvertisingRouter:0x2020202, SequenceNumber:0x80000001, Checksum:0xefd7, Length:0x24, Body:&packetv3.InterAreaPrefixLSA{Metric:packetv3.InterfaceMetric{High:0x0, Low:0x40}, Prefix:packetv3.LSAPrefix{PrefixLength:0x40, Options:packetv3.PrefixOptions{NoUnicast:false, LocalAddress:false, Propagate:false, DN:false}, Special:0x0, Address:net.IPv6(0x20010db800000000, 0x0)}}}
	body.LSAs[5] = &packetv3.LSA{Age:0x5, Type:0x8, ID:0x5, AdvertisingRouter:0x2020202, SequenceNumber:0x80000001, Checksum:0x5433, Length:0x2c, Body:&packetv3.LinkLSA{RouterPriority:0x1, Options:packetv3.RouterOptions{Flags:0x33}, LinkLocalInterfaceAddress:net.IPv6(0xfe80000000000000, 0x2), PrefixNum:0x0, Prefixes:[]packetv3.LSAPrefix(nil)}}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_017() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x2, PacketLength:0x1c, RouterID:0x1010101, AreaID:0x1, Checksum:0xda2d, InstanceID:0x0, Body:nil}
	body := &packetv3.DatabaseDescription{Options:packetv3.RouterOptions{Flags:0x13}, InterfaceMTU:0x5dc, DBFlags:0x0, DDSequenceNumber:0x1d48, LSAHeaders:[]*packetv3.LSA(nil)}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_018() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x4, PacketLength:0x3c, RouterID:0x2020202, AreaID:0x1, Checksum:0x197a, InstanceID:0x0, Body:nil}
	body := &packetv3.LinkStateUpdate{LSAs:[]*packetv3.LSA{nil}}
	body.LSAs = make([]*packetv3.LSA, 1)
	body.LSAs[0] = &packetv3.LSA{Age:0x1, Type:0x2001, ID:0x0, AdvertisingRouter:0x2020202, SequenceNumber:0x80000003, Checksum:0x37a5, Length:0x28, Body:&packetv3.RouterLSA{Flags:0x1, Options:packetv3.RouterOptions{Flags:0x33}, LinkDescriptions:[]packetv3.AreaLinkDescription{packetv3.AreaLinkDescription{Type:0x2, Metric:packetv3.InterfaceMetric{High:0x0, Low:0xa}, InterfaceID:0x5, NeighborInterfaceID:0x5, NeighborRouterID:0x1010101}}}}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_019() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x4, PacketLength:0xa8, RouterID:0x1010101, AreaID:0x1, Checksum:0x722a, InstanceID:0x0, Body:nil}
	body := &packetv3.LinkStateUpdate{LSAs:[]*packetv3.LSA{nil, nil, nil, nil}}
	body.LSAs = make([]*packetv3.LSA, 4)
	body.LSAs[0] = &packetv3.LSA{Age:0x1, Type:0x2002, ID:0x5, AdvertisingRouter:0x1010101, SequenceNumber:0x80000001, Checksum:0x27cc, Length:0x20, Body:&packetv3.NetworkLSA{Options:packetv3.RouterOptions{Flags:0x33}, AttachedRouter:[]packetv3.ID{0x1010101, 0x2020202}}}
	body.LSAs[1] = &packetv3.LSA{Age:0x1, Type:0x2009, ID:0x1400, AdvertisingRouter:0x1010101, SequenceNumber:0x80000001, Checksum:0x8f1c, Length:0x2c, Body:&packetv3.IntraAreaPrefixLSA{ReferencedLSType:0x2002, ReferencedLinkStateID:0x5, ReferencedAdvertisingRouter:0x1010101, Prefixes:[]packetv3.LSAPrefix{packetv3.LSAPrefix{PrefixLength:0x40, Options:packetv3.PrefixOptions{NoUnicast:false, LocalAddress:false, Propagate:false, DN:false}, Special:0x0, Address:net.IPv6(0x20010db800000012, 0x0)}}}}
	body.LSAs[2] = &packetv3.LSA{Age:0xe10, Type:0x2009, ID:0x0, AdvertisingRouter:0x1010101, SequenceNumber:0x80000002, Checksum:0x14f6, Length:0x20, Body:&packetv3.IntraAreaPrefixLSA{ReferencedLSType:0x2001, ReferencedLinkStateID:0x0, ReferencedAdvertisingRouter:0x1010101, Prefixes:[]packetv3.LSAPrefix(nil)}}
	body.LSAs[3] = &packetv3.LSA{Age:0x1, Type:0x2001, ID:0x0, AdvertisingRouter:0x1010101, SequenceNumber:0x80000003, Checksum:0x558b, Length:0x28, Body:&packetv3.RouterLSA{Flags:0x1, Options:packetv3.RouterOptions{Flags:0x33}, LinkDescriptions:[]packetv3.AreaLinkDescription{packetv3.AreaLinkDescription{Type:0x2, Metric:packetv3.InterfaceMetric{High:0x0, Low:0xa}, InterfaceID:0x5, NeighborInterfaceID:0x5, NeighborRouterID:0x1010101}}}}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_020() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x4, PacketLength:0x4c, RouterID:0x2020202, AreaID:0x1, Checksum:0xd39f, InstanceID:0x0, Body:nil}
	body := &packetv3.LinkStateUpdate{LSAs:[]*packetv3.LSA{nil}}
	body.LSAs = make([]*packetv3.LSA, 1)
	body.LSAs[0] = &packetv3.LSA{Age:0x1, Type:0x8, ID:0x5, AdvertisingRouter:0x2020202, SequenceNumber:0x80000002, Checksum:0x350b, Length:0x38, Body:&packetv3.LinkLSA{RouterPriority:0x1, Options:packetv3.RouterOptions{Flags:0x33}, LinkLocalInterfaceAddress:net.IPv6(0xfe80000000000000, 0x2), PrefixNum:0x1, Prefixes:[]packetv3.LSAPrefix{packetv3.LSAPrefix{PrefixLength:0x40, Options:packetv3.PrefixOptions{NoUnicast:false, LocalAddress:false, Propagate:false, DN:false}, Special:0x0, Address:net.IPv6(0x20010db800000012, 0x0)}}}}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_021() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x5, PacketLength:0x88, RouterID:0x1010101, AreaID:0x1, Checksum:0x9d2c, InstanceID:0x0, Body:nil}
	body := &packetv3.LinkStateAcknowledgement{LSAHeaders:[]*packetv3.LSA{nil, nil, nil, nil, nil, nil}}
	body.LSAHeaders = make([]*packetv3.LSA, 6)
	body.LSAHeaders[0] = &packetv3.LSA{Age:0x5, Type:0x2001, ID:0x0, AdvertisingRouter:0x2020202, SequenceNumber:0x80000002, Checksum:0xb354, Length:0x18, Body:packetv3.Serializable(nil)}
	body.LSAHeaders[1] = &packetv3.LSA{Age:0x6, Type:0x2003, ID:0x3, AdvertisingRouter:0x2020202, SequenceNumber:0x80000001, Checksum:0x4473, Length:0x24, Body:packetv3.Serializable(nil)}
	body.LSAHeaders[2] = &packetv3.LSA{Age:0x6, Type:0x2003, ID:0x2, AdvertisingRouter:0x2020202, SequenceNumber:0x80000001, Checksum:0x9c11, Length:0x24, Body:packetv3.Serializable(nil)}
	body.LSAHeaders[3] = &packetv3.LSA{Age:0x6, Type:0x2003, ID:0x1, AdvertisingRouter:0x2020202, SequenceNumber:0x80000001, Checksum:0xcdba, Length:0x24, Body:packetv3.Serializable(nil)}
	body.LSAHeaders[4] = &packetv3.LSA{Age:0x6, Type:0x2003, ID:0x0, AdvertisingRouter:0x2020202, SequenceNumber:0x80000001, Checksum:0xefd7, Length:0x24, Body:packetv3.Serializable(nil)}
	body.LSAHeaders[5] = &packetv3.LSA{Age:0x5, Type:0x8, ID:0x5, AdvertisingRouter:0x2020202, SequenceNumber:0x80000001, Checksum:0x5433, Length:0x2c, Body:packetv3.Serializable(nil)}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_022() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x5, PacketLength:0xc4, RouterID:0x2020202, AreaID:0x1, Checksum:0x8b15, InstanceID:0x0, Body:nil}
	body := &packetv3.LinkStateAcknowledgement{LSAHeaders:[]*packetv3.LSA{nil, nil, nil, nil, nil, nil, nil, nil, nil}}
	body.LSAHeaders = make([]*packetv3.LSA, 9)
	body.LSAHeaders[0] = &packetv3.LSA{Age:0x28, Type:0x2001, ID:0x0, AdvertisingRouter:0x1010101, SequenceNumber:0x80000002, Checksum:0xd13a, Length:0x18, Body:packetv3.Serializable(nil)}
	body.LSAHeaders[1] = &packetv3.LSA{Age:0x29, Type:0x2003, ID:0x3, AdvertisingRouter:0x1010101, SequenceNumber:0x80000001, Checksum:0x6259, Length:0x24, Body:packetv3.Serializable(nil)}
	body.LSAHeaders[2] = &packetv3.LSA{Age:0x29, Type:0x2003, ID:0x2, AdvertisingRouter:0x1010101, SequenceNumber:0x80000001, Checksum:0xbaf6, Length:0x24, Body:packetv3.Serializable(nil)}
	body.LSAHeaders[3] = &packetv3.LSA{Age:0x29, Type:0x2003, ID:0x1, AdvertisingRouter:0x1010101, SequenceNumber:0x80000001, Checksum:0xeba0, Length:0x24, Body:packetv3.Serializable(nil)}
	body.LSAHeaders[4] = &packetv3.LSA{Age:0x29, Type:0x2003, ID:0x0, AdvertisingRouter:0x1010101, SequenceNumber:0x80000001, Checksum:0xebd, Length:0x24, Body:packetv3.Serializable(nil)}
	body.LSAHeaders[5] = &packetv3.LSA{Age:0x23, Type:0x8, ID:0x5, AdvertisingRouter:0x1010101, SequenceNumber:0x80000002, Checksum:0x3d08, Length:0x38, Body:packetv3.Serializable(nil)}
	body.LSAHeaders[6] = &packetv3.LSA{Age:0x23, Type:0x2009, ID:0x0, AdvertisingRouter:0x1010101, SequenceNumber:0x80000001, Checksum:0xe8d2, Length:0x2c, Body:packetv3.Serializable(nil)}
	body.LSAHeaders[7] = &packetv3.LSA{Age:0x1, Type:0x2002, ID:0x5, AdvertisingRouter:0x1010101, SequenceNumber:0x80000001, Checksum:0x27cc, Length:0x20, Body:packetv3.Serializable(nil)}
	body.LSAHeaders[8] = &packetv3.LSA{Age:0x1, Type:0x2009, ID:0x1400, AdvertisingRouter:0x1010101, SequenceNumber:0x80000001, Checksum:0x8f1c, Length:0x2c, Body:packetv3.Serializable(nil)}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_023() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x1, PacketLength:0x28, RouterID:0x2020202, AreaID:0x1, Checksum:0xf173, InstanceID:0x0, Body:nil}
	body := &packetv3.Hello{InterfaceID:0x5, RouterPriority:0x1, Options:packetv3.RouterOptions{Flags:0x13}, HelloInterval:0xa, RouterDeadInterval:0x28, DesignatedRouterID:0x1010101, BackupDesignatedRouterID:0x2020202, Neighbors:[]packetv3.ID{0x1010101}}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_024() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x4, PacketLength:0x3c, RouterID:0x2020202, AreaID:0x1, Checksum:0x19fc, InstanceID:0x0, Body:nil}
	body := &packetv3.LinkStateUpdate{LSAs:[]*packetv3.LSA{nil}}
	body.LSAs = make([]*packetv3.LSA, 1)
	body.LSAs[0] = &packetv3.LSA{Age:0x5, Type:0x2001, ID:0x0, AdvertisingRouter:0x2020202, SequenceNumber:0x80000003, Checksum:0x37a5, Length:0x28, Body:&packetv3.RouterLSA{Flags:0x1, Options:packetv3.RouterOptions{Flags:0x33}, LinkDescriptions:[]packetv3.AreaLinkDescription{packetv3.AreaLinkDescription{Type:0x2, Metric:packetv3.InterfaceMetric{High:0x0, Low:0xa}, InterfaceID:0x5, NeighborInterfaceID:0x5, NeighborRouterID:0x1010101}}}}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_025() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x4, PacketLength:0x5c, RouterID:0x1010101, AreaID:0x1, Checksum:0x18a1, InstanceID:0x0, Body:nil}
	body := &packetv3.LinkStateUpdate{LSAs:[]*packetv3.LSA{nil, nil}}
	body.LSAs = make([]*packetv3.LSA, 2)
	body.LSAs[0] = &packetv3.LSA{Age:0xe10, Type:0x2009, ID:0x0, AdvertisingRouter:0x1010101, SequenceNumber:0x80000002, Checksum:0x14f6, Length:0x20, Body:&packetv3.IntraAreaPrefixLSA{ReferencedLSType:0x2001, ReferencedLinkStateID:0x0, ReferencedAdvertisingRouter:0x1010101, Prefixes:[]packetv3.LSAPrefix(nil)}}
	body.LSAs[1] = &packetv3.LSA{Age:0x6, Type:0x2001, ID:0x0, AdvertisingRouter:0x1010101, SequenceNumber:0x80000003, Checksum:0x558b, Length:0x28, Body:&packetv3.RouterLSA{Flags:0x1, Options:packetv3.RouterOptions{Flags:0x33}, LinkDescriptions:[]packetv3.AreaLinkDescription{packetv3.AreaLinkDescription{Type:0x2, Metric:packetv3.InterfaceMetric{High:0x0, Low:0xa}, InterfaceID:0x5, NeighborInterfaceID:0x5, NeighborRouterID:0x1010101}}}}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_026() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x4, PacketLength:0x3c, RouterID:0x1010101, AreaID:0x1, Checksum:0x197, InstanceID:0x0, Body:nil}
	body := &packetv3.LinkStateUpdate{LSAs:[]*packetv3.LSA{nil}}
	body.LSAs = make([]*packetv3.LSA, 1)
	body.LSAs[0] = &packetv3.LSA{Age:0x1, Type:0x2001, ID:0x0, AdvertisingRouter:0x1010101, SequenceNumber:0x80000004, Checksum:0x538c, Length:0x28, Body:&packetv3.RouterLSA{Flags:0x1, Options:packetv3.RouterOptions{Flags:0x33}, LinkDescriptions:[]packetv3.AreaLinkDescription{packetv3.AreaLinkDescription{Type:0x2, Metric:packetv3.InterfaceMetric{High:0x0, Low:0xa}, InterfaceID:0x5, NeighborInterfaceID:0x5, NeighborRouterID:0x1010101}}}}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_027() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x4, PacketLength:0x4c, RouterID:0x2020202, AreaID:0x1, Checksum:0xd421, InstanceID:0x0, Body:nil}
	body := &packetv3.LinkStateUpdate{LSAs:[]*packetv3.LSA{nil}}
	body.LSAs = make([]*packetv3.LSA, 1)
	body.LSAs[0] = &packetv3.LSA{Age:0x5, Type:0x8, ID:0x5, AdvertisingRouter:0x2020202, SequenceNumber:0x80000002, Checksum:0x350b, Length:0x38, Body:&packetv3.LinkLSA{RouterPriority:0x1, Options:packetv3.RouterOptions{Flags:0x33}, LinkLocalInterfaceAddress:net.IPv6(0xfe80000000000000, 0x2), PrefixNum:0x1, Prefixes:[]packetv3.LSAPrefix{packetv3.LSAPrefix{PrefixLength:0x40, Options:packetv3.PrefixOptions{NoUnicast:false, LocalAddress:false, Propagate:false, DN:false}, Special:0x0, Address:net.IPv6(0x20010db800000012, 0x0)}}}}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_028() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x5, PacketLength:0x38, RouterID:0x1010101, AreaID:0x1, Checksum:0x676e, InstanceID:0x0, Body:nil}
	body := &packetv3.LinkStateAcknowledgement{LSAHeaders:[]*packetv3.LSA{nil, nil}}
	body.LSAHeaders = make([]*packetv3.LSA, 2)
	body.LSAHeaders[0] = &packetv3.LSA{Age:0x5, Type:0x2001, ID:0x0, AdvertisingRouter:0x2020202, SequenceNumber:0x80000003, Checksum:0x37a5, Length:0x28, Body:packetv3.Serializable(nil)}
	body.LSAHeaders[1] = &packetv3.LSA{Age:0x5, Type:0x8, ID:0x5, AdvertisingRouter:0x2020202, SequenceNumber:0x80000002, Checksum:0x350b, Length:0x38, Body:packetv3.Serializable(nil)}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_029() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x5, PacketLength:0x38, RouterID:0x2020202, AreaID:0x1, Checksum:0x3dae, InstanceID:0x0, Body:nil}
	body := &packetv3.LinkStateAcknowledgement{LSAHeaders:[]*packetv3.LSA{nil, nil}}
	body.LSAHeaders = make([]*packetv3.LSA, 2)
	body.LSAHeaders[0] = &packetv3.LSA{Age:0xe10, Type:0x2009, ID:0x0, AdvertisingRouter:0x1010101, SequenceNumber:0x80000002, Checksum:0x14f6, Length:0x20, Body:packetv3.Serializable(nil)}
	body.LSAHeaders[1] = &packetv3.LSA{Age:0x6, Type:0x2001, ID:0x0, AdvertisingRouter:0x1010101, SequenceNumber:0x80000003, Checksum:0x558b, Length:0x28, Body:packetv3.Serializable(nil)}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_030() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x1, PacketLength:0x28, RouterID:0x1010101, AreaID:0x1, Checksum:0xf174, InstanceID:0x0, Body:nil}
	body := &packetv3.Hello{InterfaceID:0x5, RouterPriority:0x1, Options:packetv3.RouterOptions{Flags:0x13}, HelloInterval:0xa, RouterDeadInterval:0x28, DesignatedRouterID:0x1010101, BackupDesignatedRouterID:0x2020202, Neighbors:[]packetv3.ID{0x2020202}}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_031() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x4, PacketLength:0x3c, RouterID:0x1010101, AreaID:0x1, Checksum:0x218, InstanceID:0x0, Body:nil}
	body := &packetv3.LinkStateUpdate{LSAs:[]*packetv3.LSA{nil}}
	body.LSAs = make([]*packetv3.LSA, 1)
	body.LSAs[0] = &packetv3.LSA{Age:0x5, Type:0x2001, ID:0x0, AdvertisingRouter:0x1010101, SequenceNumber:0x80000004, Checksum:0x538c, Length:0x28, Body:&packetv3.RouterLSA{Flags:0x1, Options:packetv3.RouterOptions{Flags:0x33}, LinkDescriptions:[]packetv3.AreaLinkDescription{packetv3.AreaLinkDescription{Type:0x2, Metric:packetv3.InterfaceMetric{High:0x0, Low:0xa}, InterfaceID:0x5, NeighborInterfaceID:0x5, NeighborRouterID:0x1010101}}}}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_032() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x4, PacketLength:0x3c, RouterID:0x2020202, AreaID:0x1, Checksum:0x1b78, InstanceID:0x0, Body:nil}
	body := &packetv3.LinkStateUpdate{LSAs:[]*packetv3.LSA{nil}}
	body.LSAs = make([]*packetv3.LSA, 1)
	body.LSAs[0] = &packetv3.LSA{Age:0x1, Type:0x2001, ID:0x0, AdvertisingRouter:0x2020202, SequenceNumber:0x80000004, Checksum:0x35a6, Length:0x28, Body:&packetv3.RouterLSA{Flags:0x1, Options:packetv3.RouterOptions{Flags:0x33}, LinkDescriptions:[]packetv3.AreaLinkDescription{packetv3.AreaLinkDescription{Type:0x2, Metric:packetv3.InterfaceMetric{High:0x0, Low:0xa}, InterfaceID:0x5, NeighborInterfaceID:0x5, NeighborRouterID:0x1010101}}}}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_033() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x5, PacketLength:0x24, RouterID:0x2020202, AreaID:0x1, Checksum:0x509, InstanceID:0x0, Body:nil}
	body := &packetv3.LinkStateAcknowledgement{LSAHeaders:[]*packetv3.LSA{nil}}
	body.LSAHeaders = make([]*packetv3.LSA, 1)
	body.LSAHeaders[0] = &packetv3.LSA{Age:0x5, Type:0x2001, ID:0x0, AdvertisingRouter:0x1010101, SequenceNumber:0x80000004, Checksum:0x538c, Length:0x28, Body:packetv3.Serializable(nil)}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_034() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x5, PacketLength:0x24, RouterID:0x1010101, AreaID:0x1, Checksum:0x22f4, InstanceID:0x0, Body:nil}
	body := &packetv3.LinkStateAcknowledgement{LSAHeaders:[]*packetv3.LSA{nil}}
	body.LSAHeaders = make([]*packetv3.LSA, 1)
	body.LSAHeaders[0] = &packetv3.LSA{Age:0x1, Type:0x2001, ID:0x0, AdvertisingRouter:0x2020202, SequenceNumber:0x80000004, Checksum:0x35a6, Length:0x28, Body:packetv3.Serializable(nil)}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_035() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x1, PacketLength:0x28, RouterID:0x2020202, AreaID:0x1, Checksum:0xf173, InstanceID:0x0, Body:nil}
	body := &packetv3.Hello{InterfaceID:0x5, RouterPriority:0x1, Options:packetv3.RouterOptions{Flags:0x13}, HelloInterval:0xa, RouterDeadInterval:0x28, DesignatedRouterID:0x1010101, BackupDesignatedRouterID:0x2020202, Neighbors:[]packetv3.ID{0x1010101}}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_036() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x1, PacketLength:0x28, RouterID:0x1010101, AreaID:0x1, Checksum:0xf174, InstanceID:0x0, Body:nil}
	body := &packetv3.Hello{InterfaceID:0x5, RouterPriority:0x1, Options:packetv3.RouterOptions{Flags:0x13}, HelloInterval:0xa, RouterDeadInterval:0x28, DesignatedRouterID:0x1010101, BackupDesignatedRouterID:0x2020202, Neighbors:[]packetv3.ID{0x2020202}}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_037() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x1, PacketLength:0x28, RouterID:0x2020202, AreaID:0x1, Checksum:0xf173, InstanceID:0x0, Body:nil}
	body := &packetv3.Hello{InterfaceID:0x5, RouterPriority:0x1, Options:packetv3.RouterOptions{Flags:0x13}, HelloInterval:0xa, RouterDeadInterval:0x28, DesignatedRouterID:0x1010101, BackupDesignatedRouterID:0x2020202, Neighbors:[]packetv3.ID{0x1010101}}
	packet.Body = body
	return packet
}

func packet_OSPFv3_broadcast_adjacency_038() *packetv3.OSPFv3Message {
	packet := &packetv3.OSPFv3Message{Version:0x3, Type:0x1, PacketLength:0x28, RouterID:0x1010101, AreaID:0x1, Checksum:0xf174, InstanceID:0x0, Body:nil}
	body := &packetv3.Hello{InterfaceID:0x5, RouterPriority:0x1, Options:packetv3.RouterOptions{Flags:0x13}, HelloInterval:0xa, RouterDeadInterval:0x28, DesignatedRouterID:0x1010101, BackupDesignatedRouterID:0x2020202, Neighbors:[]packetv3.ID{0x2020202}}
	packet.Body = body
	return packet
}
