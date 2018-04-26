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

// Package nf9 provides structures and functions to decode and analyze
// NetFlow v9 packets.
//
// This package does only packet decoding in a single packet context. It keeps
// no state when decoding multiple packets. As a result Data FlowSets can not be
// decoded during initial packet decoding. To decode Data FlowSets user must
// keep track of Template Records and Options Template Records manually.
//
// Examples of NetFlow v9 packets:
//
//   +--------+--------------------------------------------------------+
//   |        | +----------+ +---------+     +-----------+ +---------+ |
//   | Packet | | Template | | Data    |     | Options   | | Data    | |
//   | Header | | FlowSet  | | FlowSet | ... | Template  | | FlowSet | |
//   |        | |          | |         |     | FlowSet   | |         | |
//   |        | +----------+ +---------+     +-----------+ +---------+ |
//   +--------+--------------------------------------------------------+
//
//   +--------+----------------------------------------------+
//   |        | +---------+     +---------+      +---------+ |
//   | Packet | | Data    | ... | Data    | ...  | Data    | |
//   | Header | | FlowSet | ... | FlowSet | ...  | FlowSet | |
//   |        | +---------+     +---------+      +---------+ |
//   +--------+----------------------------------------------+
//
//   +--------+-------------------------------------------------+
//   |        | +----------+     +----------+      +----------+ |
//   | Packet | | Template |     | Template |      | Options  | |
//   | Header | | FlowSet  | ... | FlowSet  | ...  | Template | |
//   |        | |          |     |          |      | FlowSet  | |
//   |        | +----------+     +----------+      +----------+ |
//   +--------+-------------------------------------------------+
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
// Most of structure names and comments are taken directly from RFC 3954.
// Reading the NetFlow v9 protocol specification is highly recommended before
// using this package.
package nf9

import "unsafe"

// Header is the NetFlow version 9 header
type Header struct {
	// A 32-bit value that identifies the Exporter Observation Domain.
	SourceID uint32

	// Incremental sequence counter of all Export Packets sent from the
	// current Observation Domain by the Exporter.
	//SequenceNumber uint32
	SequenceNumber uint32

	// Time in seconds since 0000 UTC 1970, at which the Export Packet
	// leaves the Exporter.
	//UnixSecs uint32
	UnixSecs uint32

	// Time in milliseconds since this device was first booted.
	//SysUpTime uint32
	SysUpTime uint32

	// The total number of records in the Export Packet, which is the sum
	// of Options FlowSet records, Template FlowSet records, and Data
	// FlowSet records.
	//Count uint16
	Count uint16

	// Version of Flow Record format exported in this packet. The value of
	//this field is 9 for the current version.
	//Version uint16
	Version uint16
}

// FlowSet represents a FlowSet as described in RFC3954
type FlowSet struct {
	Header *FlowSetHeader
	Flows  []byte
}

// FlowSetHeader is a decoded representation of the header of a FlowSet
type FlowSetHeader struct {
	Length    uint16
	FlowSetID uint16
}

var sizeOfFlowSetHeader = unsafe.Sizeof(FlowSetHeader{})

// Packet is a decoded representation of a single NetFlow v9 UDP packet.
type Packet struct {
	// A pointer to the packets headers
	Header *Header

	// A slice of pointers to FlowSet. Each element is instance of (Data)FlowSet
	// found in this packet
	FlowSets []*FlowSet

	// A slice of pointers to TemplateRecords. Each element is instance of TemplateRecords
	// representing a template found in this packet.
	Templates []*TemplateRecords

	// Buffer is a slice pointing to the original byte array that this packet was decoded from.
	// This field is only populated if debug level is at least 2
	Buffer []byte
}

var sizeOfHeader = unsafe.Sizeof(Header{})

// GetTemplateRecords returns a list of all Template Records in the packet.
// Template Records can be used to decode Data FlowSets to Data Records.
func (p *Packet) GetTemplateRecords() []*TemplateRecords {
	return p.Templates
}

// DataFlowSets generate a list of all Data FlowSets in the packet. If matched
// with appropriate templates Data FlowSets can be decoded to Data Records or
// Options Data Records.
func (p *Packet) DataFlowSets() []*FlowSet {
	return p.FlowSets
}
