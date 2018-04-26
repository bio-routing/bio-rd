// Copyright 2017 Google Inc. All Rights Reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package convert provides helper functions to convert data between
// various types, e.g. []byte to int, etc.
package convert

import (
	"bytes"
	"encoding/binary"
	"net"
	"strings"
)

// IPByteSlice converts a string that contains an IP address into byte slice
func IPByteSlice(ip string) []byte {
	ret := net.ParseIP(ip)
	if strings.Contains(ip, ".") {
		ipv4 := make([]byte, net.IPv4len)
		tmp := []byte(ret)
		copy(ipv4, tmp[len(tmp)-net.IPv4len:])
		ret = ipv4
	}
	return ret
}

// Uint16b converts a byte slice to uint16 assuming the slice is BigEndian
func Uint16b(data []byte) (ret uint16) {
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.BigEndian, &ret)
	return
}

// Uint32b converts a byte slice to uint32 assuming the slice is BigEndian
func Uint32b(data []byte) (ret uint32) {
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.BigEndian, &ret)
	return
}

// Uint64b converts a byte slice to uint64 assuming the slice is BigEndian
func Uint64b(data []byte) (ret uint64) {
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.BigEndian, &ret)
	return
}

// Uint16 converts a byte slice into uint16 assuming LittleEndian
func Uint16(data []byte) (ret uint16) {
	return uint16(UintX(data))
}

// Uint32 converts a byte slice into uint32 assuming LittleEndian
func Uint32(data []byte) (ret uint32) {
	return uint32(UintX(data))
}

// Uint64 converts a byte slice into uint64 assuming LittleEndian
func Uint64(data []byte) uint64 {
	return UintX(data)
}

// UintX converts a byte slice into uint64 assuming LittleEndian
func UintX(data []byte) (ret uint64) {
	size := uint8(len(data))
	var i uint8
	for i = 0; i < size; i++ {
		ret += (uint64(data[i]) << (i * 8))
	}
	return ret
}

// Uint8Byte converts a uint8 to a byte slice in BigEndian
func Uint8Byte(data uint8) (ret []byte) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, data)
	return buf.Bytes()
}

// Uint16Byte converts a uint16 to a byte slice in BigEndian
func Uint16Byte(data uint16) (ret []byte) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, data)
	return buf.Bytes()
}

// Uint32Byte converts a uint16 to a byte slice in BigEndian
func Uint32Byte(data uint32) (ret []byte) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, data)
	return buf.Bytes()
}

// Int64Byte converts a int64 to a byte slice in BigEndian
func Int64Byte(data int64) (ret []byte) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, data)
	return buf.Bytes()
}

// Uint64Byte converts a uint64 to a byte slice in BigEndian
func Uint64Byte(data uint64) (ret []byte) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, data)
	return buf.Bytes()
}

// Reverse reverses byte slice without allocating new memory
func Reverse(data []byte) []byte {
	n := len(data)
	for i := 0; i < n/2; i++ {
		data[i], data[n-i-1] = data[n-i-1], data[i]
	}
	return data
}
