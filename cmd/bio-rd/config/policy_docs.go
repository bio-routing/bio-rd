// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
// DO NOT EDIT: this file is automatically generated by docgen
package config

import (
	"github.com/projectdiscovery/yamldoc-go/encoder"
)

var (
	PolicyOptionsDoc           encoder.Doc
	PrefixListDoc              encoder.Doc
	PolicyStatementDoc         encoder.Doc
	PolicyStatementTermDoc     encoder.Doc
	PolicyStatementTermFromDoc encoder.Doc
	RouteFilterDoc             encoder.Doc
	PolicyStatementTermThenDoc encoder.Doc
	ASPathPrependDoc           encoder.Doc
	NextHopDoc                 encoder.Doc
)

func init() {
	PolicyOptionsDoc.Type = "PolicyOptions"
	PolicyOptionsDoc.Comments[encoder.LineComment] = ""
	PolicyOptionsDoc.Description = ""
	PolicyOptionsDoc.Fields = make([]encoder.Doc, 2)
	PolicyOptionsDoc.Fields[0].Name = "policy_statements"
	PolicyOptionsDoc.Fields[0].Type = "[]PolicyStatement"
	PolicyOptionsDoc.Fields[0].Note = ""
	PolicyOptionsDoc.Fields[0].Description = "Policy statements to filter route imports and exports\nExample:\n- name: \"PeerA-In\"\n    terms:\n      - name: \"Reject_certain_stuff\"\n        from:\n          route_filters:\n             - prefix: \"198.51.100.0/24\"\n               matcher: \"orlonger\"\n             - prefix: \"203.0.113.0/25\"\n               matcher: \"exact\"\n             - prefix: \"203.0.113.128/25\"\n               matcher: \"exact\""
	PolicyOptionsDoc.Fields[0].Comments[encoder.LineComment] = "Policy statements to filter route imports and exports"
	PolicyOptionsDoc.Fields[1].Name = "prefix_lists"
	PolicyOptionsDoc.Fields[1].Type = "[]PrefixList"
	PolicyOptionsDoc.Fields[1].Note = ""
	PolicyOptionsDoc.Fields[1].Description = "prefix lists to be used in the policy statements.\nExample:\n  prefix-lists:\n    prefixes:\n      - 2001:db8:0:1::/64"
	PolicyOptionsDoc.Fields[1].Comments[encoder.LineComment] = "prefix lists to be used in the policy statements."

	PrefixListDoc.Type = "PrefixList"
	PrefixListDoc.Comments[encoder.LineComment] = ""
	PrefixListDoc.Description = ""
	PrefixListDoc.AppearsIn = []encoder.Appearance{
		{
			TypeName:  "PolicyOptions",
			FieldName: "prefix_lists",
		},
	}
	PrefixListDoc.Fields = make([]encoder.Doc, 1)
	PrefixListDoc.Fields[0].Name = "prefixes"
	PrefixListDoc.Fields[0].Type = "[]string"
	PrefixListDoc.Fields[0].Note = ""
	PrefixListDoc.Fields[0].Description = "List of prefixes"
	PrefixListDoc.Fields[0].Comments[encoder.LineComment] = "List of prefixes"

	PolicyStatementDoc.Type = "PolicyStatement"
	PolicyStatementDoc.Comments[encoder.LineComment] = ""
	PolicyStatementDoc.Description = ""
	PolicyStatementDoc.AppearsIn = []encoder.Appearance{
		{
			TypeName:  "PolicyOptions",
			FieldName: "policy_statements",
		},
	}
	PolicyStatementDoc.Fields = make([]encoder.Doc, 2)
	PolicyStatementDoc.Fields[0].Name = "name"
	PolicyStatementDoc.Fields[0].Type = "string"
	PolicyStatementDoc.Fields[0].Note = ""
	PolicyStatementDoc.Fields[0].Description = "Name of the policy statement"
	PolicyStatementDoc.Fields[0].Comments[encoder.LineComment] = "Name of the policy statement"
	PolicyStatementDoc.Fields[1].Name = "terms"
	PolicyStatementDoc.Fields[1].Type = "[]PolicyStatementTerm"
	PolicyStatementDoc.Fields[1].Note = ""
	PolicyStatementDoc.Fields[1].Description = "List of terms defining the policy (see example above)"
	PolicyStatementDoc.Fields[1].Comments[encoder.LineComment] = "List of terms defining the policy (see example above)"

	PolicyStatementTermDoc.Type = "PolicyStatementTerm"
	PolicyStatementTermDoc.Comments[encoder.LineComment] = ""
	PolicyStatementTermDoc.Description = ""
	PolicyStatementTermDoc.AppearsIn = []encoder.Appearance{
		{
			TypeName:  "PolicyStatement",
			FieldName: "terms",
		},
	}
	PolicyStatementTermDoc.Fields = make([]encoder.Doc, 3)
	PolicyStatementTermDoc.Fields[0].Name = "name"
	PolicyStatementTermDoc.Fields[0].Type = "string"
	PolicyStatementTermDoc.Fields[0].Note = ""
	PolicyStatementTermDoc.Fields[0].Description = "Name of the term"
	PolicyStatementTermDoc.Fields[0].Comments[encoder.LineComment] = "Name of the term"
	PolicyStatementTermDoc.Fields[1].Name = "from"
	PolicyStatementTermDoc.Fields[1].Type = "PolicyStatementTermFrom"
	PolicyStatementTermDoc.Fields[1].Note = ""
	PolicyStatementTermDoc.Fields[1].Description = "Filter to match the term"
	PolicyStatementTermDoc.Fields[1].Comments[encoder.LineComment] = "Filter to match the term"
	PolicyStatementTermDoc.Fields[2].Name = "then"
	PolicyStatementTermDoc.Fields[2].Type = "PolicyStatementTermThen"
	PolicyStatementTermDoc.Fields[2].Note = ""
	PolicyStatementTermDoc.Fields[2].Description = "Action to execute if the filter matches\nAvailable actions are:\n  - Accept: accepts the route without modifications\n  - Reject: rejects the route\n  - MED: sets the MED to the specified value (max 4294967295)\n  - LocalPref sets the local preference to the specified value (max 4294967295)\n  - AsPathPrepend: prepends AS numbers to the route. Details bellow\n  - NextHop: modify the next-hop to the specified address"
	PolicyStatementTermDoc.Fields[2].Comments[encoder.LineComment] = "Action to execute if the filter matches"

	PolicyStatementTermFromDoc.Type = "PolicyStatementTermFrom"
	PolicyStatementTermFromDoc.Comments[encoder.LineComment] = ""
	PolicyStatementTermFromDoc.Description = ""
	PolicyStatementTermFromDoc.AppearsIn = []encoder.Appearance{
		{
			TypeName:  "PolicyStatementTerm",
			FieldName: "from",
		},
	}
	PolicyStatementTermFromDoc.Fields = make([]encoder.Doc, 1)
	PolicyStatementTermFromDoc.Fields[0].Name = "route_filters"
	PolicyStatementTermFromDoc.Fields[0].Type = "[]RouteFilter"
	PolicyStatementTermFromDoc.Fields[0].Note = ""
	PolicyStatementTermFromDoc.Fields[0].Description = "List of route filters to match incoming packets\nExample:\n  route_filters:\n     - prefix: \"198.51.100.0/24\"\n       matcher: \"orlonger\""
	PolicyStatementTermFromDoc.Fields[0].Comments[encoder.LineComment] = "List of route filters to match incoming packets"

	RouteFilterDoc.Type = "RouteFilter"
	RouteFilterDoc.Comments[encoder.LineComment] = ""
	RouteFilterDoc.Description = ""
	RouteFilterDoc.AppearsIn = []encoder.Appearance{
		{
			TypeName:  "PolicyStatementTermFrom",
			FieldName: "route_filters",
		},
	}
	RouteFilterDoc.Fields = make([]encoder.Doc, 4)
	RouteFilterDoc.Fields[0].Name = "prefix"
	RouteFilterDoc.Fields[0].Type = "string"
	RouteFilterDoc.Fields[0].Note = ""
	RouteFilterDoc.Fields[0].Description = "Prefix to match. Defined in CIDR notation"
	RouteFilterDoc.Fields[0].Comments[encoder.LineComment] = "Prefix to match. Defined in CIDR notation"
	RouteFilterDoc.Fields[1].Name = "matcher"
	RouteFilterDoc.Fields[1].Type = "string"
	RouteFilterDoc.Fields[1].Note = ""
	RouteFilterDoc.Fields[1].Description = "Qualifier for the filter.\nAvailable options:\n  - exact: matches only the exact prefix\n  - orlonger: matches if the prefix equals the filter, or has a longer prefix length\n  - longer: matches if the route has a longer prefix length than the one defined in the filter\n  - range: the route falls between the values defined in len_min and len_max"
	RouteFilterDoc.Fields[1].Comments[encoder.LineComment] = "Qualifier for the filter."
	RouteFilterDoc.Fields[2].Name = "len_min"
	RouteFilterDoc.Fields[2].Type = "uint8"
	RouteFilterDoc.Fields[2].Note = ""
	RouteFilterDoc.Fields[2].Description = "minimum length of the range"
	RouteFilterDoc.Fields[2].Comments[encoder.LineComment] = "minimum length of the range"
	RouteFilterDoc.Fields[3].Name = "len_max"
	RouteFilterDoc.Fields[3].Type = "uint8"
	RouteFilterDoc.Fields[3].Note = ""
	RouteFilterDoc.Fields[3].Description = "maximum lange of the range"
	RouteFilterDoc.Fields[3].Comments[encoder.LineComment] = "maximum lange of the range"

	PolicyStatementTermThenDoc.Type = "PolicyStatementTermThen"
	PolicyStatementTermThenDoc.Comments[encoder.LineComment] = ""
	PolicyStatementTermThenDoc.Description = ""
	PolicyStatementTermThenDoc.AppearsIn = []encoder.Appearance{
		{
			TypeName:  "PolicyStatementTerm",
			FieldName: "then",
		},
	}
	PolicyStatementTermThenDoc.Fields = make([]encoder.Doc, 6)
	PolicyStatementTermThenDoc.Fields[0].Name = "accept"
	PolicyStatementTermThenDoc.Fields[0].Type = "bool"
	PolicyStatementTermThenDoc.Fields[0].Note = ""
	PolicyStatementTermThenDoc.Fields[0].Description = "accept the route"
	PolicyStatementTermThenDoc.Fields[0].Comments[encoder.LineComment] = "accept the route"
	PolicyStatementTermThenDoc.Fields[1].Name = "reject"
	PolicyStatementTermThenDoc.Fields[1].Type = "bool"
	PolicyStatementTermThenDoc.Fields[1].Note = ""
	PolicyStatementTermThenDoc.Fields[1].Description = "reject the route"
	PolicyStatementTermThenDoc.Fields[1].Comments[encoder.LineComment] = "reject the route"
	PolicyStatementTermThenDoc.Fields[2].Name = "med"
	PolicyStatementTermThenDoc.Fields[2].Type = "uint32"
	PolicyStatementTermThenDoc.Fields[2].Note = ""
	PolicyStatementTermThenDoc.Fields[2].Description = "Multi-exit discriminator"
	PolicyStatementTermThenDoc.Fields[2].Comments[encoder.LineComment] = "Multi-exit discriminator"
	PolicyStatementTermThenDoc.Fields[3].Name = "local_pref"
	PolicyStatementTermThenDoc.Fields[3].Type = "uint32"
	PolicyStatementTermThenDoc.Fields[3].Note = ""
	PolicyStatementTermThenDoc.Fields[3].Description = "Local preference"
	PolicyStatementTermThenDoc.Fields[3].Comments[encoder.LineComment] = "Local preference"
	PolicyStatementTermThenDoc.Fields[4].Name = "as_path_prepend"
	PolicyStatementTermThenDoc.Fields[4].Type = "ASPathPrepend"
	PolicyStatementTermThenDoc.Fields[4].Note = ""
	PolicyStatementTermThenDoc.Fields[4].Description = "ASN to prepend.\n  Values:\n    - ASN: asn number to prepend\n    - count: amount of times that the number will be prepended"
	PolicyStatementTermThenDoc.Fields[4].Comments[encoder.LineComment] = "ASN to prepend."
	PolicyStatementTermThenDoc.Fields[5].Name = "next_hop"
	PolicyStatementTermThenDoc.Fields[5].Type = "NextHop"
	PolicyStatementTermThenDoc.Fields[5].Note = ""
	PolicyStatementTermThenDoc.Fields[5].Description = "IP address to be used as a next-hop for the route"
	PolicyStatementTermThenDoc.Fields[5].Comments[encoder.LineComment] = "IP address to be used as a next-hop for the route"

	ASPathPrependDoc.Type = "ASPathPrepend"
	ASPathPrependDoc.Comments[encoder.LineComment] = ""
	ASPathPrependDoc.Description = ""
	ASPathPrependDoc.AppearsIn = []encoder.Appearance{
		{
			TypeName:  "PolicyStatementTermThen",
			FieldName: "as_path_prepend",
		},
	}
	ASPathPrependDoc.Fields = make([]encoder.Doc, 2)
	ASPathPrependDoc.Fields[0].Name = "asn"
	ASPathPrependDoc.Fields[0].Type = "uint32"
	ASPathPrependDoc.Fields[0].Note = ""
	ASPathPrependDoc.Fields[0].Description = "AS number"
	ASPathPrependDoc.Fields[0].Comments[encoder.LineComment] = "AS number"
	ASPathPrependDoc.Fields[1].Name = "count"
	ASPathPrependDoc.Fields[1].Type = "uint16"
	ASPathPrependDoc.Fields[1].Note = ""
	ASPathPrependDoc.Fields[1].Description = "times to prepend"
	ASPathPrependDoc.Fields[1].Comments[encoder.LineComment] = "times to prepend"

	NextHopDoc.Type = "NextHop"
	NextHopDoc.Comments[encoder.LineComment] = ""
	NextHopDoc.Description = ""
	NextHopDoc.AppearsIn = []encoder.Appearance{
		{
			TypeName:  "PolicyStatementTermThen",
			FieldName: "next_hop",
		},
	}
	NextHopDoc.Fields = make([]encoder.Doc, 1)
	NextHopDoc.Fields[0].Name = "address"
	NextHopDoc.Fields[0].Type = "string"
	NextHopDoc.Fields[0].Note = ""
	NextHopDoc.Fields[0].Description = "IP address to be used as next hop"
	NextHopDoc.Fields[0].Comments[encoder.LineComment] = "IP address to be used as next hop"
}

