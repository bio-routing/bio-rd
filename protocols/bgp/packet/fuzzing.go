// +build gofuzz

package packet

import "bytes"

const (
	INC_PRIO = 1
	KEEP     = 0
	DISMISS  = -1
)

func Fuzz(data []byte) int {

	buf := bytes.NewBuffer(data)
	msg, err := Decode(buf)
	if err != nil {
		if msg != nil {
			panic("msg != nil on error")
		}
		return KEEP
	}

	return INC_PRIO
}
