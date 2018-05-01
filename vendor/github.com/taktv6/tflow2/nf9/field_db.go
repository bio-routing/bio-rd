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

const (
	InBytes                   = 1
	InPkts                    = 2
	Flows                     = 3
	Protocol                  = 4
	SrcTos                    = 5
	TCPFlags                  = 6
	L4SrcPort                 = 7
	IPv4SrcAddr               = 8
	SrcMask                   = 9
	InputSnmp                 = 10
	L4DstPort                 = 11
	IPv4DstAddr               = 12
	DstMask                   = 13
	OutputSnmp                = 14
	IPv4NextHop               = 15
	SrcAs                     = 16
	DstAs                     = 17
	BGPIPv4NextHop            = 18
	MulDstPkts                = 19
	MulDstBytes               = 20
	LastSwitched              = 21
	FirstSwitched             = 22
	OutBytes                  = 23
	OutPkts                   = 24
	MinPktLngth               = 25
	MaxPktLngth               = 26
	IPv6SrcAddr               = 27
	IPv6DstAddr               = 28
	IPv6SrcMask               = 29
	IPv6DstMask               = 30
	IPv6FlowLabel             = 31
	IcmpType                  = 32
	MulIgmpType               = 33
	SamplingInterval          = 34
	SamplingAlgorithm         = 35
	FlowActiveTimeout         = 36
	FlowInactiveTimeout       = 37
	EngineType                = 38
	EngineID                  = 39
	TotalBytesExp             = 40
	TotalPktsExp              = 41
	TotalFlowsExp             = 42
	VendorProprietary43       = 43
	IPv4SrcPrefix             = 44
	IPv4DstPrefix             = 45
	MplsTopLabelType          = 46
	MplsTopLabelIPAddr        = 47
	FlowSamplerID             = 48
	FlowSamplerMode           = 49
	FlowSamplerRandomInterval = 50
	VendorProprietary51       = 51
	MinTTL                    = 52
	MaxTTL                    = 53
	IPv4Ident                 = 54
	DstTos                    = 55
	InSrcMac                  = 56
	OutDstMac                 = 57
	SrcVlan                   = 58
	DstVlan                   = 59
	IPProtocolVersion         = 60
	Direction                 = 61
	IPv6NextHop               = 62
	BgpIPv6NextHop            = 63
	IPv6OptionsHeaders        = 64
	VendorProprietary65       = 65
	VendorProprietary66       = 66
	VendorProprietary67       = 67
	VendorProprietary68       = 68
	VendorProprietary69       = 69
	MplsLabel1                = 70
	MplsLabel2                = 71
	MplsLabel3                = 72
	MplsLabel4                = 73
	MplsLabel5                = 74
	MplsLabel6                = 75
	MplsLabel7                = 76
	MplsLabel8                = 77
	MplsLabel9                = 78
	MplsLabel10               = 79
	InDstMac                  = 80
	OutSrcMac                 = 81
	IfName                    = 82
	IfDesc                    = 83
	SamplerName               = 84
	InPermanentBytes          = 85
	InPermanentPkts           = 86
	VendorProprietary87       = 87
	FragmentOffset            = 88
	ForwardingStatus          = 89
	MplsPalRd                 = 90
	MplsPrefixLen             = 91
	SrcTrafficIndex           = 92
	DstTrafficIndex           = 93
	ApplicationDescription    = 94
	ApplicationTag            = 95
	ApplicationName           = 96
)
