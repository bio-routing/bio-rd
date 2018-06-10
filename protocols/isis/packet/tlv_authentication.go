package packet

const AuthenticationType = 10

type AuthenticationTLV struct {
	TLVType            uint8
	TLVLength          uint8
	AuthenticationType uint8
	Password           []byte
}
