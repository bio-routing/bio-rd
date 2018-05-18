package packet

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/bio-routing/bio-rd/net"
	"github.com/stretchr/testify/assert"
	"github.com/taktv6/tflow2/convert"
)

type test struct {
	testNum  int
	input    []byte
	wantFail bool
	expected interface{}
}

type decodeFunc func(*bytes.Buffer) (interface{}, error)

func BenchmarkDecodeUpdateMsg(b *testing.B) {
	input := []byte{0, 5, 8, 10, 16, 192, 168,
		0, 53, // Total Path Attribute Length

		255,  // Attribute flags
		1,    // Attribute Type code (ORIGIN)
		0, 1, // Length
		2, // INCOMPLETE

		0,      // Attribute flags
		2,      // Attribute Type code (AS Path)
		12,     // Length
		2,      // Type = AS_SEQUENCE
		2,      // Path Segement Length
		59, 65, // AS15169
		12, 248, // AS3320
		1,      // Type = AS_SET
		2,      // Path Segement Length
		59, 65, // AS15169
		12, 248, // AS3320

		0,              // Attribute flags
		3,              // Attribute Type code (Next Hop)
		4,              // Length
		10, 11, 12, 13, // Next Hop

		0,          // Attribute flags
		4,          // Attribute Type code (MED)
		4,          // Length
		0, 0, 1, 0, // MED 256

		0,          // Attribute flags
		5,          // Attribute Type code (Local Pref)
		4,          // Length
		0, 0, 1, 0, // Local Pref 256

		0, // Attribute flags
		6, // Attribute Type code (Atomic Aggregate)
		0, // Length

		0,    // Attribute flags
		7,    // Attribute Type code (Atomic Aggregate)
		6,    // Length
		1, 2, // ASN
		10, 11, 12, 13, // Address

		8, 11, // 11.0.0.0/8
	}

	for i := 0; i < b.N; i++ {
		buf := bytes.NewBuffer(input)
		_, err := decodeUpdateMsg(buf, uint16(len(input)))
		if err != nil {
			fmt.Printf("decodeUpdateMsg failed: %v\n", err)
		}
		//buf.Next(1)
	}
}

func TestDecode(t *testing.T) {
	tests := []test{
		{
			// Proper packet
			testNum: 1,
			input: []byte{
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, // Marker
				0, 19, // Length
				4, // Type = Keepalive

			},
			wantFail: false,
			expected: &BGPMessage{
				Header: &BGPHeader{
					Length: 19,
					Type:   4,
				},
			},
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
			expected: &BGPMessage{
				Header: &BGPHeader{
					Length: 19,
					Type:   4,
				},
			},
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
			expected: &BGPMessage{
				Header: &BGPHeader{
					Length: 21,
					Type:   3,
				},
				Body: &BGPNotification{
					ErrorCode:    1,
					ErrorSubcode: 1,
				},
			},
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
			expected: &BGPMessage{
				Header: &BGPHeader{
					Length: 29,
					Type:   1,
				},
				Body: &BGPOpen{
					Version:       4,
					AS:            200,
					HoldTime:      15,
					BGPIdentifier: uint32(169090600),
					OptParmLen:    0,
					OptParams:     []OptParam{},
				},
			},
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
			expected: &BGPMessage{
				Header: &BGPHeader{
					Length: 28,
					Type:   1,
				},
				Body: &BGPOpen{
					Version:       4,
					AS:            200,
					HoldTime:      15,
					BGPIdentifier: uint32(100),
				},
			},
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
			expected: &BGPMessage{
				Header: &BGPHeader{
					Length: 28,
					Type:   2,
				},
				Body: &BGPUpdate{
					WithdrawnRoutesLen: 5,
					WithdrawnRoutes: &NLRI{
						IP:     strAddr("10.0.0.0"),
						Pfxlen: 8,
						Next: &NLRI{
							IP:     strAddr("192.168.0.0"),
							Pfxlen: 16,
						},
					},
				},
			},
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

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		msg, err := Decode(buf)

		if err != nil && !test.wantFail {
			t.Errorf("Unexpected error in test %d: %v", test.testNum, err)
			continue
		}

		if err == nil && test.wantFail {
			t.Errorf("Expected error did not happen in test %d", test.testNum)
			continue
		}

		if err != nil && test.wantFail {
			continue
		}

		if msg == nil {
			t.Errorf("Unexpected nil result in test %d. Expected: %v", test.testNum, test.expected)
			continue
		}

		assert.Equal(t, test.expected, msg)
	}
}

func TestDecodeNotificationMsg(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantFail bool
		expected interface{}
	}{
		{
			name:     "Invalid error code",
			input:    []byte{0, 0},
			wantFail: true,
		},
		{
			name:     "Invalid error code #2",
			input:    []byte{7, 0},
			wantFail: true,
		},
		{
			name:     "Invalid ErrSubCode (Header)",
			input:    []byte{1, 0},
			wantFail: true,
		},
		{
			name:     "Invalid ErrSubCode (Header) #2",
			input:    []byte{1, 4},
			wantFail: true,
		},
		{
			name:     "Invalid ErrSubCode (Open)",
			input:    []byte{2, 0},
			wantFail: true,
		},
		{
			name:     "Invalid ErrSubCode (Open) #2",
			input:    []byte{2, 7},
			wantFail: true,
		},
		{
			name:     "Invalid ErrSubCode (Open) #3",
			input:    []byte{2, 5},
			wantFail: true,
		},
		{
			name:     "Invalid ErrSubCode (Update)",
			input:    []byte{3, 0},
			wantFail: true,
		},
		{
			name:     "Invalid ErrSubCode (Update) #2",
			input:    []byte{3, 12},
			wantFail: true,
		},
		{
			name:     "Invalid ErrSubCode (Update) #3",
			input:    []byte{3, 7},
			wantFail: true,
		},
		{
			name:     "Valid Notification",
			input:    []byte{2, 2},
			wantFail: false,
			expected: &BGPNotification{
				ErrorCode:    2,
				ErrorSubcode: 2,
			},
		},
		{
			name:     "Empty input",
			input:    []byte{},
			wantFail: true,
		},
		{
			name:     "Hold Timer Expired",
			input:    []byte{4, 0},
			wantFail: false,
			expected: &BGPNotification{
				ErrorCode:    4,
				ErrorSubcode: 0,
			},
		},
		{
			name:     "Hold Timer Expired (invalid subcode)",
			input:    []byte{4, 1},
			wantFail: true,
		},
		{
			name:     "FSM Error",
			input:    []byte{5, 0},
			wantFail: false,
			expected: &BGPNotification{
				ErrorCode:    5,
				ErrorSubcode: 0,
			},
		},
		{
			name:     "FSM Error (invalid subcode)",
			input:    []byte{5, 1},
			wantFail: true,
		},
		{
			name:     "Cease",
			input:    []byte{6, 0},
			wantFail: false,
			expected: &BGPNotification{
				ErrorCode:    6,
				ErrorSubcode: 0,
			},
		},
		{
			name:     "Cease (invalid subcode)",
			input:    []byte{6, 1},
			wantFail: true,
		},
	}

	for _, test := range tests {
		res, err := decodeNotificationMsg(bytes.NewBuffer(test.input))

		if test.wantFail {
			if err != nil {
				continue
			}
			t.Errorf("Expected error did not happen for test %q", test.name)
			continue
		}

		if err != nil {
			t.Errorf("Unexpected failure for test %q: %v", test.name, err)
			continue
		}

		assert.Equal(t, test.expected, res)
	}
}

