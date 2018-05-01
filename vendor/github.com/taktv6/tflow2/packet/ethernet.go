// Copyright 2017 EXARING AG. All Rights Reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package packet

import (
	"fmt"
	"net"
	"unsafe"

	"github.com/taktv6/tflow2/convert"
)

const (
	// EtherTypeARP is Address Resolution Protocol EtherType value
	EtherTypeARP = 0x0806

	// EtherTypeIPv4 is Internet Protocol version 4 EtherType value
	EtherTypeIPv4 = 0x0800

	// EtherTypeIPv6 is Internet Protocol Version 6 EtherType value
	EtherTypeIPv6 = 0x86DD

	// EtherTypeLACP is Link Aggregation Control Protocol EtherType value
	EtherTypeLACP = 0x8809

	// EtherTypeIEEE8021Q is VLAN-tagged frame (IEEE 802.1Q) EtherType value
	EtherTypeIEEE8021Q = 0x8100
)

var (
	// SizeOfEthernetII is the size of an EthernetII header in bytes
	SizeOfEthernetII = unsafe.Sizeof(ethernetII{})
)

// EthernetHeader represents layer two IEEE 802.11
type EthernetHeader struct {
	SrcMAC    net.HardwareAddr
	DstMAC    net.HardwareAddr
	EtherType uint16
}

type ethernetII struct {
	EtherType uint16
	SrcMAC    [6]byte
	DstMAC    [6]byte
}

// DecodeEthernet decodes an EthernetII header
func DecodeEthernet(raw unsafe.Pointer, length uint32) (*EthernetHeader, error) {
	if SizeOfEthernetII > uintptr(length) {
		return nil, fmt.Errorf("Frame is too short: %d", length)
	}

	ptr := unsafe.Pointer(uintptr(raw) - SizeOfEthernetII)
	ethHeader := (*ethernetII)(ptr)

	srcMAC := ethHeader.SrcMAC[:]
	dstMAC := ethHeader.DstMAC[:]

	srcMAC = convert.Reverse(srcMAC)
	dstMAC = convert.Reverse(dstMAC)

	h := &EthernetHeader{
		SrcMAC:    net.HardwareAddr(srcMAC),
		DstMAC:    net.HardwareAddr(dstMAC),
		EtherType: ethHeader.EtherType,
	}

	return h, nil
}
