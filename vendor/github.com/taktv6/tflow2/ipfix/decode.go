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

package ipfix

import (
	"fmt"
	"net"
	"unsafe"

	"github.com/taktv6/tflow2/convert"
)

const (
	// numPreAllocTmplRecs is the number of elements to pre allocate in TemplateRecords slice
	numPreAllocRecs = 20
)

// SetIDTemplateMax is the maximum FlowSetID being used for templates according to RFC3954
const SetIDTemplateMax = 255

// TemplateSetID is the set ID reserved for template sets
const TemplateSetID = 2

// errorIncompatibleVersion prints an error message in case the detected version is not supported
func errorIncompatibleVersion(version uint16) error {
	return fmt.Errorf("IPFIX: Incompatible protocol version v%d, only v10 is supported", version)
}

// Decode is the main function of this package. It converts raw packet bytes to Packet struct.
func Decode(raw []byte, remote net.IP) (*Packet, error) {
	data := convert.Reverse(raw) //TODO: Make it endian aware. This assumes a little endian machine

	pSize := len(data)
	bufSize := 1500
	buffer := [1500]byte{}

	if pSize > bufSize {
		panic("Buffer too small\n")
	}

	// copy data into array as arrays allow us to cast the shit out of it
	for i := 0; i < pSize; i++ {
		buffer[bufSize-pSize+i] = data[i]
	}

	bufferPtr := unsafe.Pointer(&buffer)
	bufferMinPtr := unsafe.Pointer(uintptr(bufferPtr) + uintptr(bufSize) - uintptr(pSize))
	headerPtr := unsafe.Pointer(uintptr(bufferPtr) + uintptr(bufSize) - uintptr(sizeOfHeader))

	var packet Packet
	packet.Buffer = buffer[:]
	packet.Header = (*Header)(headerPtr)

	if packet.Header.Version != 10 {
		return nil, errorIncompatibleVersion(packet.Header.Version)
	}

	//Pre-allocate some room for templates to avoid later copying
	packet.Templates = make([]*TemplateRecords, 0, numPreAllocRecs)

	for uintptr(headerPtr) > uintptr(bufferMinPtr) {
		ptr := unsafe.Pointer(uintptr(headerPtr) - sizeOfSetHeader)

		fls := &Set{
			Header: (*SetHeader)(ptr),
		}

		if fls.Header.SetID == TemplateSetID {
			// Template
			decodeTemplate(&packet, ptr, uintptr(fls.Header.Length)-sizeOfSetHeader, remote)
		} else if fls.Header.SetID > SetIDTemplateMax {
			// Actual data packet
			decodeData(&packet, ptr, uintptr(fls.Header.Length)-sizeOfSetHeader)
		}

		headerPtr = unsafe.Pointer(uintptr(headerPtr) - uintptr(fls.Header.Length))
	}

	return &packet, nil
}

// decodeData decodes a flowSet from `packet`
func decodeData(packet *Packet, headerPtr unsafe.Pointer, size uintptr) {
	flsh := (*SetHeader)(unsafe.Pointer(headerPtr))
	data := unsafe.Pointer(uintptr(headerPtr) - uintptr(flsh.Length))

	fls := &Set{
		Header:  flsh,
		Records: (*(*[1<<31 - 1]byte)(data))[sizeOfSetHeader:flsh.Length],
	}

	packet.FlowSets = append(packet.FlowSets, fls)
}

// decodeTemplate decodes a template from `packet`
func decodeTemplate(packet *Packet, end unsafe.Pointer, size uintptr, remote net.IP) {
	min := uintptr(end) - size
	for uintptr(end) > min {
		headerPtr := unsafe.Pointer(uintptr(end) - sizeOfTemplateRecordHeader)

		tmplRecs := &TemplateRecords{}
		tmplRecs.Header = (*TemplateRecordHeader)(unsafe.Pointer(headerPtr))
		tmplRecs.Packet = packet
		tmplRecs.Records = make([]*TemplateRecord, 0, numPreAllocRecs)

		ptr := unsafe.Pointer(uintptr(headerPtr) - sizeOfTemplateRecordHeader)
		var i uint16
		for i = 0; i < tmplRecs.Header.FieldCount; i++ {
			rec := (*TemplateRecord)(unsafe.Pointer(ptr))
			tmplRecs.Records = append(tmplRecs.Records, rec)
			ptr = unsafe.Pointer(uintptr(ptr) - sizeOfTemplateRecord)
		}

		packet.Templates = append(packet.Templates, tmplRecs)
		end = unsafe.Pointer(uintptr(end) - uintptr(tmplRecs.Header.FieldCount)*sizeOfTemplateRecord - sizeOfTemplateRecordHeader)
	}
}

// PrintHeader prints the header of `packet`
func PrintHeader(p *Packet) {
	fmt.Printf("Version: %d\n", p.Header.Version)
	fmt.Printf("Length: %d\n", p.Header.Length)
	fmt.Printf("UnixSecs: %d\n", p.Header.ExportTime)
	fmt.Printf("Sequence: %d\n", p.Header.SequenceNumber)
	fmt.Printf("DomainId: %d\n", p.Header.DomainID)
}
