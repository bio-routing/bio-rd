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
	SizeOfIPv6Header = unsafe.Sizeof(IPv6Header{})
)

type IPv6Header struct {
	DstAddr                      [16]byte
	SrcAddr                      [16]byte
	HopLimit                     uint8
	NextHeader                   uint8
	PayloadLength                uint16
	VersionTrafficClassFlowLabel uint32
}

func DecodeIPv6(raw unsafe.Pointer, length uint32) (*IPv6Header, error) {
	if SizeOfIPv6Header > uintptr(length) {
		return nil, fmt.Errorf("Frame is too short: %d", length)
	}

	return (*IPv6Header)(unsafe.Pointer(uintptr(raw) - SizeOfIPv6Header)), nil
}
