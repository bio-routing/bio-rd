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
	TCP = 6
)

var (
	SizeOfTCPHeader = unsafe.Sizeof(TCPHeader{})
)

type TCPHeader struct {
	UrgentPointer  uint16
	Checksum       uint16
	Window         uint16
	Flags          uint8
	DataOffset     uint8
	ACKNumber      uint32
	SequenceNumber uint32
	DstPort        uint16
	SrcPort        uint16
}

func DecodeTCP(raw unsafe.Pointer, length uint32) (*TCPHeader, error) {
	if SizeOfTCPHeader > uintptr(length) {
		return nil, fmt.Errorf("Frame is too short: %d", length)
	}

	return (*TCPHeader)(unsafe.Pointer(uintptr(raw) - SizeOfTCPHeader)), nil
}
