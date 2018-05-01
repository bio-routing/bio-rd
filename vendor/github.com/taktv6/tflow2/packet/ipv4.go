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

var (
	SizeOfIPv4Header = unsafe.Sizeof(IPv4Header{})
)

type IPv4Header struct {
	DstAddr             [4]byte
	SrcAddr             [4]byte
	HeaderChecksum      uint16
	Protocol            uint8
	TTL                 uint8
	FlagsFragmentOffset uint16
	Identification      uint16
	TotalLength         uint16
	DSCP                uint8
	VersionHeaderLength uint8
}

func DecodeIPv4(raw unsafe.Pointer, length uint32) (*IPv4Header, error) {
	if SizeOfIPv4Header > uintptr(length) {
		return nil, fmt.Errorf("Frame is too short: %d", length)
	}

	return (*IPv4Header)(unsafe.Pointer(uintptr(raw) - SizeOfIPv4Header)), nil
}
