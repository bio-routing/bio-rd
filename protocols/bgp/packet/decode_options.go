package packet

// DecodeOptions are options for decoding BGP Update messages
type DecodeOptions struct {
	Use32BitASN bool
	AddPath     bool
}
