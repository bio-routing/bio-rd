package packet

// AuthenticationType is the type value of an authentication TLV
const AuthenticationType = 10

// AuthenticationTLV represents an authentication TLV
type AuthenticationTLV struct {
	TLVType            uint8
	TLVLength          uint8
	AuthenticationType uint8
	Password           []byte
}
