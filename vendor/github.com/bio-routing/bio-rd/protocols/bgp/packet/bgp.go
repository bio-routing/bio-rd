package packet

const (
	OctetLen       = 8
	MaxASNsSegment = 255
	BGP4Version    = 4
	MinOpenLen     = 29

	MarkerLen         = 16
	HeaderLen         = 19
	MinLen            = 19
	MaxLen            = 4096
	MinUpdateLen      = 4
	NLRIMaxLen        = 5
	AFILen            = 2
	SAFILen           = 1
	CommunityLen      = 4
	LargeCommunityLen = 12
	IPv4Len           = 4
	IPv6Len           = 16
	ClusterIDLen      = 4

	OpenMsg         = 1
	UpdateMsg       = 2
	NotificationMsg = 3
	KeepaliveMsg    = 4

	MessageHeaderError      = 1
	OpenMessageError        = 2
	UpdateMessageError      = 3
	HoldTimeExpired         = 4
	FiniteStateMachineError = 5
	Cease                   = 6

	// Msg Header Errors
	ConnectionNotSync = 1
	BadMessageLength  = 2
	BadMessageType    = 3

	// Open Msg Errors
	UnsupportedVersionNumber     = 1
	BadPeerAS                    = 2
	BadBGPIdentifier             = 3
	UnsupportedOptionalParameter = 4
	DeprecatedOpenMsgError5      = 5
	UnacceptableHoldTime         = 6

	// Update Msg Errors
	MalformedAttributeList    = 1
	UnrecognizedWellKnownAttr = 2
	MissingWellKnownAttr      = 3
	AttrFlagsError            = 4
	AttrLengthError           = 5
	InvalidOriginAttr         = 6
	DeprecatedUpdateMsgError7 = 7
	InvalidNextHopAttr        = 8
	OptionalAttrError         = 9
	InvalidNetworkField       = 10
	MalformedASPath           = 11

	// Notification Msg Subcodes
	AdministrativeShutdown = 2
	AdministrativeReset    = 4

	// Attribute Type Codes
	OriginAttr           = 1
	ASPathAttr           = 2
	NextHopAttr          = 3
	MEDAttr              = 4
	LocalPrefAttr        = 5
	AtomicAggrAttr       = 6
	AggregatorAttr       = 7
	CommunitiesAttr      = 8
	OriginatorIDAttr     = 9
	ClusterListAttr      = 10
	AS4PathAttr          = 17
	AS4AggregatorAttr    = 18
	LargeCommunitiesAttr = 32

	// ORIGIN values
	IGP        = 0
	EGP        = 1
	INCOMPLETE = 2

	// NOTIFICATION Cease error SubCodes (RFC4486)
	MaxPrefReached                = 1
	AdminShut                     = 2
	PeerDeconfigured              = 3
	AdminReset                    = 4
	ConnectionRejected            = 5
	OtherConfigChange             = 8
	ConnectionCollisionResolution = 7
	OutOfResources                = 8

	IPv4AFI                      = 1
	IPv6AFI                      = 2
	UnicastSAFI                  = 1
	CapabilitiesParamType        = 2
	MultiProtocolCapabilityCode  = 1
	MultiProtocolReachNLRICode   = 14
	MultiProtocolUnreachNLRICode = 15
	AddPathCapabilityCode        = 69
	ASN4CapabilityCode           = 65
	AddPathReceive               = 1
	AddPathSend                  = 2
	AddPathSendReceive           = 3
	ASTransASN                   = 23456
)

var (
	afiAddrLenBytes = map[uint16]uint8{
		1: 4,
		2: 16,
	}
)

type BGPError struct {
	ErrorCode    uint8
	ErrorSubCode uint8
	ErrorStr     string
}

func (b BGPError) Error() string {
	return b.ErrorStr
}

type BGPMessage struct {
	Header *BGPHeader
	Body   interface{}
}

type BGPHeader struct {
	Length uint16
	Type   uint8
}

type BGPOpen struct {
	Version       uint8
	ASN           uint16
	HoldTime      uint16
	BGPIdentifier uint32
	OptParmLen    uint8
	OptParams     []OptParam
}

type BGPNotification struct {
	ErrorCode    uint8
	ErrorSubcode uint8
}

type PathAttribute struct {
	Length         uint16
	Optional       bool
	Transitive     bool
	Partial        bool
	ExtendedLength bool
	TypeCode       uint8
	Value          interface{}
	Next           *PathAttribute
}

// AFIName returns the name of an address family
func AFIName(afi uint16) string {
	switch afi {
	case IPv4AFI:
		return "IPv4"
	case IPv6AFI:
		return "IPv6"
	default:
		return "Unknown AFI"
	}
}
