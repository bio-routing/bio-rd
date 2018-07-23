// foobar
// +bu ild go fuzz

package packet

import (
	"bytes"

	"github.com/bio-routing/bio-rd/protocols/bgp/types"
)

const (
	INC_PRIO = 1
	KEEP     = 0
	DISMISS  = -1
)

func Fuzz(data []byte) int {

	buf := bytes.NewBuffer(data)
	for _, option := range getAllOptions() {
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

func getAllOptions() []types.Options {
	parameters := []bool{true, false}
	var ret []types.Options
	for _, octet := range parameters {
		for _, mpbgp4 := range parameters {
			for _, mpbgp6 := range parameters {
				for _, addPathX := range parameters {
					ret = append(ret, types.Options{
						Supports4OctetASN: octet,
						MultiProtocolIPv4: mpbgp4,
						MultiProtocolIPv6: mpbgp6,
						AddPathRX:         addPathX,
					})
				}
			}
		}
	}
	return ret
}
