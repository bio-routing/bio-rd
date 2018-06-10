package packet

const DynamicHostNameTLVType = 137

type DynamicHostNameTLV struct {
	TLVType   uint8
	TLVLength uint8
	Hostname  []byte
}