func TestDecodeUpdateMsg(t *testing.T) {
	tests := []struct {
		testNum        int
		input          []byte
		explicitLength uint16
		wantFail       bool
		expected       interface{}
	}{
		{
			// 2 withdraws only, valid update
			testNum:  1,
			input:    []byte{0, 5, 8, 10, 16, 192, 168, 0, 0},
			wantFail: false,
			expected: &BGPUpdate{
				WithdrawnRoutesLen: 5,
				WithdrawnRoutes: &NLRI{
					IP:     strAddr("10.0.0.0"),
					Pfxlen: 8,
					Next: &NLRI{
						IP:     strAddr("192.168.0.0"),
						Pfxlen: 16,
					},
				},
			},
		},
		{
			// 2 withdraws with one path attribute (ORIGIN), valid update
			testNum: 2,
			input: []byte{
				0, 5, // Withdrawn Routes Length
				8, 10, // 10.0.0.0/8
				16, 192, 168, // 192.168.0.0/16
				0, 5, // Total Path Attribute Length
				255,  // Attribute flags
				1,    // Attribute Type code
				0, 1, // Length
				2, // INCOMPLETE
			},
			wantFail: false,
			expected: &BGPUpdate{
				WithdrawnRoutesLen: 5,
				WithdrawnRoutes: &NLRI{
					IP:     strAddr("10.0.0.0"),
					Pfxlen: 8,
					Next: &NLRI{
						IP:     strAddr("192.168.0.0"),
						Pfxlen: 16,
					},
				},
				TotalPathAttrLen: 5,
				PathAttributes: &PathAttribute{
					Optional:       true,
					Transitive:     true,
					Partial:        true,
					ExtendedLength: true,
					Length:         1,
					TypeCode:       1,
					Value:          uint8(2),
				},
			},
		},
		{
			// 2 withdraws with two path attributes (ORIGIN + ASPath), valid update
			testNum: 3,
			input: []byte{0, 5, 8, 10, 16, 192, 168,
				0, 14, // Total Path Attribute Length

				255,  // Attribute flags
				1,    // Attribute Type code (ORIGIN)
				0, 1, // Length
				2, // INCOMPLETE

				0,      // Attribute flags
				2,      // Attribute Type code (AS Path)
				6,      // Length
				2,      // Type = AS_SEQUENCE
				2,      // Path Segement Length
				59, 65, // AS15169
				12, 248, // AS3320
			},
			wantFail: false,
			expected: &BGPUpdate{
				WithdrawnRoutesLen: 5,
				WithdrawnRoutes: &NLRI{
					IP:     strAddr("10.0.0.0"),
					Pfxlen: 8,
					Next: &NLRI{
						IP:     strAddr("192.168.0.0"),
						Pfxlen: 16,
					},
				},
				TotalPathAttrLen: 14,
				PathAttributes: &PathAttribute{
					Optional:       true,
					Transitive:     true,
					Partial:        true,
					ExtendedLength: true,
					Length:         1,
					TypeCode:       1,
					Value:          uint8(2),
					Next: &PathAttribute{
						Optional:       false,
						Transitive:     false,
						Partial:        false,
						ExtendedLength: false,
						Length:         6,
						TypeCode:       2,
						Value: ASPath{
							{
								Type:  2,
								Count: 2,
								ASNs: []uint32{
									15169,
									3320,
								},
							},
						},
					},
				},
			},
		},
		{
			// 2 withdraws with two path attributes (ORIGIN + ASPath), invalid AS Path segment type
			testNum: 4,
			input: []byte{0, 5, 8, 10, 16, 192, 168,
				0, 13, // Total Path Attribute Length

				255,  // Attribute flags
				1,    // Attribute Type code (ORIGIN)
				0, 1, // Length
				2, // INCOMPLETE

				0, // Attribute flags
				2, // Attribute Type code (AS Path)
				6, // Length
				1, // Type = AS_SET
				0, // Path Segement Length
			},
			wantFail: true,
		},
		{
			// 2 withdraws with two path attributes (ORIGIN + ASPath), invalid AS Path segment member count
			testNum: 5,
			input: []byte{0, 5, 8, 10, 16, 192, 168,
				0, 13, // Total Path Attribute Length

				255,  // Attribute flags
				1,    // Attribute Type code (ORIGIN)
				0, 1, // Length
				2, // INCOMPLETE

				0,      // Attribute flags
				2,      // Attribute Type code (AS Path)
				6,      // Length
				3,      // Type = INVALID
				2,      // Path Segement Length
				59, 65, // AS15169
				12, 248, // AS3320
			},
			wantFail: true,
		},
		{
			// 2 withdraws with two path attributes (ORIGIN + ASPath), valid update
			testNum: 6,
			input: []byte{0, 5, 8, 10, 16, 192, 168,
				0, 20, // Total Path Attribute Length

				255,  // Attribute flags
				1,    // Attribute Type code (ORIGIN)
				0, 1, // Length
				2, // INCOMPLETE

				0,      // Attribute flags
				2,      // Attribute Type code (AS Path)
				12,     // Length
				2,      // Type = AS_SEQUENCE
				2,      // Path Segement Length
				59, 65, // AS15169
				12, 248, // AS3320
				1,      // Type = AS_SET
				2,      // Path Segement Length
				59, 65, // AS15169
				12, 248, // AS3320
			},
			wantFail: false,
			expected: &BGPUpdate{
				WithdrawnRoutesLen: 5,
				WithdrawnRoutes: &NLRI{
					IP:     strAddr("10.0.0.0"),
					Pfxlen: 8,
					Next: &NLRI{
						IP:     strAddr("192.168.0.0"),
						Pfxlen: 16,
					},
				},
				TotalPathAttrLen: 20,
				PathAttributes: &PathAttribute{
					Optional:       true,
					Transitive:     true,
					Partial:        true,
					ExtendedLength: true,
					Length:         1,
					TypeCode:       1,
					Value:          uint8(2),
					Next: &PathAttribute{
						Optional:       false,
						Transitive:     false,
						Partial:        false,
						ExtendedLength: false,
						Length:         12,
						TypeCode:       2,
						Value: ASPath{
							{
								Type:  2,
								Count: 2,
								ASNs: []uint32{
									15169,
									3320,
								},
							},
							{
								Type:  1,
								Count: 2,
								ASNs: []uint32{
									15169,
									3320,
								},
							},
						},
					},
				},
			},
		},
		{
			// 2 withdraws with 3 path attributes (ORIGIN + ASPath, NH), valid update
			testNum: 7,
			input: []byte{0, 5, 8, 10, 16, 192, 168,
				0, 27, // Total Path Attribute Length

				255,  // Attribute flags
				1,    // Attribute Type code (ORIGIN)
				0, 1, // Length
				2, // INCOMPLETE

				0,      // Attribute flags
				2,      // Attribute Type code (AS Path)
				12,     // Length
				2,      // Type = AS_SEQUENCE
				2,      // Path Segement Length
				59, 65, // AS15169
				12, 248, // AS3320
				1,      // Type = AS_SET
				2,      // Path Segement Length
				59, 65, // AS15169
				12, 248, // AS3320

				0,              // Attribute flags
				3,              // Attribute Type code (Next Hop)
				4,              // Length
				10, 11, 12, 13, // Next Hop

			},
			wantFail: false,
			expected: &BGPUpdate{
				WithdrawnRoutesLen: 5,
				WithdrawnRoutes: &NLRI{
					IP:     strAddr("10.0.0.0"),
					Pfxlen: 8,
					Next: &NLRI{
						IP:     strAddr("192.168.0.0"),
						Pfxlen: 16,
					},
				},
				TotalPathAttrLen: 27,
				PathAttributes: &PathAttribute{

					Optional:       true,
					Transitive:     true,
					Partial:        true,
					ExtendedLength: true,
					Length:         1,
					TypeCode:       1,
					Value:          uint8(2),
					Next: &PathAttribute{

						Optional:       false,
						Transitive:     false,
						Partial:        false,
						ExtendedLength: false,
						Length:         12,
						TypeCode:       2,
						Value: ASPath{
							{
								Type:  2,
								Count: 2,
								ASNs: []uint32{
									15169,
									3320,
								},
							},
							{
								Type:  1,
								Count: 2,
								ASNs: []uint32{
									15169,
									3320,
								},
							},
						},
						Next: &PathAttribute{
							Optional:       false,
							Transitive:     false,
							Partial:        false,
							ExtendedLength: false,
							Length:         4,
							TypeCode:       3,
							Value:          strAddr("10.11.12.13"),
						},
					},
				},
			},
		},
		{
			// 2 withdraws with 4 path attributes (ORIGIN + ASPath, NH, MED), valid update
			testNum: 8,
			input: []byte{0, 5, 8, 10, 16, 192, 168,
				0, 34, // Total Path Attribute Length

				255,  // Attribute flags
				1,    // Attribute Type code (ORIGIN)
				0, 1, // Length
				2, // INCOMPLETE

				0,      // Attribute flags
				2,      // Attribute Type code (AS Path)
				12,     // Length
				2,      // Type = AS_SEQUENCE
				2,      // Path Segement Length
				59, 65, // AS15169
				12, 248, // AS3320
				1,      // Type = AS_SET
				2,      // Path Segement Length
				59, 65, // AS15169
				12, 248, // AS3320

				0,              // Attribute flags
				3,              // Attribute Type code (Next Hop)
				4,              // Length
				10, 11, 12, 13, // Next Hop

				0,          // Attribute flags
				4,          // Attribute Type code (Next Hop)
				4,          // Length
				0, 0, 1, 0, // MED 256

			},
			wantFail: false,
			expected: &BGPUpdate{
				WithdrawnRoutesLen: 5,
				WithdrawnRoutes: &NLRI{
					IP:     strAddr("10.0.0.0"),
					Pfxlen: 8,
					Next: &NLRI{
						IP:     strAddr("192.168.0.0"),
						Pfxlen: 16,
					},
				},
				TotalPathAttrLen: 34,
				PathAttributes: &PathAttribute{

					Optional:       true,
					Transitive:     true,
					Partial:        true,
					ExtendedLength: true,
					Length:         1,
					TypeCode:       1,
					Value:          uint8(2),
					Next: &PathAttribute{
						Optional:       false,
						Transitive:     false,
						Partial:        false,
						ExtendedLength: false,
						Length:         12,
						TypeCode:       2,
						Value: ASPath{
							{
								Type:  2,
								Count: 2,
								ASNs: []uint32{
									15169,
									3320,
								},
							},
							{
								Type:  1,
								Count: 2,
								ASNs: []uint32{
									15169,
									3320,
								},
							},
						},
						Next: &PathAttribute{
							Optional:       false,
							Transitive:     false,
							Partial:        false,
							ExtendedLength: false,
							Length:         4,
							TypeCode:       3,
							Value:          strAddr("10.11.12.13"),
							Next: &PathAttribute{
								Optional:       false,
								Transitive:     false,
								Partial:        false,
								ExtendedLength: false,
								Length:         4,
								TypeCode:       4,
								Value:          uint32(256),
							},
						},
					},
				},
			},
		},
		{
			// 2 withdraws with 4 path attributes (ORIGIN + ASPath, NH, MED, Local Pref), valid update
			testNum: 9,
			input: []byte{0, 5, 8, 10, 16, 192, 168,
				0, 41, // Total Path Attribute Length

				255,  // Attribute flags
				1,    // Attribute Type code (ORIGIN)
				0, 1, // Length
				2, // INCOMPLETE

				0,      // Attribute flags
				2,      // Attribute Type code (AS Path)
				12,     // Length
				2,      // Type = AS_SEQUENCE
				2,      // Path Segement Length
				59, 65, // AS15169
				12, 248, // AS3320
				1,      // Type = AS_SET
				2,      // Path Segement Length
				59, 65, // AS15169
				12, 248, // AS3320

				0,              // Attribute flags
				3,              // Attribute Type code (Next Hop)
				4,              // Length
				10, 11, 12, 13, // Next Hop

				0,          // Attribute flags
				4,          // Attribute Type code (MED)
				4,          // Length
				0, 0, 1, 0, // MED 256

				0,          // Attribute flags
				5,          // Attribute Type code (Local Pref)
				4,          // Length
				0, 0, 1, 0, // Local Pref 256
			},
			wantFail: false,
			expected: &BGPUpdate{
				WithdrawnRoutesLen: 5,
				WithdrawnRoutes: &NLRI{
					IP:     strAddr("10.0.0.0"),
					Pfxlen: 8,
					Next: &NLRI{
						IP:     strAddr("192.168.0.0"),
						Pfxlen: 16,
					},
				},
				TotalPathAttrLen: 41,
				PathAttributes: &PathAttribute{
					Optional:       true,
					Transitive:     true,
					Partial:        true,
					ExtendedLength: true,
					Length:         1,
					TypeCode:       1,
					Value:          uint8(2),
					Next: &PathAttribute{
						Optional:       false,
						Transitive:     false,
						Partial:        false,
						ExtendedLength: false,
						Length:         12,
						TypeCode:       2,
						Value: ASPath{
							{
								Type:  2,
								Count: 2,
								ASNs: []uint32{
									15169,
									3320,
								},
							},
							{
								Type:  1,
								Count: 2,
								ASNs: []uint32{
									15169,
									3320,
								},
							},
						},
						Next: &PathAttribute{
							Optional:       false,
							Transitive:     false,
							Partial:        false,
							ExtendedLength: false,
							Length:         4,
							TypeCode:       3,
							Value:          strAddr("10.11.12.13"),
							Next: &PathAttribute{
								Optional:       false,
								Transitive:     false,
								Partial:        false,
								ExtendedLength: false,
								Length:         4,
								TypeCode:       4,
								Value:          uint32(256),
								Next: &PathAttribute{
									Optional:       false,
									Transitive:     false,
									Partial:        false,
									ExtendedLength: false,
									Length:         4,
									TypeCode:       5,
									Value:          uint32(256),
								},
							},
						},
					},
				},
			},
		},
		{
			// 2 withdraws with 6 path attributes (ORIGIN, ASPath, NH, MED, Local Pref, Atomi Aggregate), valid update
			testNum: 9,
			input: []byte{0, 5, 8, 10, 16, 192, 168,
				0, 44, // Total Path Attribute Length

				255,  // Attribute flags
				1,    // Attribute Type code (ORIGIN)
				0, 1, // Length
				2, // INCOMPLETE

				0,      // Attribute flags
				2,      // Attribute Type code (AS Path)
				12,     // Length
				2,      // Type = AS_SEQUENCE
				2,      // Path Segement Length
				59, 65, // AS15169
				12, 248, // AS3320
				1,      // Type = AS_SET
				2,      // Path Segement Length
				59, 65, // AS15169
				12, 248, // AS3320

				0,              // Attribute flags
				3,              // Attribute Type code (Next Hop)
				4,              // Length
				10, 11, 12, 13, // Next Hop

				0,          // Attribute flags
				4,          // Attribute Type code (MED)
				4,          // Length
				0, 0, 1, 0, // MED 256

				0,          // Attribute flags
				5,          // Attribute Type code (Local Pref)
				4,          // Length
				0, 0, 1, 0, // Local Pref 256

				0, // Attribute flags
				6, // Attribute Type code (Atomic Aggregate)
				0, // Length
			},
			wantFail: false,
			expected: &BGPUpdate{
				WithdrawnRoutesLen: 5,
				WithdrawnRoutes: &NLRI{
					IP:     strAddr("10.0.0.0"),
					Pfxlen: 8,
					Next: &NLRI{
						IP:     strAddr("192.168.0.0"),
						Pfxlen: 16,
					},
				},
				TotalPathAttrLen: 44,
				PathAttributes: &PathAttribute{
					Optional:       true,
					Transitive:     true,
					Partial:        true,
					ExtendedLength: true,
					Length:         1,
					TypeCode:       1,
					Value:          uint8(2),
					Next: &PathAttribute{
						Optional:       false,
						Transitive:     false,
						Partial:        false,
						ExtendedLength: false,
						Length:         12,
						TypeCode:       2,
						Value: ASPath{
							{
								Type:  2,
								Count: 2,
								ASNs: []uint32{
									15169,
									3320,
								},
							},
							{
								Type:  1,
								Count: 2,
								ASNs: []uint32{
									15169,
									3320,
								},
							},
						},
						Next: &PathAttribute{
							Optional:       false,
							Transitive:     false,
							Partial:        false,
							ExtendedLength: false,
							Length:         4,
							TypeCode:       3,
							Value:          strAddr("10.11.12.13"),
							Next: &PathAttribute{
								Optional:       false,
								Transitive:     false,
								Partial:        false,
								ExtendedLength: false,
								Length:         4,
								TypeCode:       4,
								Value:          uint32(256),
								Next: &PathAttribute{
									Optional:       false,
									Transitive:     false,
									Partial:        false,
									ExtendedLength: false,
									Length:         4,
									TypeCode:       5,
									Value:          uint32(256),
									Next: &PathAttribute{
										Optional:       false,
										Transitive:     false,
										Partial:        false,
										ExtendedLength: false,
										Length:         0,
										TypeCode:       6,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			// 2 withdraws with 7 path attributes (ORIGIN, ASPath, NH, MED, Local Pref, Atomic Aggregate), valid update
			testNum: 10,
			input: []byte{0, 5, 8, 10, 16, 192, 168,
				0, 53, // Total Path Attribute Length

				255,  // Attribute flags
				1,    // Attribute Type code (ORIGIN)
				0, 1, // Length
				2, // INCOMPLETE

				0,      // Attribute flags
				2,      // Attribute Type code (AS Path)
				12,     // Length
				2,      // Type = AS_SEQUENCE
				2,      // Path Segement Length
				59, 65, // AS15169
				12, 248, // AS3320
				1,      // Type = AS_SET
				2,      // Path Segement Length
				59, 65, // AS15169
				12, 248, // AS3320

				0,              // Attribute flags
				3,              // Attribute Type code (Next Hop)
				4,              // Length
				10, 11, 12, 13, // Next Hop

				0,          // Attribute flags
				4,          // Attribute Type code (MED)
				4,          // Length
				0, 0, 1, 0, // MED 256

				0,          // Attribute flags
				5,          // Attribute Type code (Local Pref)
				4,          // Length
				0, 0, 1, 0, // Local Pref 256

				0, // Attribute flags
				6, // Attribute Type code (Atomic Aggregate)
				0, // Length

				0,    // Attribute flags
				7,    // Attribute Type code (Atomic Aggregate)
				6,    // Length
				1, 2, // ASN
				10, 11, 12, 13, // Address

				8, 11, // 11.0.0.0/8
			},
			wantFail: false,
			expected: &BGPUpdate{
				WithdrawnRoutesLen: 5,
				WithdrawnRoutes: &NLRI{
					IP:     strAddr("10.0.0.0"),
					Pfxlen: 8,
					Next: &NLRI{
						IP:     strAddr("192.168.0.0"),
						Pfxlen: 16,
					},
				},
				TotalPathAttrLen: 53,
				PathAttributes: &PathAttribute{
					Optional:       true,
					Transitive:     true,
					Partial:        true,
					ExtendedLength: true,
					Length:         1,
					TypeCode:       1,
					Value:          uint8(2),
					Next: &PathAttribute{
						Optional:       false,
						Transitive:     false,
						Partial:        false,
						ExtendedLength: false,
						Length:         12,
						TypeCode:       2,
						Value: ASPath{
							{
								Type:  2,
								Count: 2,
								ASNs: []uint32{
									15169,
									3320,
								},
							},
							{
								Type:  1,
								Count: 2,
								ASNs: []uint32{
									15169,
									3320,
								},
							},
						},
						Next: &PathAttribute{
							Optional:       false,
							Transitive:     false,
							Partial:        false,
							ExtendedLength: false,
							Length:         4,
							TypeCode:       3,
							Value:          strAddr("10.11.12.13"),
							Next: &PathAttribute{
								Optional:       false,
								Transitive:     false,
								Partial:        false,
								ExtendedLength: false,
								Length:         4,
								TypeCode:       4,
								Value:          uint32(256),
								Next: &PathAttribute{
									Optional:       false,
									Transitive:     false,
									Partial:        false,
									ExtendedLength: false,
									Length:         4,
									TypeCode:       5,
									Value:          uint32(256),
									Next: &PathAttribute{
										Optional:       false,
										Transitive:     false,
										Partial:        false,
										ExtendedLength: false,
										Length:         0,
										TypeCode:       6,
										Next: &PathAttribute{
											Optional:       false,
											Transitive:     false,
											Partial:        false,
											ExtendedLength: false,
											Length:         6,
											TypeCode:       7,
											Value: Aggretator{
												ASN:  uint16(258),
												Addr: strAddr("10.11.12.13"),
											},
										},
									},
								},
							},
						},
					},
				},
				NLRI: &NLRI{
					Pfxlen: 8,
					IP:     strAddr("11.0.0.0"),
				},
			},
		},
		{
			testNum: 11, // Incomplete Withdraw
			input: []byte{
				0, 5, // Length
			},
			wantFail: true,
		},
		{
			testNum:  12, // Empty buffer
			input:    []byte{},
			wantFail: true,
		},
		{
			testNum: 13,
			input: []byte{
				0, 0, // No Withdraws
				0, 5, // Total Path Attributes Length
			},
			wantFail: true,
		},
		{
			testNum: 14,
			input: []byte{
				0, 0, // No Withdraws
				0, 0, // Total Path Attributes Length
				24, // Incomplete NLRI
			},
			wantFail: true,
		},
		{
			testNum: 15, // Cut at Total Path Attributes Length
			input: []byte{
				0, 0, // No Withdraws
			},
			explicitLength: 5,
			wantFail:       true,
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		l := test.explicitLength
		if l == 0 {
			l = uint16(len(test.input))
		}
		msg, err := decodeUpdateMsg(buf, l)

		if err != nil && !test.wantFail {
			t.Errorf("Unexpected error in test %d: %v", test.testNum, err)
			continue
		}

		if err == nil && test.wantFail {
			t.Errorf("Expected error did not happen in test %d", test.testNum)
			continue
		}

		if err != nil && test.wantFail {
			continue
		}

		assert.Equalf(t, test.expected, msg, "%d", test.testNum)
	}
}

func TestDecodeMsgBody(t *testing.T) {
	tests := []struct {
		name     string
		buffer   *bytes.Buffer
		msgType  uint8
		length   uint16
		wantFail bool
		expected interface{}
	}{
		{
			name:     "Unknown msgType",
			msgType:  5,
			wantFail: true,
		},
	}

	for _, test := range tests {
		res, err := decodeMsgBody(test.buffer, test.msgType, test.length)
		if test.wantFail && err == nil {
			t.Errorf("Expected error dit not happen in test %q", test.name)
		}

		if !test.wantFail && err != nil {
			t.Errorf("Unexpected error in test %q: %v", test.name, err)
		}

		assert.Equal(t, test.expected, res)
	}
}

func TestDecodeOpenMsg(t *testing.T) {
	tests := []test{
		{
			// Valid message
			testNum:  1,
			input:    []byte{4, 1, 1, 0, 15, 10, 20, 30, 40, 0},
			wantFail: false,
			expected: &BGPOpen{
				Version:       4,
				AS:            257,
				HoldTime:      15,
				BGPIdentifier: 169090600,
				OptParmLen:    0,
				OptParams:     make([]OptParam, 0),
			},
		},
		{
			// Invalid Version
			testNum:  2,
			input:    []byte{3, 1, 1, 0, 15, 10, 10, 10, 11, 0},
			wantFail: true,
		},
	}

	genericTest(_decodeOpenMsg, tests, t)
}

func TestDecodeHeader(t *testing.T) {
	tests := []test{
		{
			// Valid header
			testNum:  1,
			input:    []byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 0, 19, KeepaliveMsg},
			wantFail: false,
			expected: &BGPHeader{
				Length: 19,
				Type:   KeepaliveMsg,
			},
		},
		{
			// Invalid length too short
			testNum:  2,
			input:    []byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 0, 18, KeepaliveMsg},
			wantFail: true,
			expected: &BGPHeader{
				Length: 18,
				Type:   KeepaliveMsg,
			},
		},
		{
			// Invalid length too long
			testNum:  3,
			input:    []byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 16, 1, KeepaliveMsg},
			wantFail: true,
			expected: &BGPHeader{
				Length: 18,
				Type:   KeepaliveMsg,
			},
		},
		{
			// Invalid message type 5
			testNum:  4,
			input:    []byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 0, 19, 5},
			wantFail: true,
			expected: &BGPHeader{
				Length: 19,
				Type:   KeepaliveMsg,
			},
		},
		{
			// Invalid message type 0
			testNum:  5,
			input:    []byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 0, 19, 0},
			wantFail: true,
			expected: &BGPHeader{
				Length: 19,
				Type:   KeepaliveMsg,
			},
		},
		{
			// Invalid marker
			testNum:  6,
			input:    []byte{1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 19, KeepaliveMsg},
			wantFail: true,
			expected: &BGPHeader{
				Length: 19,
				Type:   KeepaliveMsg,
			},
		},
		{
			// Incomplete Marker
			testNum:  7,
			input:    []byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
			wantFail: true,
		},
		{
			// Incomplete Header
			testNum:  8,
			input:    []byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 0, 19},
			wantFail: true,
		},
		{
			// Empty input
			testNum:  9,
			input:    []byte{},
			wantFail: true,
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		res, err := decodeHeader(buf)

		if err != nil {
			if test.wantFail {
				continue
			}
			t.Errorf("Unexpected failure for test %d: %v", test.testNum, err)
			continue
		}

		if test.wantFail {
			t.Errorf("Unexpected success fo test %d", test.testNum)
		}

		assert.Equal(t, test.expected, res)
	}
}

func genericTest(f decodeFunc, tests []test, t *testing.T) {
	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		msg, err := f(buf)

		if err != nil && !test.wantFail {
			t.Errorf("Unexpected error in test %d: %v", test.testNum, err)
			continue
		}

		if err == nil && test.wantFail {
			t.Errorf("Expected error did not happen in test %d", test.testNum)
			continue
		}

		if err != nil && test.wantFail {
			continue
		}

		if msg == nil {
			t.Errorf("Unexpected nil result in test %d. Expected: %v", test.testNum, test.expected)
			continue
		}

		assert.Equalf(t, test.expected, msg, "%d", test.testNum)
	}
}

func TestIsValidIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		input    uint32
		expected bool
	}{
		{
			name:     "Valid #1",
			input:    convert.Uint32b([]byte{8, 8, 8, 8}),
			expected: true,
		},
		{
			name:     "Multicast",
			input:    convert.Uint32b([]byte{239, 8, 8, 8}),
			expected: false,
		},
		{
			name:     "Loopback",
			input:    convert.Uint32b([]byte{127, 8, 8, 8}),
			expected: false,
		},
		{
			name:     "First byte 0",
			input:    convert.Uint32b([]byte{0, 8, 8, 8}),
			expected: false,
		},
		{
			name:     "All bytes 255",
			input:    convert.Uint32b([]byte{255, 255, 255, 255}),
			expected: false,
		},
	}

	for _, test := range tests {
		res := isValidIdentifier(test.input)
		assert.Equal(t, test.expected, res)
	}
}

