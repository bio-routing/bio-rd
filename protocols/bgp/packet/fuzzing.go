//go:build go || fuzz
// +build go fuzz

package packet

import (
	"bytes"
)

const (
	INC_PRIO = 1
	KEEP     = 0
	DISMISS  = -1
)

func Fuzz(data []byte) int {
	buf := bytes.NewBuffer(data)
	for _, option := range getAllDecodingOptions() {
		msg, err := Decode(buf, &option)
		if err != nil {
			if msg != nil {
				panic("msg != nil on error")
			}

		}

		return INC_PRIO
	}

	return KEEP
}

func getAllDecodingOptions() []DecodeOptions {
	parameters := []bool{true, false}
	var ret []DecodeOptions
	for _, octet := range parameters {
		ret = append(ret, DecodeOptions{
			Use32BitASN: octet,
		})
	}

	return ret
}
