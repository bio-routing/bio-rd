package ethernet

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"syscall"
	"unsafe"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/tflow2/convert"
)

var (
	wordWidth uint8
)

func init() {
	wordWidth = uint8(unsafe.Sizeof(int(0)))
}

// BPF represents a Berkeley Packet Filter
type BPF struct {
	terms []BPFTerm
}

// NewBPF creates a new empty BPF
func NewBPF() *BPF {
	return &BPF{
		terms: make([]BPFTerm, 0),
	}
}

func (b *BPF) termCount() int {
	return len(b.terms)
}

// AddTerm adds a term to a BPF
func (b *BPF) AddTerm(t BPFTerm) {
	b.terms = append(b.terms, t)
}

// BPFTerm is a BPF Term
type BPFTerm struct {
	Code uint16
	Jt   uint8
	Jf   uint8
	K    uint32
}

func (b *BPF) serializeTerms() []byte {
	directives := bytes.NewBuffer(nil)
	for _, t := range b.terms {
		directives.Write(bnet.BigEndianToLocal(convert.Uint16Byte(t.Code)))
		directives.WriteByte(t.Jt)
		directives.WriteByte(t.Jf)
		directives.Write(bnet.BigEndianToLocal(convert.Uint32Byte(t.K)))
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
		return errors.New("unknown word width")
	}

	err := syscall.SetsockoptString(e.socket, syscall.SOL_SOCKET, syscall.SO_ATTACH_FILTER, string(buf.Bytes()))
	if err != nil {
		return fmt.Errorf("setsockopt failed (SO_ATTACH_FILTER): %w", err)
	}

	return nil
}
