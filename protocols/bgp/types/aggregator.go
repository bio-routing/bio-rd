package types

// Aggregator represents an AGGREGATOR attribute (type code 7) as in RFC4271
type Aggregator struct {
	Address uint32
	ASN     uint16
}
