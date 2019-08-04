// taken from https://go.googlesource.com/net/+/refs/changes/17/112817/2/ipv4/header.go#102
//
// Copyright (c) 2009 The Go Authors. All rights reserved.
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//    * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//    * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package checksum

import "encoding/binary"

// IPChecksum calculates the checksum for arbitrary data
// the way it would for an Internet Protocol header.
//
// Specify the position of the checksum using sumAt.
// Use a value lower than 0 to not skip a checksum field.
func IPChecksum(b []byte, sumAt int) uint16 {
	// Algorithm taken from: https://en.wikipedia.org/wiki/IPv4_header_checksum.
	// "First calculate the sum of each 16 bit value within the header,
	// skipping only the checksum field itself."
	var chk uint32
	for i := 0; i < len(b); i += 2 {
		if i == sumAt {
			continue
		}
		chk += uint32(binary.BigEndian.Uint16(b[i : i+2]))
	}

	// "The first 4 bits are the carry and will be added to the rest of
	// the value."
	carry := uint16(chk >> 16)
	sum := carry + uint16(chk&0x0ffff)

	// "Next, we flip every bit in that value, to obtain the checksum."
	return uint16(^sum)
}
