package ethernet

import (
	"bytes"
	"reflect"
	"syscall"
	"unsafe"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/tflow2/convert"
	"github.com/pkg/errors"
)

var (
	wordWidth uint8
)

func init() {
	wordWidth = uint8(unsafe.Sizeof(int(0)))
}

// BPF represents a Berkeley Packet Filter
type BPF struct {
	len   uint16
	terms []BPFTerm
}

func (b BPF) termCount() int {
	return len(b.terms)
}

// BPFTerm is a BPF Term
type BPFTerm struct {
	code uint16
	jt   uint8
	jf   uint8
	k    uint32
}

func (b *BPF) serializeTerms() []byte {
	directives := bytes.NewBuffer(nil)
	for _, t := range b.terms {
		directives.Write(bnet.BigEndianToLocal(convert.Uint16Byte(t.code)))
		directives.WriteByte(t.jt)
		directives.WriteByte(t.jf)
		directives.Write(bnet.BigEndianToLocal(convert.Uint32Byte(t.k)))
	}

	return directives.Bytes()
}

func (e *Handler) loadBPF(b *BPF) error {
	if b == nil {
		return nil
	}

	prog := b.serializeTerms()
	buf := bytes.NewBuffer(nil)

	bpfProgTermCount := b.termCount()
	buf.Write(bnet.BigEndianToLocal(convert.Uint16Byte(uint16(bpfProgTermCount))))

	// Align to next word
	for i := 0; i < int(wordWidth)-int(unsafe.Sizeof(uint16(0))); i++ {
		buf.WriteByte(0)
	}

	p := (*reflect.SliceHeader)(unsafe.Pointer(&prog)).Data
	switch wordWidth {
	case 4:
		buf.Write(bnet.BigEndianToLocal(convert.Uint32Byte(uint32(uintptr(p)))))
	case 8:
		buf.Write(bnet.BigEndianToLocal(convert.Uint64Byte(uint64(uintptr(p)))))
	default:
		panic("Unknown word width")
	}

	err := syscall.SetsockoptString(e.socket, syscall.SOL_SOCKET, syscall.SO_ATTACH_FILTER, string(buf.Bytes()))
	if err != nil {
		return errors.Wrap(err, "Setsockopt failed (SO_ATTACH_FILTER)")
	}

	return nil
}
