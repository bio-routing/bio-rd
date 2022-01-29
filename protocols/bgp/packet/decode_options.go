package packet

// DecodeOptions represents options for the BGP message decoder
type DecodeOptions struct {
	AddPathIPv4Unicast bool
	AddPathIPv6Unicast bool
	Use32BitASN        bool
}

func (d *DecodeOptions) addPath(afi uint16, safi uint8) bool {
	switch afi {
	case AFIIPv4:
		switch safi {
		case SAFIUnicast:
			return d.AddPathIPv4Unicast
		}
	case AFIIPv6:
		switch safi {
		case SAFIUnicast:
			return d.AddPathIPv6Unicast
		}
	}

	return false
}
