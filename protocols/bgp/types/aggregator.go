package types

// Aggregator represents an AGGREGATOR attribute (type code 7) as in RFC4271
type Aggregator struct {
	ASN     uint16
	Address uint32
}
