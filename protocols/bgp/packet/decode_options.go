package packet

// DecodeOptions represents options for the BGP message decoder
type DecodeOptions struct {
	AddPathIPv4Unicast bool
	AddPathIPv6Unicast bool
	Use32BitASN        bool
}

func (d *DecodeOptions) addPath(afi uint16, safi uint8) bool {
	switch afi {
	case IPv4AFI:
		switch safi {
		case UnicastSAFI:
			return d.AddPathIPv4Unicast
		}
	case IPv6AFI:
		switch safi {
		case UnicastSAFI:
			return d.AddPathIPv6Unicast
		}
	}

	return false
}