func (_ PolicyOptions) Doc() *encoder.Doc {
	return &PolicyOptionsDoc
}

func (_ PrefixList) Doc() *encoder.Doc {
	return &PrefixListDoc
}

func (_ PolicyStatement) Doc() *encoder.Doc {
	return &PolicyStatementDoc
}

func (_ PolicyStatementTerm) Doc() *encoder.Doc {
	return &PolicyStatementTermDoc
}

func (_ PolicyStatementTermFrom) Doc() *encoder.Doc {
	return &PolicyStatementTermFromDoc
}

func (_ RouteFilter) Doc() *encoder.Doc {
	return &RouteFilterDoc
}

func (_ PolicyStatementTermThen) Doc() *encoder.Doc {
	return &PolicyStatementTermThenDoc
}

func (_ ASPathPrepend) Doc() *encoder.Doc {
	return &ASPathPrependDoc
}

func (_ NextHop) Doc() *encoder.Doc {
	return &NextHopDoc
}

// GetpolicyDoc returns documentation for the file cmd/bio-rd/config/policy_docs.go.
func GetpolicyDoc() *encoder.FileDoc {
	return &encoder.FileDoc{
		Name:        "policy",
		Description: "",
		Structs: []*encoder.Doc{
			&PolicyOptionsDoc,
			&PrefixListDoc,
			&PolicyStatementDoc,
			&PolicyStatementTermDoc,
			&PolicyStatementTermFromDoc,
			&RouteFilterDoc,
			&PolicyStatementTermThenDoc,
			&ASPathPrependDoc,
			&NextHopDoc,
		},
	}
}