func TestValidateOpenMessage(t *testing.T) {
	tests := []struct {
		name     string
		input    *BGPOpen
		wantFail bool
	}{
		{
			name: "Valid #1",
			input: &BGPOpen{
				Version:       4,
				BGPIdentifier: convert.Uint32b([]byte{8, 8, 8, 8}),
			},
			wantFail: false,
		},
		{
			name: "Invalid Identifier",
			input: &BGPOpen{
				Version:       4,
				BGPIdentifier: convert.Uint32b([]byte{0, 8, 8, 8}),
			},
			wantFail: true,
		},
	}

	for _, test := range tests {
		res := validateOpen(test.input)

		if res != nil {
			if test.wantFail {
				continue
			}
			t.Errorf("Unexpected failure for test %q: %v", test.name, res)
			continue
		}

		if test.wantFail {
			t.Errorf("Unexpected success for test %q", test.name)
		}
	}
}

func TestDecodeOptParams(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantFail bool
		expected []OptParam
	}{
		{
			name: "Add path capability",
			input: []byte{
				2,    // Type
				6,    // Length
				69,   // Code
				4,    // Length
				0, 1, // AFI
				1, // SAFI
				3, // Send/Receive
			},
			wantFail: false,
			expected: []OptParam{
				{
					Type:   2,
					Length: 6,
					Value: Capabilities{
						{
							Code:   69,
							Length: 4,
							Value: AddPathCapability{
								AFI:         1,
								SAFI:        1,
								SendReceive: 3,
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		res, err := decodeOptParams(buf, uint8(len(test.input)))
		if err != nil {
			if test.wantFail {
				continue
			}

			t.Errorf("Unexpected failure for test %q: %v", test.name, err)
			continue
		}

		if test.wantFail {
			t.Errorf("Unexpected success for test %q", test.name)
			continue
		}

		assert.Equal(t, test.expected, res)
	}
}

func TestDecodeCapability(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected Capability
		wantFail bool
	}{
		{
			name:  "Add Path",
			input: []byte{69, 4, 0, 1, 1, 3},
			expected: Capability{
				Code:   69,
				Length: 4,
				Value: AddPathCapability{
					AFI:         1,
					SAFI:        1,
					SendReceive: 3,
				},
			},
			wantFail: false,
		},
		{
			name:     "Fail",
			input:    []byte{69, 4, 0, 1},
			wantFail: true,
		},
	}

	for _, test := range tests {
		cap, err := decodeCapability(bytes.NewBuffer(test.input))
		if err != nil {
			if test.wantFail {
				continue
			}

			t.Errorf("Unexpected failure for test %q: %v", test.name, err)
			continue
		}

		if test.wantFail {
			t.Errorf("Unexpected success for test %q", err)
			continue
		}

		assert.Equal(t, test.expected, cap)
	}
}

func TestDecodeAddPathCapability(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected AddPathCapability
		wantFail bool
	}{
		{
			name:     "ok",
			input:    []byte{0, 1, 1, 3},
			wantFail: false,
			expected: AddPathCapability{
				AFI:         1,
				SAFI:        1,
				SendReceive: 3,
			},
		},
		{
			name:     "Incomplete",
			input:    []byte{0, 1, 1},
			wantFail: true,
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		cap, err := decodeAddPathCapability(buf)
		if err != nil {
			if test.wantFail {
				continue
			}

			t.Errorf("Unexpected failure for test %q: %v", test.name, err)
			continue
		}

		if test.wantFail {
			t.Errorf("Unexpected success for test %q", test.name)
			continue
		}

		assert.Equal(t, test.expected, cap)
	}
}

func strAddr(s string) uint32 {
	ret, _ := net.StrToAddr(s)
	return ret
}
