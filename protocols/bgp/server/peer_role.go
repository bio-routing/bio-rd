package server

import "github.com/bio-routing/bio-rd/protocols/bgp/packet"

const (
	// BGP Role configuration, these values will be translated into values defined by the RFC on peer setup
	PeerConfigRoleOff      = 0
	PeerConfigRoleProvider = 1
	PeerConfigRoleRS       = 2
	PeerConfigRoleRSClient = 3
	PeerConfigRoleCustomer = 4
	PeerConfigRolePeer     = 5
)

func translatePeerRole(pr uint8) uint8 {
	switch pr {
	case PeerConfigRoleProvider:
		return packet.PeerRoleRoleProvider
	case PeerConfigRoleCustomer:
		return packet.PeerRoleRoleCustomer
	case PeerConfigRoleRS:
		return packet.PeerRoleRoleRS
	case PeerConfigRoleRSClient:
		return packet.PeerRoleRoleRSClient
	case PeerConfigRolePeer:
		return packet.PeerRoleRolePeer
	default:
		return 255
	}
}

func peerRoleEnabled(pr uint8) bool {
	return pr >= PeerConfigRoleProvider && pr <= PeerConfigRolePeer
}
