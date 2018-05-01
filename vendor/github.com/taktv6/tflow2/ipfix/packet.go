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

// Package ipfix provides structures and functions to decode and analyze
// IPFIX packets.
//
// This package does only packet decoding in a single packet context. It keeps
// no state when decoding multiple packets. As a result Data FlowSets can not be
// decoded during initial packet decoding. To decode Data FlowSets user must
// keep track of Template Records and Options Template Records manually.
//
// Examples of IPFIX packets:
//
//   1. An IPFIX Message consisting of interleaved Template, Data, and
//      Options Template Sets, as shown in Figure C.  Here, Template and
//      Options Template Sets are transmitted "on demand", before the
//      first Data Set whose structure they define.
//
//     +--------+--------------------------------------------------------+
//     |        | +----------+ +---------+     +-----------+ +---------+ |
//     |Message | | Template | | Data    |     | Options   | | Data    | |
//     | Header | | Set      | | Set     | ... | Template  | | Set     | |
//     |        | |          | |         |     | Set       | |         | |
//     |        | +----------+ +---------+     +-----------+ +---------+ |
//     +--------+--------------------------------------------------------+
//
//       +--------+----------------------------------------------+
//       |        | +---------+     +---------+      +---------+ |
//       |Message | | Data    |     | Data    |      | Data    | |
//       | Header | | Set     | ... | Set     | ...  | Set     | |
//       |        | +---------+     +---------+      +---------+ |
//       +--------+----------------------------------------------+
//
//                    Figure D: IPFIX Message: Example 2
//
//   3. An IPFIX Message consisting entirely of Template and Options
//      Template Sets, as shown in Figure E.  Such a message can be used
//      to define or redefine Templates and Options Templates in bulk.
//
//      +--------+-------------------------------------------------+
//      |        | +----------+     +----------+      +----------+ |
//      |Message | | Template |     | Template |      | Options  | |
//      | Header | | Set      | ... | Set      | ...  | Template | |
//      |        | |          |     |          |      | Set      | |
//      |        | +----------+     +----------+      +----------+ |
//      +--------+-------------------------------------------------+
//
//
// Example of struct hierarchy after packet decoding:
//  Package
//  |
//  +--TemplateFlowSet
//  |  |
//  |  +--TemplateRecord
//  |  |  |
//  |  |  +--Field
//  |  |  +--...
//  |  |  +--Field
//  |  |
//  |  +--...
//  |  |
//  |  +--TemplateRecord
//  |     |
//  |     +--Field
//  |     +--...
//  |     +--Field
//  |
//  +--DataFlowSet
//  |
//  +--...
//  |
//  +--OptionsTemplateFlowSet
//  |  |
//  |  +--OptionsTemplateRecord
//  |  |  |
//  |  |  +--Field (scope)
//  |  |  +--...   (scope)
//  |  |  +--Field (scope)
//  |  |  |
//  |  |  +--Field (option)
//  |  |  +--...   (option)
//  |  |  +--Field (option)
//  |  |
//  |  +--...
//  |  |
//  |  +--OptionsTemplateRecord
//  |     |
//  |     +--Field (scope)
//  |     +--...   (scope)
//  |     +--Field (scope)
//  |     |
//  |     +--Field (option)
//  |     +--...   (option)
//  |     +--Field (option)
//  |
//  +--DataFlowSet
//
// When matched with appropriate template Data FlowSet can be decoded to list of
// Flow Data Records or list of Options Data Records. Struct hierarchy example:
//
//  []FlowDataRecord
//    |
//    +--FlowDataRecord
//    |  |
//    |  +--[]byte
//    |  +--...
//    |  +--[]byte
//    |
//    +--...
//    |
//    +--FlowDataRecord
//       |
//       +--[]byte
//       +--...
//       +--[]byte
//
//  []OptionsDataRecord
//    |
//    +--OptionsDataRecord
//    |  |
//    |  +--[]byte (scope)
//    |  +--...    (scope)
//    |  +--[]byte (scope)
//    |  |
//    |  +--[]byte (option)
//    |  +--...    (option)
//    |  +--[]byte (option)
//    |
//    +--...
//    |
//    +--OptionsDataRecord
//       |
//       +--[]byte
//       +--...
//       +--[]byte
//       |
//       +--[]byte (option)
//       +--...    (option)
//       +--[]byte (option)
//
// Most of structure names and comments are taken directly from RFC 7011.
// Reading the IPFIX protocol specification is highly recommended before
// using this package.
package ipfix

import "unsafe"

// Header is an IPFIX message header
type Header struct {
	// A 32-bit value that identifies the Exporter Observation Domain.
	DomainID uint32

	// Incremental sequence counter of all Export Packets sent from the
	// current Observation Domain by the Exporter.
	SequenceNumber uint32

	// Time in seconds since 0000 UTC 1970, at which the Export Packet
	// leaves the Exporter.
	ExportTime uint32

	// The total number of bytes in this Export Packet
	Length uint16

	// Version of Flow Record format exported in this packet. The value of
	//this field is 9 for the current version.
	Version uint16
}

// Set represents a Set as described in RFC7011
type Set struct {
	Header  *SetHeader
	Records []byte
}

// SetHeader is a decoded representation of the header of a Set
type SetHeader struct {
	Length uint16
	SetID  uint16
}

var sizeOfSetHeader = unsafe.Sizeof(SetHeader{})

// Packet is a decoded representation of a single NetFlow v9 UDP packet.
type Packet struct {
	// A pointer to the packets headers
	Header *Header

	// A slice of pointers to FlowSet. Each element is instance of (Data)FlowSet
	// found in this packet
	FlowSets []*Set

	// A slice of pointers to TemplateRecords. Each element is instance of TemplateRecords
	// representing a template found in this packet.
	Templates []*TemplateRecords

	// Buffer is a slice pointing to the original byte array that this packet was decoded from.
	// This field is only populated if debug level is at least 2
	Buffer []byte
}

var sizeOfHeader = unsafe.Sizeof(Header{})

// GetTemplateRecords generate a list of all Template Records in the packet.
// Template Records can be used to decode Data FlowSets to Data Records.
func (p *Packet) GetTemplateRecords() []*TemplateRecords {
	return p.Templates
}

// DataFlowSets generate a list of all Data FlowSets in the packet. If matched
// with appropriate templates Data FlowSets can be decoded to Data Records or
// Options Data Records.
func (p *Packet) DataFlowSets() []*Set {
	return p.FlowSets
}
