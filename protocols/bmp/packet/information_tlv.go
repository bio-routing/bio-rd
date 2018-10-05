package packet

import (
	"bytes"

	"github.com/bio-routing/bio-rd/util/decoder"
)

const (
	MinInformationTLVLen = 4
)

// InformationTLV represents an information TLV
type InformationTLV struct {
	InformationType   uint16
	InformationLength uint16
	Information       []byte
}

func decodeInformationTLV(buf *bytes.Buffer) (*InformationTLV, error) {
	infoTLV := &InformationTLV{}

	fields := []interface{}{
		&infoTLV.InformationType,
		&infoTLV.InformationLength,
	}

	err := decoder.Decode(buf, fields)
	if err != nil {
		return infoTLV, err
	}

	infoTLV.Information = make([]byte, infoTLV.InformationLength)
	fields = []interface{}{
		&infoTLV.Information,
	}

	err = decoder.Decode(buf, fields)
	if err != nil {
		return infoTLV, err
	}

	return infoTLV, nil
}
