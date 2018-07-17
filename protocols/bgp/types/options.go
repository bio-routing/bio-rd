package types

// Options represents options to the update sender, decoder and encoder
type Options struct {
	Supports4OctetASN     bool
	SupportsMultiProtocol bool
	AddPathRX             bool
}
