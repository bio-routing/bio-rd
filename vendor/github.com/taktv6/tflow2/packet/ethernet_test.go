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
	"testing"
	"unsafe"
)

func TestDecode(t *testing.T) {
	data := []byte{
		128,               // Header Length
		92, 180, 133, 203, // ACK Number
		31, 4, 191, 24, // Sequence Number
		222, 148, // DST port
		80, 0, // SRC port

		19, 131, 191, 87, // DST IP
		238, 153, 37, 185, // SRC IP
		186, 25, // Header Checksum
		6,     // Protocol
		62,    // TTL
		0, 64, // Flags + Fragment offset
		131, 239, // Identifier
		212, 5, // Total Length
		0,  // TOS
		69, // Version + Length

		0, 8, // EtherType
		185, 28, 4, 113, 78, 32, // Source MAC
		148, 2, 127, 31, 113, 128, // Destination MAC
	}

	pSize := len(data)
	bufSize := 128
	buffer := [128]byte{}

	if pSize > bufSize {
		panic("Buffer too small\n")
	}

	// copy data into array as arrays allow us to cast the shit out of it
	for i := 0; i < pSize; i++ {
		buffer[bufSize-pSize+i] = data[i]
	}

	bufferPtr := unsafe.Pointer(&buffer)
	headerPtr := unsafe.Pointer(uintptr(bufferPtr) + uintptr(bufSize))

	etherHeader, err := DecodeEthernet(headerPtr, 128)
	if err != nil {
		t.Errorf("Decoding packet failed: %v\n", err)
	}

	if etherHeader.DstMAC.String() != "80:71:1f:7f:02:94" {
		t.Errorf("Unexpected DST MAC address. Expected %s. Got %s", "80:71:1f:7f:02:94", etherHeader.DstMAC.String())
	}

	if etherHeader.SrcMAC.String() != "20:4e:71:04:1c:b9" {
		t.Errorf("Unexpected DST MAC address. Expected %s. Got %s", "20:4e:71:04:1c:b9", etherHeader.SrcMAC.String())
	}

	if etherHeader.EtherType != EtherTypeIPv4 {
		t.Errorf("Unexpected ethertyp. Expected %d. Got %d", EtherTypeIPv4, etherHeader.EtherType)
	}
}
