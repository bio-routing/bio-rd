package server

import (
	"context"

	netapi "github.com/bio-routing/bio-rd/net/api"
	"github.com/bio-routing/bio-rd/protocols/isis/api"
	"github.com/bio-routing/bio-rd/protocols/isis/packet"
)

type ISISAPIServer struct {
	api.UnimplementedIsisServiceServer
	srv ISISServer
}

// NewISISAPIServer creates a new ISIS API Server
func NewISISAPIServer(s ISISServer) *ISISAPIServer {
	return &ISISAPIServer{
		srv: s,
	}
}

func (s *ISISAPIServer) ListAdjacencies(context.Context, *api.ListAdjacenciesRequest) (*api.ListAdjacenciesResponse, error) {
	res := &api.ListAdjacenciesResponse{
		Adjacencies: make([]*api.Adjacency, 0),
	}

	for _, a := range s.srv.GetAdjacencies() {
		addrs := make([]*netapi.IP, 0, len(a.IPAddresses))
		for _, addr := range a.IPAddresses {
			addrs = append(addrs, addr.ToProto())
		}

		adj := &api.Adjacency{
			Name:               a.Name,
			SystemId:           a.SystemID[:],
			Address:            a.Address[:],
			InterfaceName:      a.InterfaceName,
			Level:              uint32(a.Level),
			Priority:           uint32(a.Priority),
			IpAddresses:        addrs,
			LastTransitionUnix: a.LastStateChange.Unix(),
			ExpiresInSeconds:   uint32(a.Timeout.Sub(clock.Now()).Seconds()),
			Status:             api.Adjacency_State(a.Status),
		}

		res.Adjacencies = append(res.Adjacencies, adj)
	}

	return res, nil
}

func (s *ISISAPIServer) GetLSDB(context.Context, *api.GetLSDBRequest) (*api.GetLSDBResponse, error) {
	resp := &api.GetLSDBResponse{
		LsdbEntries: make([]*api.LSDBEntry, 0),
	}

	for _, e := range s.srv.GetLSDB() {
		resp.LsdbEntries = append(resp.LsdbEntries, lsdbEntryToProto(e))
	}

	return resp, nil
}

func lsdbEntryToProto(e *LSDBEntry) *api.LSDBEntry {
	l := &api.LSDBEntry{
		Lsp:                   lspduToProto(e.lspdu),
		InterfacesWithSsnFlag: make([]string, 0, len(e.ssnFlags)),
		InterfacesWithSrmFlag: make([]string, 0, len(e.srmFlags)),
	}

	copy(l.InterfacesWithSsnFlag, e.ssnFlags)
	copy(l.InterfacesWithSrmFlag, e.srmFlags)

	return l
}

func lspduToProto(e *packet.LSPDU) *api.LSPDU {
	ret := &api.LSPDU{
		LspId: &api.LSPID{
			SystemId:     e.LSPID.SystemID[:],
			PseudonodeId: uint32(e.LSPID.PseudonodeID),
			LspNumber:    uint32(e.LSPID.LSPNumber),
		},
		Length:                   uint32(e.Length),
		RemainingLifetime:        uint32(e.RemainingLifetime),
		SequenceNumber:           e.SequenceNumber,
		Checksum:                 uint32(e.Checksum),
		TypeBlock:                uint32(e.TypeBlock),
		AreaIds:                  make([][]byte, 0),
		ProtocolsSupported:       make([]api.LSPDU_Protocol, 0),
		IpInterfacesAddresses:    make([]uint32, 0),
		ExtendedIsReachabilities: make([]*api.ExtendedISReachability, 0),
	}

	for _, tlv := range e.TLVs {
		switch tlv.Type() {
		case packet.AreaAddressesTLVType:
			for _, aid := range tlv.(*packet.AreaAddressesTLV).AreaIDs {
				ret.AreaIds = append(ret.AreaIds, aid[:])
			}
		case packet.ProtocolsSupportedTLVType:
			for _, pid := range tlv.(*packet.ProtocolsSupportedTLV).NetworkLayerProtocolIDs {
				switch pid {
				case packet.NLPIDIPv4:
					ret.ProtocolsSupported = append(ret.ProtocolsSupported, api.LSPDU_IPv4)
				case packet.NLPIDIPv6:
					ret.ProtocolsSupported = append(ret.ProtocolsSupported, api.LSPDU_IPv6)
				}
			}
		case packet.IPInterfaceAddressesTLVType:
			for _, addr := range tlv.(*packet.IPInterfaceAddressesTLV).IPv4Addresses {
				ret.IpInterfacesAddresses = append(ret.IpInterfacesAddresses, addr)
			}
		case packet.ExtendedISReachabilityType:
			for _, n := range tlv.(*packet.ExtendedISReachabilityTLV).Neighbors {
				ret.ExtendedIsReachabilities = append(ret.ExtendedIsReachabilities, &api.ExtendedISReachability{
					NeighborId:    n.NeighborID.Serialize(),
					DefaultMetric: n.Metric,
				})
			}
		case packet.ExtendedIPReachabilityTLVType:
			for _, eipr := range tlv.(*packet.ExtendedIPReachabilityTLV).ExtendedIPReachabilities {
				ret.ExtendedIpReachabilities = append(ret.ExtendedIpReachabilities, &api.ExtendedIPReachability{
					Metric:       eipr.Metric,
					IpAddress:    eipr.Address,
					PrefixLength: uint32(eipr.PfxLen()),
				})
			}
		case packet.DynamicHostNameTLVType:
			ret.Hostname = string(tlv.(*packet.DynamicHostNameTLV).Hostname)
		case packet.TrafficEngineeringRouterIDTLVType:
			ret.Ipv4TeRouterId = tlv.(*packet.TrafficEngineeringRouterIDTLV).Address
		}
	}

	return ret
}
