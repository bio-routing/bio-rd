package types

// Options represents options to the update sender, decoder and encoder
type Options struct {
	Supports4OctetASN bool
	AddPathRX         bool
	MultiProtocolIPv4 bool
	MultiProtocolIPv6 bool
}
