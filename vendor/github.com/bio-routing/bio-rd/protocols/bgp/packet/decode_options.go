package packet

// DecodeOptions represents options for the BGP message decoder
type DecodeOptions struct {
	AddPath     bool
	Use32BitASN bool
}
