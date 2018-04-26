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

package sflow

import (
	"fmt"
	"net"
	"unsafe"

	"github.com/golang/glog"
	"github.com/taktv6/tflow2/convert"
)

const (
	dataFlowSample     = 1
	dataCounterSample  = 2
	standardSflow      = 0
	rawPacketHeader    = 1
	extendedSwitchData = 1001
	extendedRouterData = 1002
)

// errorIncompatibleVersion prints an error message in case the detected version is not supported
func errorIncompatibleVersion(version uint32) error {
	return fmt.Errorf("Sflow: Incompatible protocol version v%d, only v5 is supported", version)
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
	//bufferMinPtr := unsafe.Pointer(uintptr(bufferPtr) + uintptr(bufSize) - uintptr(pSize))
	headerPtr := unsafe.Pointer(uintptr(bufferPtr) + uintptr(bufSize) - uintptr(sizeOfHeaderTop))

	var p Packet
	p.Buffer = buffer[:]
	p.headerTop = (*headerTop)(headerPtr)

	if p.headerTop.Version != 5 {
		return nil, errorIncompatibleVersion(p.Header.Version)
	}

	agentAddressLen := uint64(0)
	switch p.headerTop.AgentAddressType {
	default:
		return nil, fmt.Errorf("Unknown AgentAddressType %d", p.headerTop.AgentAddressType)
	case 1:
		agentAddressLen = 4
	case 2:
		agentAddressLen = 16
	}

	headerBottomPtr := unsafe.Pointer(uintptr(bufferPtr) + uintptr(bufSize) - uintptr(sizeOfHeaderTop) - uintptr(agentAddressLen) - uintptr(sizeOfHeaderBottom))
	p.headerBottom = (*headerBottom)(headerBottomPtr)

	h := Header{
		Version:          p.headerTop.Version,
		AgentAddressType: p.headerTop.AgentAddressType,
		AgentAddress:     getNetIP(headerPtr, agentAddressLen),
		SubAgentID:       p.headerBottom.SubAgentID,
		SequenceNumber:   p.headerBottom.SequenceNumber,
		SysUpTime:        p.headerBottom.SysUpTime,
		NumSamples:       p.headerBottom.NumSamples,
	}
	p.Header = &h

	flowSamples, err := decodeFlows(headerBottomPtr, h.NumSamples)
	if err != nil {
		return nil, fmt.Errorf("Unable to dissect flows: %v", err)
	}
	p.FlowSamples = flowSamples

	return &p, nil
}

func extractEnterpriseFormat(sfType uint32) (sfTypeEnterprise uint32, sfTypeFormat uint32) {
	return sfType >> 12, sfType & 0xfff
}

func decodeFlows(samplesPtr unsafe.Pointer, NumSamples uint32) ([]*FlowSample, error) {
	flowSamples := make([]*FlowSample, 0)
	for i := uint32(0); i < NumSamples; i++ {
		sfTypeEnterprise, sfTypeFormat := extractEnterpriseFormat(*(*uint32)(unsafe.Pointer(uintptr(samplesPtr) - uintptr(4))))

		if sfTypeEnterprise != 0 {
			return nil, fmt.Errorf("Unknown Enterprise: %d", sfTypeEnterprise)
		}

		sampleLengthPtr := unsafe.Pointer(uintptr(samplesPtr) - uintptr(8))
		sampleLength := *(*uint32)(sampleLengthPtr)

		if sfTypeFormat == dataFlowSample {
			fs, err := decodeFlowSample(samplesPtr)
			if err != nil {
				return nil, fmt.Errorf("Unable to decode flow sample: %v", err)
			}
			flowSamples = append(flowSamples, fs)
		}

		samplesPtr = unsafe.Pointer(uintptr(samplesPtr) - uintptr(sampleLength+8))
	}

	return flowSamples, nil
}

func decodeFlowSample(flowSamplePtr unsafe.Pointer) (*FlowSample, error) {
	flowSamplePtr = unsafe.Pointer(uintptr(flowSamplePtr) - uintptr(sizeOfFlowSampleHeader))
	fsh := (*FlowSampleHeader)(flowSamplePtr)

	var rph *RawPacketHeader
	var rphd unsafe.Pointer
	var erd *ExtendedRouterData

	for i := uint32(0); i < fsh.FlowRecord; i++ {
		sfTypeEnterprise, sfTypeFormat := extractEnterpriseFormat(*(*uint32)(unsafe.Pointer(uintptr(flowSamplePtr) - uintptr(4))))
		flowDataLength := *(*uint32)(unsafe.Pointer(uintptr(flowSamplePtr) - uintptr(8)))

		if sfTypeEnterprise == standardSflow {
			var err error
			switch sfTypeFormat {
			case rawPacketHeader:
				rph = decodeRawPacketHeader(flowSamplePtr)
				rphd = unsafe.Pointer(uintptr(flowSamplePtr) - sizeOfRawPacketHeader)

			case extendedRouterData:
				erd, err = decodeExtendRouterData(flowSamplePtr)
				if err != nil {
					return nil, fmt.Errorf("Unable to decide extended router data: %v", err)
				}

			case extendedSwitchData:

			default:
				glog.Infof("Unknown sfTypeFormat\n")
			}

		}

		flowSamplePtr = unsafe.Pointer(uintptr(flowSamplePtr) - uintptr(8) - uintptr(flowDataLength))
	}

	fs := &FlowSample{
		FlowSampleHeader:    fsh,
		RawPacketHeader:     rph,
		RawPacketHeaderData: rphd,
		ExtendedRouterData:  erd,
	}

	return fs, nil
}

func decodeRawPacketHeader(rphPtr unsafe.Pointer) *RawPacketHeader {
	rphPtr = unsafe.Pointer(uintptr(rphPtr) - uintptr(sizeOfRawPacketHeader))
	rph := (*RawPacketHeader)(rphPtr)
	return rph
}

func decodeExtendRouterData(erhPtr unsafe.Pointer) (*ExtendedRouterData, error) {
	erhTopPtr := unsafe.Pointer(uintptr(erhPtr) - uintptr(sizeOfextendedRouterDataTop))
	erhTop := (*extendedRouterDataTop)(erhTopPtr)

	addressLen := uint64(0)
	switch erhTop.AddressType {
	default:
		return nil, fmt.Errorf("Unknown AgentAddressType %d", erhTop.AddressType)
	case 1:
		addressLen = 4
	case 2:
		addressLen = 16
	}

	erhBottomPtr := unsafe.Pointer(uintptr(erhTopPtr) - uintptr(sizeOfextendedRouterDataBottom) - uintptr(addressLen) - uintptr(sizeOfextendedRouterDataBottom))
	erhBottom := (*extendedRouterDataBottom)(erhBottomPtr)

	return &ExtendedRouterData{
		EnterpriseType:         erhTop.EnterpriseType,
		FlowDataLength:         erhTop.FlowDataLength,
		AddressType:            erhTop.AddressType,
		NextHop:                getNetIP(unsafe.Pointer(uintptr(erhTopPtr)), addressLen),
		NextHopSourceMask:      erhBottom.NextHopSourceMask,
		NextHopDestinationMask: erhBottom.NextHopDestinationMask,
	}, nil
}

func getNetIP(headerPtr unsafe.Pointer, addressLen uint64) net.IP {
	ptr := unsafe.Pointer(uintptr(headerPtr) - uintptr(1))
	addr := make([]byte, addressLen)
	for i := uint64(0); i < addressLen; i++ {
		addr[i] = *(*byte)(unsafe.Pointer(uintptr(ptr) - uintptr(i)))
	}

	return net.IP(addr)
}
