package packet

// DecodeOptions represents options for the BGP message decoder
type DecodeOptions struct {
	AddPathIPv4Unicast bool
	AddPathIPv6Unicast bool
	Use32BitASN        bool
}
