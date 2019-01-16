package server

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	btesting "github.com/bio-routing/bio-rd/testing"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type mockDev struct {
	wantFailProcessP2PHello       bool
	wantFailProcessIngressPacket  bool
	wantFailProcessLSPDU          bool
	wantFailProcessCSNP           bool
	wantFailProcessPSNP           bool
	callCountProcessP2PHello      int
	callCountProcessIngressPacket int
	callCountProcessLSPDU         int
	callCountProcessCSNP          int
	callCountProcessPSNP          int
}

func (md *mockDev) processP2PHello(*packet.P2PHello, types.MACAddress) error {
	md.callCountProcessP2PHello++
	if md.wantFailProcessP2PHello {
		return fmt.Errorf("processP2PHello failed")
	}

	return nil
}

func (md *mockDev) processIngressPacket([]byte, types.MACAddress) error {
	md.callCountProcessIngressPacket++
	if md.wantFailProcessIngressPacket {
		return fmt.Errorf("processIngressPacket failed")
	}

	return nil
}

func (md *mockDev) processLSPDU(l *packet.LSPDU, srv types.MACAddress) error {
	md.callCountProcessLSPDU++
	if md.wantFailProcessLSPDU {
		return fmt.Errorf("processLSPDU failed")
	}

	return nil
}

func (md *mockDev) processCSNP(l *packet.CSNP, srv types.MACAddress) error {
	md.callCountProcessCSNP++
	if md.wantFailProcessCSNP {
		return fmt.Errorf("ProcessCSNP failed")
	}

	return nil
}

func (md *mockDev) processPSNP(l *packet.PSNP, srv types.MACAddress) error {
	md.callCountProcessPSNP++
	if md.wantFailProcessPSNP {
		return fmt.Errorf("ProcessPSNP failed")
	}

	return nil
}

func TestReceiverRoutine(t *testing.T) {
	tests := []struct {
		name     string
		dev      *dev
		expected []byte
	}{
		{
			name: "All OK",
			dev: &dev{
				name: "eth0",
				srv: &Server{
					log: log.New(),
				},
				sys: &mockSys{
					wantFailRecvPacket:  false,
					recvPktCount:        1,
					stopRecvPktFail:     make(chan struct{}),
					stopRecvPktGraceful: make(chan struct{}),
				},
				self: &mockDev{},
			},
			expected: []byte("level=error msg=\"recvPacket() failed: Blocking\""),
		},
		{
			name: "Recv fail",
			dev: &dev{
				name: "eth0",
				srv: &Server{
					log: log.New(),
				},
				sys: &mockSys{
					wantFailRecvPacket:  true,
					stopRecvPktFail:     make(chan struct{}),
					stopRecvPktGraceful: make(chan struct{}),
				},
				self: &mockDev{},
			},
			expected: []byte("level=error msg=\"recvPacket() failed: Stopped\""),
		},
		{
			name: "process fail",
			dev: &dev{
				name: "eth0",
				srv: &Server{
					log: log.New(),
				},
				sys: &mockSys{
					wantFailRecvPacket:  false,
					recvPktCount:        1,
					stopRecvPktFail:     make(chan struct{}),
					stopRecvPktGraceful: make(chan struct{}),
				},
				self: &mockDev{},
			},
			expected: []byte("level=error msg=\"recvPacket() failed: Blocking\""),
		},
	}

	for _, test := range tests {
		test.dev.srv.log.Out = bytes.NewBuffer([]byte{})
		test.dev.srv.log.Formatter = btesting.NewLogFormatter()
		test.dev.wg.Add(1)
		go test.dev.receiverRoutine()
		close(test.dev.sys.(*mockSys).stopRecvPktGraceful)
		test.dev.wg.Wait()
		assert.Equal(t, test.expected, test.dev.srv.log.Out.(*bytes.Buffer).Bytes(), test.name)
	}
}

