//go:build !test
// +build !test

package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	tests := []struct {
		testNum  int
		input    []byte
		wantFail bool
		expected interface{}
	}{
		{
			// Proper packet
			testNum: 1,
			input: []byte{
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, // Marker
				0, 19, // Length
				4, // Type = Keepalive

			},
			wantFail: false,
		},
		{
			// Invalid marker
			testNum: 2,
			input: []byte{
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2, // Marker
				0, 19, // Length
				4, // Type = Keepalive

			},
			wantFail: true,
		},
		{
			// Proper NOTIFICATION packet
			testNum: 3,
			input: []byte{
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, // Marker
				0, 21, // Length
				3,    // Type = Notification
				1, 1, // Message Header Error, Connection Not Synchronized.
			},
			wantFail: false,
		},
		{
			// Proper OPEN packet
			testNum: 4,
			input: []byte{
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, // Marker
				0, 29, // Length
				1,      // Type = Open
				4,      // Version
				0, 200, //ASN,
				0, 15, // Holdtime
				10, 20, 30, 40, // BGP Identifier
				0, // Opt Parm Len
			},
			wantFail: false,
		},
		{
			// Incomplete OPEN packet
			testNum: 5,
			input: []byte{
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, // Marker
				0, 28, // Length
				1,      // Type = Open
				4,      // Version
				0, 200, //ASN,
				0, 15, // Holdtime
				0, 0, 0, 100, // BGP Identifier
			},
			wantFail: true,
		},
		{
			testNum: 6,
			input: []byte{
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, // Marker
				0, 28, // Length
				2,                               // Type = Update
				0, 5, 8, 10, 16, 192, 168, 0, 0, // 2 withdraws
			},
			wantFail: false,
		},
		{
			testNum: 7,
			input: []byte{
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, // Marker
				0, 28, // Length
				5,                               // Type = Invalid
				0, 5, 8, 10, 16, 192, 168, 0, 0, // Some more stuff
			},
			wantFail: true,
		},
	}
	for i, t := range tests {
		f, err := os.Create(fmt.Sprintf("corpus/%v.bytes", i))
		if err != nil {
			log.Fatalf(err.Error())
		}
		f.Write(t.input)
		f.Close()
	}
}
