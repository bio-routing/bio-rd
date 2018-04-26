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
	"unsafe"
)

const (
	// UDP IP protocol number
	UDP = 17
)

var (
	// SizeOfUDPHeader is the size of a UDP header in bytes
	SizeOfUDPHeader = unsafe.Sizeof(UDPHeader{})
)

// UDPHeader represents a UDP header
type UDPHeader struct {
	Checksum uint16
	Length   uint16
	DstPort  uint16
	SrcPort  uint16
}

// DecodeUDP decodes a UDP header
func DecodeUDP(raw unsafe.Pointer, length uint32) (*UDPHeader, error) {
	if SizeOfTCPHeader > uintptr(length) {
		return nil, fmt.Errorf("Frame is too short: %d", length)
	}

	return (*UDPHeader)(unsafe.Pointer(uintptr(raw) - SizeOfUDPHeader)), nil
}