func TestProcessIngressPacket(t *testing.T) {
	tests := []struct {
		name        string
		dev         *dev
		mockDev     *mockDev
		pkt         []byte
		wantFail    bool
		expectedErr string
		expected    *mockDev
	}{
		{
			name:    "Decode Fail",
			dev:     newDev(nil, &config.ISISInterfaceConfig{Name: "eth0"}),
			mockDev: &mockDev{},
			pkt: []byte{
				// LLC
				0xfe, // DSAP
				0xfe, // SSAP
				0x03, // Control Fields

				// Header
				0x83,
				20,
				1,
				0,
				17, // PDU Type P2P Hello
				1,
			},
			wantFail:    true,
			expectedErr: "Unable to decode packet from [0 0 0 0 0 0] on eth0: [254 254 3 131 20 1 0 17 1]: Unable to decode header: Unable to decode fields: Unable to read from buffer: EOF",
			expected:    &mockDev{},
		},
		{
			name: "Invalid P2P Hello",
			dev:  newDev(nil, &config.ISISInterfaceConfig{Name: "eth0"}),
			mockDev: &mockDev{
				wantFailProcessP2PHello: true,
			},
			pkt: []byte{
				// LLC
				0xfe, // DSAP
				0xfe, // SSAP
				0x03, // Control Fields

				// Header
				0x83,
				20,
				1,
				0,
				17, // PDU Type P2P Hello
				1,
				0,
				0,

				// P2P Hello
				02,
				0, 0, 0, 0, 0, 2,
				0, 27,
				0, 50,
				1,

				//TLVs
				240, 5, 0x02, 0x00, 0x00, 0x01, 0x4b,
				129, 2, 0xcc, 0x8e,
				132, 4, 192, 168, 1, 0,
				1, 6, 0x05, 0x49, 0x00, 0x01, 0x00, 0x10,
				211, 3, 0, 0, 0,
			},
			wantFail: true,
			expected: &mockDev{
				wantFailProcessP2PHello:  true,
				callCountProcessP2PHello: 1,
			},
			expectedErr: "Unable to process P2P Hello: processP2PHello failed",
		},
		{
			name: "LSPDU fail",
			dev:  newDev(nil, &config.ISISInterfaceConfig{Name: "eth0"}),
			mockDev: &mockDev{
				wantFailProcessLSPDU: true,
			},
			pkt: []byte{
				// LLC
				0xfe, // DSAP
				0xfe, // SSAP
				0x03, // Control Fields

				// Header
				0x83,
				20,
				1,
				0,
				0x14, // PDU Type L2 LSPDU
				1,
				0,
				0,

				0, 30, // Length
				0, 200, // Lifetime
				10, 20, 30, 40, 50, 60, 0, 10, // LSPID
				0, 0, 1, 0, // Sequence Number
				0, 0, // Checksum
				3,                     // Typeblock
				137, 5, 1, 2, 3, 4, 5, // Hostname TLV
				12, 2, 0, 2, // Checksum TLV

				// P2P Hello
				02,
				0, 0, 0, 0, 0, 2,
				0, 27,
				0, 50,
				1,

				//TLVs
				240, 5, 0x02, 0x00, 0x00, 0x01, 0x4b,
				129, 2, 0xcc, 0x8e,
				132, 4, 192, 168, 1, 0,
				1, 6, 0x05, 0x49, 0x00, 0x01, 0x00, 0x10,
				211, 3, 0, 0, 0,
			},
			wantFail: true,
			expected: &mockDev{
				wantFailProcessLSPDU:  true,
				callCountProcessLSPDU: 1,
			},
			expectedErr: "Unable to process LSPDU: processLSPDU failed",
		},
		{
			name: "CSNP fail",
			dev:  newDev(nil, &config.ISISInterfaceConfig{Name: "eth0"}),
			mockDev: &mockDev{
				wantFailProcessCSNP: true,
			},
			pkt: []byte{
				// LLC
				0xfe, // DSAP
				0xfe, // SSAP
				0x03, // Control Fields

				// Header
				0x83,
				20,
				1,
				0,
				0x19, // PDU Type L2 CSNP
				1,
				0,
				0,

				0, 41, // PDU Length
				10, 20, 30, 40, 50, 60, 0, // Source ID
				11, 22, 33, 44, 55, 66, 0, 100, // StartLSPID
				11, 22, 33, 77, 88, 99, 0, 200, // EndLSPID
				9,    // TLV Type
				16,   // TLV Length
				1, 0, // Remaining Lifetime
				11, 22, 33, 44, 55, 66, // SystemID
				0, 20, // Pseudonode ID
				0, 0, 0, 20, // Sequence Number
				2, 0, // Checksum
			},
			wantFail: true,
			expected: &mockDev{
				wantFailProcessCSNP:  true,
				callCountProcessCSNP: 1,
			},
			expectedErr: "Unable to process CSNP: ProcessCSNP failed",
		},
		{
			name: "PSNP fail",
			dev:  newDev(nil, &config.ISISInterfaceConfig{Name: "eth0"}),
			mockDev: &mockDev{
				wantFailProcessPSNP: true,
			},
			pkt: []byte{
				// LLC
				0xfe, // DSAP
				0xfe, // SSAP
				0x03, // Control Fields

				// Header
				0x83,
				20,
				1,
				0,
				0x1b, // PDU Type L2 PSNP
				1,
				0,
				0,

				0, 33, // Length
				10, 20, 30, 40, 50, 60, 0, // Source ID

				1, 0, // Remaining Lifetime
				11, 22, 33, 44, 55, 66, // SystemID
				0,           // Pseudonode ID
				20,          // LSPNumber
				0, 0, 0, 20, // Sequence Number
				2, 0, // Checksum
			},
			wantFail: true,
			expected: &mockDev{
				wantFailProcessPSNP:  true,
				callCountProcessPSNP: 1,
			},
			expectedErr: "Unable to process PSNP: ProcessPSNP failed",
		},
		{
			name:    "PSNP Ok",
			dev:     newDev(nil, &config.ISISInterfaceConfig{Name: "eth0"}),
			mockDev: &mockDev{},
			pkt: []byte{
				// LLC
				0xfe, // DSAP
				0xfe, // SSAP
				0x03, // Control Fields

				// Header
				0x83,
				20,
				1,
				0,
				0x1b, // PDU Type L2 PSNP
				1,
				0,
				0,

				0, 33, // Length
				10, 20, 30, 40, 50, 60, 0, // Source ID

				1, 0, // Remaining Lifetime
				11, 22, 33, 44, 55, 66, // SystemID
				0,           // Pseudonode ID
				20,          // LSPNumber
				0, 0, 0, 20, // Sequence Number
				2, 0, // Checksum
			},
			wantFail: false,
			expected: &mockDev{
				callCountProcessPSNP: 1,
			},
		},
	}

	for _, test := range tests {
		test.dev.self = test.mockDev
		err := test.dev.processIngressPacket(test.pkt, types.MACAddress{})
		if err != nil && !test.wantFail {
			t.Errorf("Unexpected failure for test %q", test.name)
			continue
		}

		if test.wantFail && err == nil {
			t.Errorf("Unexpected success for test %q", test.name)
			continue
		}

		if err != nil {
			assert.Equal(t, test.expectedErr, err.Error(), test.name)
		}
		assert.Equal(t, test.expected, test.dev.self, test.name)
	}
}
