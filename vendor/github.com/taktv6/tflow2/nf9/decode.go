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

package nf9

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

// FlowSetIDTemplateMax is the maximum FlowSetID being used for templates according to RFC3954
const FlowSetIDTemplateMax = 255

// TemplateFlowSetID is the FlowSetID reserved for template flow sets
const TemplateFlowSetID = 0

// OptionTemplateFlowSetID is the FlowSetID reserved for option template flow sets
const OptionTemplateFlowSetID = 1

// errorIncompatibleVersion prints an error message in case the detected version is not supported
func errorIncompatibleVersion(version uint16) error {
	return fmt.Errorf("NF9: Incompatible protocol version v%d, only v9 is supported", version)
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

	if packet.Header.Version != 9 {
		return nil, errorIncompatibleVersion(packet.Header.Version)
	}

	//Pre-allocate some room for templates to avoid later copying
	packet.Templates = make([]*TemplateRecords, 0, numPreAllocRecs)

	for uintptr(headerPtr) > uintptr(bufferMinPtr) {
		ptr := unsafe.Pointer(uintptr(headerPtr) - sizeOfFlowSetHeader)

		fls := &FlowSet{
			Header: (*FlowSetHeader)(ptr),
		}

		if fls.Header.FlowSetID == TemplateFlowSetID {
			// Template
			decodeTemplate(&packet, ptr, uintptr(fls.Header.Length)-sizeOfFlowSetHeader, remote)
		} else if fls.Header.FlowSetID == OptionTemplateFlowSetID {
			// Option Template
			decodeOption(&packet, ptr, uintptr(fls.Header.Length)-sizeOfFlowSetHeader, remote)
		} else if fls.Header.FlowSetID > FlowSetIDTemplateMax {
			// Actual data packet
			decodeData(&packet, ptr, uintptr(fls.Header.Length)-sizeOfFlowSetHeader)
		}

		headerPtr = unsafe.Pointer(uintptr(headerPtr) - uintptr(fls.Header.Length))
	}

	return &packet, nil
}

// decodeOption decodes an option template from `packet`
func decodeOption(packet *Packet, end unsafe.Pointer, size uintptr, remote net.IP) {
	min := uintptr(end) - size

	for uintptr(end) > min {
		headerPtr := unsafe.Pointer(uintptr(end) - sizeOfOptionsTemplateRecordHeader)

		tmplRecs := &TemplateRecords{}
		hdr := (*OptionsTemplateRecordHeader)(unsafe.Pointer(headerPtr))
		tmplRecs.Header = &TemplateRecordHeader{TemplateID: hdr.TemplateID}
		tmplRecs.Packet = packet
		tmplRecs.Records = make([]*TemplateRecord, 0, numPreAllocRecs)

		ptr := headerPtr
		// Process option scopes
		for i := uint16(0); i < hdr.OptionScopeLength/uint16(sizeOfOptionScope); i++ {
			optScope := (*OptionScope)(ptr)
			tmplRecs.OptionScopes = append(tmplRecs.OptionScopes, optScope)
			ptr = unsafe.Pointer(uintptr(ptr) - sizeOfOptionScope)
		}

		// Process option fields
		for i := uint16(0); i < hdr.OptionLength/uint16(sizeOfTemplateRecord); i++ {
			opt := (*TemplateRecord)(ptr)
			tmplRecs.Records = append(tmplRecs.Records, opt)
			ptr = unsafe.Pointer(uintptr(ptr) - sizeOfTemplateRecord)
		}

		//packet.OptionsTemplates = append(packet.OptionsTemplates, tmplRecs)
		packet.Templates = append(packet.Templates, tmplRecs)

		end = unsafe.Pointer(uintptr(end) - uintptr(hdr.OptionScopeLength) - uintptr(hdr.OptionLength) - sizeOfOptionsTemplateRecordHeader)
	}
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

// decodeData decodes a flowSet from `packet`
func decodeData(packet *Packet, headerPtr unsafe.Pointer, size uintptr) {
	flsh := (*FlowSetHeader)(unsafe.Pointer(headerPtr))
	data := unsafe.Pointer(uintptr(headerPtr) - uintptr(flsh.Length))

	fls := &FlowSet{
		Header: flsh,
		Flows:  (*(*[1<<31 - 1]byte)(data))[sizeOfFlowSetHeader:flsh.Length],
	}

	packet.FlowSets = append(packet.FlowSets, fls)
}

// PrintHeader prints the header of `packet`
func PrintHeader(p *Packet) {
	fmt.Printf("Version: %d\n", p.Header.Version)
	fmt.Printf("Count: %d\n", p.Header.Count)
	fmt.Printf("SysUpTime: %d\n", p.Header.SysUpTime)
	fmt.Printf("UnixSecs: %d\n", p.Header.UnixSecs)
	fmt.Printf("Sequence: %d\n", p.Header.SequenceNumber)
	fmt.Printf("SourceId: %d\n", p.Header.SourceID)
}
