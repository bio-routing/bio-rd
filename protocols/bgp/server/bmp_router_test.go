package server

/*func TestBMPRouterServe(t *testing.T) {
	tests := []struct {
		name     string
		msg      []byte
		wantFail bool
	}{
		{
			name:     "Test #1",
			msg:      []byte{1, 2, 3},
			wantFail: true,
		},
	}

	for _, test := range tests {
		addr := net.IP{10, 20, 30, 40}
		port := uint16(123)
		rib4 := locRIB.New("inet.0")
		rib6 := locRIB.New("inet6.0")
		conA, conB := net.Pipe()

		r := newRouter(addr, port, rib4, rib6)
		buf := bytes.NewBuffer(nil)
		r.logger.Out = buf
		go r.serve(conA)

		conB.Write(test.msg)

		assert.Equalf(t, 0, len(buf.Bytes()), "Test %q", test.name)
	}
}

func TestStartStopBMP(t *testing.T) {
	addr := net.IP{10, 20, 30, 40}
	port := uint16(123)
	rib4 := locRIB.New("inet.0")
	rib6 := locRIB.New("inet6.0")

	con := biotesting.NewMockConn()

	r := newRouter(addr, port, rib4, rib6)
	go r.serve(con)

	r.stop <- struct{}{}

	r.runMu.Lock()
	assert.Equal(t, true, r.con.(*biotesting.MockConn).Closed)
}

func TestConfigureBySentOpen(t *testing.T) {
	tests := []struct {
		name     string
		p        *peer
		openMsg  *packet.BGPOpen
		expected *peer
	}{
		{
			name: "Test 32bit ASN",
			p:    &peer{},
			openMsg: &packet.BGPOpen{
				OptParams: []packet.OptParam{
					{
						Type: 2,
						Value: packet.Capabilities{
							{
								Code: packet.ASN4CapabilityCode,
								Value: packet.ASN4Capability{
									ASN4: 201701,
								},
							},
						},
					},
				},
			},
			expected: &peer{
				localASN: 201701,
			},
		},
		{
			name: "Test Add Path TX",
			p: &peer{
				ipv4: &peerAddressFamily{},
			},
			openMsg: &packet.BGPOpen{
				OptParams: []packet.OptParam{
					{
						Type: 2,
						Value: packet.Capabilities{
							{
								Code: packet.AddPathCapabilityCode,
								Value: packet.AddPathCapability{
									AFI:         1,
									SAFI:        1,
									SendReceive: 2,
								},
							},
						},
					},
				},
			},
			expected: &peer{
				ipv4: &peerAddressFamily{
					addPathSend: routingtable.ClientOptions{
						MaxPaths: 10,
					},
				},
			},
		},
		{
			name: "Test Add Path RX",
			p: &peer{
				ipv4: &peerAddressFamily{},
			},
			openMsg: &packet.BGPOpen{
				OptParams: []packet.OptParam{
					{
						Type: 2,
						Value: packet.Capabilities{
							{
								Code: packet.AddPathCapabilityCode,
								Value: packet.AddPathCapability{
									AFI:         1,
									SAFI:        1,
									SendReceive: 1,
								},
							},
						},
					},
				},
			},
			expected: &peer{
				ipv4: &peerAddressFamily{
					addPathReceive: true,
				},
			},
		},
		{
			name: "Test Add Path RX/TX",
			p: &peer{
				ipv4: &peerAddressFamily{},
			},
			openMsg: &packet.BGPOpen{
				OptParams: []packet.OptParam{
					{
						Type: 2,
						Value: packet.Capabilities{
							{
								Code: packet.AddPathCapabilityCode,
								Value: packet.AddPathCapability{
									AFI:         1,
									SAFI:        1,
									SendReceive: 3,
								},
							},
						},
					},
				},
			},
			expected: &peer{
				ipv4: &peerAddressFamily{
					addPathSend: routingtable.ClientOptions{
						MaxPaths: 10,
					},
					addPathReceive: true,
				},
			},
		},
	}

	for _, test := range tests {
		test.p.configureBySentOpen(test.openMsg)

		assert.Equalf(t, test.expected, test.p, "Test %q", test.name)
	}
}

func TestProcessPeerUpNotification(t *testing.T) {
	tests := []struct {
		name     string
		router   *router
		pkt      *bmppkt.PeerUpNotification
		wantFail bool
		expected *router
	}{
		{
			name: "Invalid sent open message",
			router: &router{
				neighbors: make(map[[16]byte]*neighbor),
			},
			pkt: &bmppkt.PeerUpNotification{
				PerPeerHeader: &bmppkt.PerPeerHeader{
					PeerType:          0,
					PeerFlags:         0,
					PeerDistinguisher: 0,
					PeerAddress:       [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 0, 255, 1},
					PeerAS:            51324,
					PeerBGPID:         100,
				},
				LocalAddress:    [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 0, 255, 0},
				LocalPort:       179,
				RemotePort:      34542,
				SentOpenMsg:     []byte{},
				ReceivedOpenMsg: []byte{},
				Information:     []byte{},
			},
			wantFail: true,
		},
		{
			name: "Invalid received open message",
			router: &router{
				neighbors: make(map[[16]byte]*neighbor),
			},
			pkt: &bmppkt.PeerUpNotification{
				PerPeerHeader: &bmppkt.PerPeerHeader{
					PeerType:          0,
					PeerFlags:         0,
					PeerDistinguisher: 0,
					PeerAddress:       [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 0, 255, 1},
					PeerAS:            51324,
					PeerBGPID:         100,
				},
				LocalAddress: [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 0, 255, 0},
				LocalPort:    179,
				RemotePort:   34542,
				SentOpenMsg: []byte{
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					0, 29,
					1,
					4,
					100, 200,
					20, 0,
					10, 20, 30, 40,
					0,
				},
				ReceivedOpenMsg: []byte{},
				Information:     []byte{},
			},
			wantFail: true,
		},
		{
			name: "Regular BGP by RFC4271",
			router: &router{
				rib4:      locRIB.New("inet.0"),
				rib6:      locRIB.New("inet6.0"),
				neighbors: make(map[[16]byte]*neighbor),
			},
			pkt: &bmppkt.PeerUpNotification{
				PerPeerHeader: &bmppkt.PerPeerHeader{
					PeerType:          0,
					PeerFlags:         0,
					PeerDistinguisher: 0,
					PeerAddress:       [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 0, 255, 1},
					PeerAS:            100,
					PeerBGPID:         100,
				},
				LocalAddress: [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 0, 255, 0},
				LocalPort:    179,
				RemotePort:   34542,
				SentOpenMsg: []byte{
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					0, 29,
					1,
					4,
					0, 200,
					20, 0,
					10, 20, 30, 40,
					0,
				},
				ReceivedOpenMsg: []byte{
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					0, 29,
					1,
					4,
					0, 100,
					20, 0,
					10, 20, 30, 50,
					0,
				},
				Information: []byte{},
			},
			wantFail: false,
			expected: &router{
				rib4: locRIB.New("inet.0"),
				rib6: locRIB.New("inet6.0"),
				neighbors: map[[16]byte]*neighbor{
					{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 0, 255, 1}: {
						localAS:     200,
						peerAS:      100,
						peerAddress: [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 0, 255, 1},
						routerID:    169090610,
						opt: &packet.DecodeOptions{
							AddPath:     false,
							Use32BitASN: false,
						},
						fsm: &FSM{
							isBMP:      true,
							neighborID: 169090610,
							state:      &establishedState{},
							peer: &peer{
								routerID:  169090600,
								addr:      bnet.IPv4FromOctets(10, 0, 255, 1),
								localAddr: bnet.IPv4FromOctets(10, 0, 255, 0),
								peerASN:   100,
								localASN:  200,
								ipv4:      &peerAddressFamily{},
								ipv6:      &peerAddressFamily{},
							},
							ipv4Unicast: &fsmAddressFamily{
								afi:          1,
								safi:         1,
								adjRIBIn:     adjRIBIn.New(filter.NewAcceptAllFilter(), &routingtable.ContributingASNs{}, 169090600, 0, false),
								importFilter: filter.NewAcceptAllFilter(),
								addPathTX:    routingtable.ClientOptions{BestOnly: true},
							},
							ipv6Unicast: &fsmAddressFamily{
								afi:          2,
								safi:         1,
								adjRIBIn:     adjRIBIn.New(filter.NewAcceptAllFilter(), &routingtable.ContributingASNs{}, 169090600, 0, false),
								importFilter: filter.NewAcceptAllFilter(),
								addPathTX:    routingtable.ClientOptions{BestOnly: true},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		err := test.router.processPeerUpNotification(test.pkt)
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

		test.expected.neighbors[test.pkt.PerPeerHeader.PeerAddress].fsm.state = &establishedState{fsm: test.expected.neighbors[test.pkt.PerPeerHeader.PeerAddress].fsm}

		if test.expected.neighbors[test.pkt.PerPeerHeader.PeerAddress].fsm.ipv4Unicast != nil {
			test.expected.neighbors[test.pkt.PerPeerHeader.PeerAddress].fsm.ipv4Unicast.rib = test.router.rib4
			test.expected.neighbors[test.pkt.PerPeerHeader.PeerAddress].fsm.ipv4Unicast.fsm = test.expected.neighbors[test.pkt.PerPeerHeader.PeerAddress].fsm
			test.expected.neighbors[test.pkt.PerPeerHeader.PeerAddress].fsm.ipv4Unicast.adjRIBIn.Register(test.router.rib4)
		}

		if test.expected.neighbors[test.pkt.PerPeerHeader.PeerAddress].fsm.ipv6Unicast != nil {
			test.expected.neighbors[test.pkt.PerPeerHeader.PeerAddress].fsm.ipv6Unicast.rib = test.router.rib6
			test.expected.neighbors[test.pkt.PerPeerHeader.PeerAddress].fsm.ipv6Unicast.fsm = test.expected.neighbors[test.pkt.PerPeerHeader.PeerAddress].fsm
			test.expected.neighbors[test.pkt.PerPeerHeader.PeerAddress].fsm.ipv6Unicast.adjRIBIn.Register(test.router.rib6)
		}

		assert.Equalf(t, test.expected, test.router, "Test %q", test.name)
	}

}

func TestProcessRouteMonitoringMsg(t *testing.T) {
	tests := []struct {
		name           string
		r              *router
		msg            *bmppkt.RouteMonitoringMsg
		expectedLogBuf string
		logOnly        bool
		expected       *router
	}{
		{
			name: "Unknown peer address",
			r: &router{
				address: net.IP{10, 20, 30, 40},
				logger:  log.New(),
			},
			msg: &bmppkt.RouteMonitoringMsg{
				PerPeerHeader: &bmppkt.PerPeerHeader{
					PeerAddress: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
				},
			},
			expectedLogBuf: "level=error msg=\"Received route monitoring message for non-existent neighbor [1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16] on 10.20.30.40\"",
			logOnly:        true,
			expected:       &router{},
		},
	}

	for _, test := range tests {
		logBuf := bytes.NewBuffer(nil)
		test.r.logger.Out = logBuf
		test.r.logger.Formatter = biotesting.NewLogFormatter()

		test.expected.logger = test.r.logger

		test.r.processRouteMonitoringMsg(test.msg)

		assert.Equalf(t, test.expectedLogBuf, string(logBuf.Bytes()), "Test %q", test.name)
		if test.logOnly {
			continue
		}

		assert.Equalf(t, test.expected, test.r, "Test %q", test.name)
	}
}

func TestProcessInitiationMsg(t *testing.T) {
	tests := []struct {
		name        string
		r           *router
		msg         *bmppkt.InitiationMessage
		expectedLog string
	}{
		{
			name: "Test #1",
			r: &router{
				address: net.IP{10, 20, 30, 40},
				logger:  log.New(),
			},
			msg: &bmppkt.InitiationMessage{
				TLVs: []*bmppkt.InformationTLV{
					{
						InformationType: 0,
						Information:     []byte("Foo Bar"),
					},
					{
						InformationType: 1,
						Information:     []byte("SYS DESCR"),
					},
					{
						InformationType: 2,
						Information:     []byte("core01.fra01"),
					},
				},
			},
			expectedLog: "level=info msg=\"Received initiation message from 10.20.30.40: Message: \\\"Foo Bar\\\" sysDescr.: SYS DESCR sysName.: core01.fra01\"",
		},
	}

	for _, test := range tests {
		logBuf := bytes.NewBuffer(nil)
		test.r.logger.Out = logBuf
		test.r.logger.Formatter = biotesting.NewLogFormatter()

		test.r.processInitiationMsg(test.msg)

		assert.Equalf(t, test.expectedLog, string(logBuf.Bytes()), "Test %q", test.name)
	}
}

func TestProcessTerminationMsg(t *testing.T) {
	tests := []struct {
		name        string
		r           *router
		msg         *bmppkt.TerminationMessage
		expectedLog string
		expected    *router
	}{
		{
			name: "Test shutdown",
			r: &router{
				con:     &biotesting.MockConn{},
				address: net.IP{10, 20, 30, 40},
				logger:  log.New(),
				neighbors: map[[16]byte]*neighbor{
					{1, 2, 3}: {},
				},
			},
			msg: &bmppkt.TerminationMessage{
				TLVs: []*bmppkt.InformationTLV{
					{
						InformationType: 0, // string type
						Information:     []byte("Foo Bar"),
					},
				},
			},
			expectedLog: "level=warning msg=\"Received termination message from 10.20.30.40: Message: \\\"Foo Bar\\\"\"",
			expected: &router{
				con: &biotesting.MockConn{
					Closed: true,
				},
				address:   net.IP{10, 20, 30, 40},
				neighbors: map[[16]byte]*neighbor{},
			},
		},
		{
			name: "Test logs",
			r: &router{
				con:     &biotesting.MockConn{},
				address: net.IP{10, 20, 30, 40},
				logger:  log.New(),
				neighbors: map[[16]byte]*neighbor{
					{1, 2, 3}: {},
				},
			},
			msg: &bmppkt.TerminationMessage{
				TLVs: []*bmppkt.InformationTLV{
					{
						InformationType: 1, // reason type
						Information:     []byte{0, 0},
					},
					{
						InformationType: 1, // reason type
						Information:     []byte{0, 1},
					},
					{
						InformationType: 1, // reason type
						Information:     []byte{0, 2},
					},
					{
						InformationType: 1, // reason type
						Information:     []byte{0, 3},
					},
					{
						InformationType: 1, // reason type
						Information:     []byte{0, 4},
					},
				},
			},
			expectedLog: "level=warning msg=\"Received termination message from 10.20.30.40: Session administratively downUnespcified reasonOut of resourcesRedundant connectionSession permanently administratively closed\"",
			expected: &router{
				con: &biotesting.MockConn{
					Closed: true,
				},
				address:   net.IP{10, 20, 30, 40},
				neighbors: map[[16]byte]*neighbor{},
			},
		},
	}

	for _, test := range tests {
		logBuf := bytes.NewBuffer(nil)
		test.r.logger.Out = logBuf
		test.r.logger.Formatter = biotesting.NewLogFormatter()

		test.expected.logger = test.r.logger

		test.r.processTerminationMsg(test.msg)

		assert.Equalf(t, test.expectedLog, string(logBuf.Bytes()), "Test %q", test.name)
		assert.Equalf(t, test.expected, test.r, "Test %q", test.name)
	}
}

func TestProcessPeerDownNotification(t *testing.T) {
	tests := []struct {
		name        string
		r           *router
		msg         *bmppkt.PeerDownNotification
		expectedLog string
		expected    *router
	}{
		{
			name: "Peer down notification for existing peer",
			r: &router{
				address: net.IP{10, 20, 30, 40},
				logger:  log.New(),
				neighbors: map[[16]byte]*neighbor{
					{1, 2, 3}: {},
				},
			},
			msg: &bmppkt.PeerDownNotification{
				PerPeerHeader: &bmppkt.PerPeerHeader{
					PeerAddress: [16]byte{1, 2, 3},
				},
			},
			expectedLog: "",
			expected: &router{
				address:   net.IP{10, 20, 30, 40},
				neighbors: map[[16]byte]*neighbor{},
			},
		},
		{
			name: "Peer down notification for non-existing peer",
			r: &router{
				address: net.IP{10, 20, 30, 40},
				logger:  log.New(),
				neighbors: map[[16]byte]*neighbor{
					{1, 2, 3}: {},
				},
			},
			msg: &bmppkt.PeerDownNotification{
				PerPeerHeader: &bmppkt.PerPeerHeader{
					PeerAddress: [16]byte{10, 20, 30},
				},
			},
			expectedLog: "level=warning msg=\"Received peer down notification for [10 20 30 0 0 0 0 0 0 0 0 0 0 0 0 0]: Peer doesn't exist.\"",
			expected: &router{
				address: net.IP{10, 20, 30, 40},
				neighbors: map[[16]byte]*neighbor{
					{1, 2, 3}: {},
				},
			},
		},
	}

	for _, test := range tests {
		logBuf := bytes.NewBuffer(nil)
		test.r.logger.Out = logBuf
		test.r.logger.Formatter = biotesting.NewLogFormatter()

		test.expected.logger = test.r.logger

		test.r.processPeerDownNotification(test.msg)

		assert.Equalf(t, test.expectedLog, string(logBuf.Bytes()), "Test %q", test.name)
		assert.Equalf(t, test.expected, test.r, "Test %q", test.name)
	}
}

func TestRegisterClients(t *testing.T) {
	n := &neighbor{
		fsm: &FSM{
			ipv4Unicast: &fsmAddressFamily{
				adjRIBIn: locRIB.New("inet.0"),
			},
			ipv6Unicast: &fsmAddressFamily{
				adjRIBIn: locRIB.New("inet6.0"),
			},
		},
	}

	client4 := locRIB.New("inet.0")
	client6 := locRIB.New("inet6.0")
	ac4 := afiClient{
		afi:    packet.IPv4AFI,
		client: client4,
	}
	ac6 := afiClient{
		afi:    packet.IPv6AFI,
		client: client6,
	}
	clients4 := map[afiClient]struct{}{
		ac4: {},
	}
	clients6 := map[afiClient]struct{}{
		ac6: {},
	}

	n.registerClients(clients4)
	n.registerClients(clients6)
	n.fsm.ipv4Unicast.adjRIBIn.AddPath(bnet.NewPfx(bnet.IPv4(0), 0), &route.Path{
		BGPPath: &route.BGPPath{
			LocalPref: 100,
		},
	})
	n.fsm.ipv6Unicast.adjRIBIn.AddPath(bnet.NewPfx(bnet.IPv6(0, 0), 0), &route.Path{
		BGPPath: &route.BGPPath{
			LocalPref: 200,
		},
	})

	assert.Equal(t, int64(1), n.fsm.ipv4Unicast.adjRIBIn.RouteCount())
	assert.Equal(t, int64(1), n.fsm.ipv6Unicast.adjRIBIn.RouteCount())
}

func TestIntegrationPeerUpRouteMonitor(t *testing.T) {
	addr := net.IP{10, 20, 30, 40}
	port := uint16(12346)

	rib4 := locRIB.New("inet.0")
	rib6 := locRIB.New("inet6.0")

	r := newRouter(addr, port, rib4, rib6)
	conA, conB := net.Pipe()

	go r.serve(conB)

	// Peer Up Notification
	_, err := conA.Write([]byte{
		// Common Header
		3,            // Version
		0, 0, 0, 126, // Message Length
		3, // Msg Type = Peer Up Notification

		// Per Peer Header
		0,                      // Peer Type
		0,                      // Peer Flags
		0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
		0, 0, 0, 100, // Peer AS
		0, 0, 0, 255, // Peer BGP ID
		0, 0, 0, 0, // Timestamp s
		0, 0, 0, 0, // Timestamp µs

		// Peer Up Notification
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 41, // 10.20.30.41 local address
		0, 123, // Local Port
		0, 234, // Remote Port

		// Sent OPEN message
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 29, // Length
		1,      // Open message type
		4,      // BGP Version
		0, 200, // AS
		0, 180, // Hold Time
		1, 0, 0, 1, // BGP Identifier
		0, // Opt param length

		// Received OPEN message
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 29, // Length
		1,      // Open message type
		4,      // BGP Version
		0, 100, // AS
		0, 180, // Hold Time
		1, 0, 0, 255, // BGP Identifier
		0, // Opt param length

		// SECOND MESSAGE:

		// Common Header
		3,            // Version
		0, 0, 0, 116, // Message Length
		0, // Msg Type = Route Monitoring Message

		// Per Peer Header
		0,                      // Peer Type
		0,                      // Peer Flags
		0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
		0, 0, 0, 100, // Peer AS
		0, 0, 0, 255, // Peer BGP ID
		0, 0, 0, 0, // Timestamp s
		0, 0, 0, 0, // Timestamp µs

		// BGP Update
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 68, // Length
		2, // Update

		0, 0, // Withdrawn Routes Length
		0, 41, // Total Path Attribute Length

		255,  // Attribute flags
		1,    // Attribute Type code (ORIGIN)
		0, 1, // Length
		2, // INCOMPLETE

		0,      // Attribute flags
		2,      // Attribute Type code (AS Path)
		12,     // Length
		2,      // Type = AS_SEQUENCE
		2,      // Path Segment Length
		59, 65, // AS15169
		12, 248, // AS3320
		1,      // Type = AS_SET
		2,      // Path Segment Length
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

		// NLRI
		24,
		192, 168, 0,
	})

	if err != nil {
		panic("write failed")
	}

	time.Sleep(time.Millisecond * 50)
	assert.NotEmpty(t, r.neighbors)

	time.Sleep(time.Millisecond * 50)

	count := rib4.RouteCount()
	if count != 1 {
		t.Errorf("Unexpected route count. Expected: 1 Got: %d", count)
	}

	conA.Close()
}

func TestIntegrationPeerUpRouteMonitorIPv6IPv4(t *testing.T) {
	addr := net.IP{10, 20, 30, 40}
	port := uint16(12346)

	rib4 := locRIB.New("inet.0")
	rib6 := locRIB.New("inet6.0")

	r := newRouter(addr, port, rib4, rib6)
	conA, conB := net.Pipe()

	go r.serve(conB)

	// Peer Up Notification
	_, err := conA.Write([]byte{
		// Common Header
		3,            // Version
		0, 0, 0, 142, // Message Length
		3, // Msg Type = Peer Up Notification

		// Per Peer Header
		0,                      // Peer Type
		0,                      // Peer Flags
		0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
		0, 0, 0, 100, // Peer AS
		0, 0, 0, 255, // Peer BGP ID
		0, 0, 0, 0, // Timestamp s
		0, 0, 0, 0, // Timestamp µs

		// Peer Up Notification
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 41, // 10.20.30.41 local address
		0, 123, // Local Port
		0, 234, // Remote Port

		// Sent OPEN message
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 37, // Length
		1,      // Open message type
		4,      // BGP Version
		0, 200, // AS
		0, 180, // Hold Time
		1, 0, 0, 1, // BGP Identifier
		8, // Opt param length
		2, 6,
		1, 4, // MP BGP
		0, 2, // IPv6
		0, 1, // Unicast

		// Received OPEN message
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 37, // Length
		1,      // Open message type
		4,      // BGP Version
		0, 100, // AS
		0, 180, // Hold Time
		1, 0, 0, 255, // BGP Identifier
		8, // Opt param length
		2, 6,
		1, 4, // MP BGP
		0, 2, // IPv6
		0, 1, // Unicast

		// SECOND MESSAGE:

		// Common Header
		3,            // Version
		0, 0, 0, 138, // Message Length
		0, // Msg Type = Route Monitoring Message

		// Per Peer Header
		0,                      // Peer Type
		0,                      // Peer Flags
		0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
		0, 0, 0, 100, // Peer AS
		0, 0, 0, 255, // Peer BGP ID
		0, 0, 0, 0, // Timestamp s
		0, 0, 0, 0, // Timestamp µs

		// BGP Update
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 90, // Length
		2, // Update

		0, 0, // Withdrawn Routes Length
		0, 67, // Total Path Attribute Length

		255,
		14,    // MP REACH NLRI
		0, 22, // Length
		0, 2, // IPv6
		1,  // Unicast
		16, // IPv6 Next Hop
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0, // Reserved
		0, // Pfxlen /0

		255,  // Attribute flags
		1,    // Attribute Type code (ORIGIN)
		0, 1, // Length
		2, // INCOMPLETE

		0,      // Attribute flags
		2,      // Attribute Type code (AS Path)
		12,     // Length
		2,      // Type = AS_SEQUENCE
		2,      // Path Segment Length
		59, 65, // AS15169
		12, 248, // AS3320
		1,      // Type = AS_SET
		2,      // Path Segment Length
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

		// THIRD MESSAGE
		// Common Header
		3,            // Version
		0, 0, 0, 113, // Message Length
		0, // Msg Type = Route Monitoring Message

		// Per Peer Header
		0,                      // Peer Type
		0,                      // Peer Flags
		0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
		0, 0, 0, 100, // Peer AS
		0, 0, 0, 255, // Peer BGP ID
		0, 0, 0, 0, // Timestamp s
		0, 0, 0, 0, // Timestamp µs

		// BGP Update
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 65, // Length
		2, // Update

		0, 0, // Withdrawn Routes Length
		0, 41, // Total Path Attribute Length

		255,  // Attribute flags
		1,    // Attribute Type code (ORIGIN)
		0, 1, // Length
		2, // INCOMPLETE

		0,      // Attribute flags
		2,      // Attribute Type code (AS Path)
		12,     // Length
		2,      // Type = AS_SEQUENCE
		2,      // Path Segment Length
		59, 65, // AS15169
		12, 248, // AS3320
		1,      // Type = AS_SET
		2,      // Path Segment Length
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

		0, // /0
	})

	if err != nil {
		panic("write failed")
	}

	time.Sleep(time.Millisecond * 50)
	assert.NotEmpty(t, r.neighbors)

	time.Sleep(time.Millisecond * 50)

	count := rib6.RouteCount()
	if count != 1 {
		t.Errorf("Unexpected IPv6 route count. Expected: 1 Got: %d", count)
	}

	count = rib4.RouteCount()
	if count != 1 {
		t.Errorf("Unexpected IPv4 route count. Expected: 1 Got: %d", count)
	}

	conA.Close()
}

func TestIntegrationPeerUpRouteMonitorIPv4IPv6(t *testing.T) {
	addr := net.IP{10, 20, 30, 40}
	port := uint16(12346)

	rib4 := locRIB.New("inet.0")
	rib6 := locRIB.New("inet6.0")

	r := newRouter(addr, port, rib4, rib6)
	conA, conB := net.Pipe()

	go r.serve(conB)

	// Peer Up Notification
	_, err := conA.Write([]byte{
		// Common Header
		3,            // Version
		0, 0, 0, 142, // Message Length
		3, // Msg Type = Peer Up Notification

		// Per Peer Header
		0,                      // Peer Type
		0,                      // Peer Flags
		0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
		0, 0, 0, 100, // Peer AS
		0, 0, 0, 255, // Peer BGP ID
		0, 0, 0, 0, // Timestamp s
		0, 0, 0, 0, // Timestamp µs

		// Peer Up Notification
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 41, // 10.20.30.41 local address
		0, 123, // Local Port
		0, 234, // Remote Port

		// Sent OPEN message
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 37, // Length
		1,      // Open message type
		4,      // BGP Version
		0, 200, // AS
		0, 180, // Hold Time
		1, 0, 0, 1, // BGP Identifier
		8, // Opt param length
		2, 6,
		1, 4, // MP BGP
		0, 2, // IPv6
		0, 1, // Unicast

		// Received OPEN message
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 37, // Length
		1,      // Open message type
		4,      // BGP Version
		0, 100, // AS
		0, 180, // Hold Time
		1, 0, 0, 255, // BGP Identifier
		8, // Opt param length
		2, 6,
		1, 4, // MP BGP
		0, 2, // IPv6
		0, 1, // Unicast

		// SECOND MESSAGE:

		// Common Header
		3,            // Version
		0, 0, 0, 113, // Message Length
		0, // Msg Type = Route Monitoring Message

		// Per Peer Header
		0,                      // Peer Type
		0,                      // Peer Flags
		0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
		0, 0, 0, 100, // Peer AS
		0, 0, 0, 255, // Peer BGP ID
		0, 0, 0, 0, // Timestamp s
		0, 0, 0, 0, // Timestamp µs

		// BGP Update
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 65, // Length
		2, // Update

		0, 0, // Withdrawn Routes Length
		0, 41, // Total Path Attribute Length

		255,  // Attribute flags
		1,    // Attribute Type code (ORIGIN)
		0, 1, // Length
		2, // INCOMPLETE

		0,      // Attribute flags
		2,      // Attribute Type code (AS Path)
		12,     // Length
		2,      // Type = AS_SEQUENCE
		2,      // Path Segment Length
		59, 65, // AS15169
		12, 248, // AS3320
		1,      // Type = AS_SET
		2,      // Path Segment Length
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

		0, // /0

		// THIRD MESSAGE

		// Common Header
		3,            // Version
		0, 0, 0, 138, // Message Length
		0, // Msg Type = Route Monitoring Message

		// Per Peer Header
		0,                      // Peer Type
		0,                      // Peer Flags
		0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
		0, 0, 0, 100, // Peer AS
		0, 0, 0, 255, // Peer BGP ID
		0, 0, 0, 0, // Timestamp s
		0, 0, 0, 0, // Timestamp µs

		// BGP Update
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 90, // Length
		2, // Update

		0, 0, // Withdrawn Routes Length
		0, 67, // Total Path Attribute Length

		255,
		14,    // MP REACH NLRI
		0, 22, // Length
		0, 2, // IPv6
		1,  // Unicast
		16, // IPv6 Next Hop
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0, // Reserved
		0, // Pfxlen /0

		255,  // Attribute flags
		1,    // Attribute Type code (ORIGIN)
		0, 1, // Length
		2, // INCOMPLETE

		0,      // Attribute flags
		2,      // Attribute Type code (AS Path)
		12,     // Length
		2,      // Type = AS_SEQUENCE
		2,      // Path Segment Length
		59, 65, // AS15169
		12, 248, // AS3320
		1,      // Type = AS_SET
		2,      // Path Segment Length
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
	})

	if err != nil {
		panic("write failed")
	}

	time.Sleep(time.Millisecond * 50)
	assert.NotEmpty(t, r.neighbors)

	time.Sleep(time.Millisecond * 50)

	count := rib6.RouteCount()
	if count != 1 {
		t.Errorf("Unexpected IPv6 route count. Expected: 1 Got: %d", count)
	}

	count = rib4.RouteCount()
	if count != 1 {
		t.Errorf("Unexpected IPv4 route count. Expected: 1 Got: %d", count)
	}

	conA.Close()
}

func TestIntegrationPeerUpRouteMonitorIPv6(t *testing.T) {
	addr := net.IP{10, 20, 30, 40}
	port := uint16(12346)

	rib4 := locRIB.New("inet.0")
	rib6 := locRIB.New("inet6.0")

	r := newRouter(addr, port, rib4, rib6)
	conA, conB := net.Pipe()

	go r.serve(conB)

	// Peer Up Notification
	_, err := conA.Write([]byte{
		// Common Header
		3,            // Version
		0, 0, 0, 142, // Message Length
		3, // Msg Type = Peer Up Notification

		// Per Peer Header
		0,                      // Peer Type
		0,                      // Peer Flags
		0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
		0, 0, 0, 100, // Peer AS
		0, 0, 0, 255, // Peer BGP ID
		0, 0, 0, 0, // Timestamp s
		0, 0, 0, 0, // Timestamp µs

		// Peer Up Notification
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 41, // 10.20.30.41 local address
		0, 123, // Local Port
		0, 234, // Remote Port

		// Sent OPEN message
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 37, // Length
		1,      // Open message type
		4,      // BGP Version
		0, 200, // AS
		0, 180, // Hold Time
		1, 0, 0, 1, // BGP Identifier
		8, // Opt param length
		2, 6,
		1, 4, // MP BGP
		0, 2, // IPv6
		0, 1, // Unicast

		// Received OPEN message
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 37, // Length
		1,      // Open message type
		4,      // BGP Version
		0, 100, // AS
		0, 180, // Hold Time
		1, 0, 0, 255, // BGP Identifier
		8, // Opt param length
		2, 6,
		1, 4, // MP BGP
		0, 2, // IPv6
		0, 1, // Unicast

		// SECOND MESSAGE:

		// Common Header
		3,            // Version
		0, 0, 0, 138, // Message Length
		0, // Msg Type = Route Monitoring Message

		// Per Peer Header
		0,                      // Peer Type
		0,                      // Peer Flags
		0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
		0, 0, 0, 100, // Peer AS
		0, 0, 0, 255, // Peer BGP ID
		0, 0, 0, 0, // Timestamp s
		0, 0, 0, 0, // Timestamp µs

		// BGP Update
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 90, // Length
		2, // Update

		0, 0, // Withdrawn Routes Length
		0, 67, // Total Path Attribute Length

		255,
		14,    // MP REACH NLRI
		0, 22, // Length
		0, 2, // IPv6
		1,  // Unicast
		16, // IPv6 Next Hop
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0, // Reserved
		0, // Pfxlen /0

		255,  // Attribute flags
		1,    // Attribute Type code (ORIGIN)
		0, 1, // Length
		2, // INCOMPLETE

		0,      // Attribute flags
		2,      // Attribute Type code (AS Path)
		12,     // Length
		2,      // Type = AS_SEQUENCE
		2,      // Path Segment Length
		59, 65, // AS15169
		12, 248, // AS3320
		1,      // Type = AS_SET
		2,      // Path Segment Length
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
	})

	if err != nil {
		panic("write failed")
	}

	time.Sleep(time.Millisecond * 50)
	assert.NotEmpty(t, r.neighbors)

	time.Sleep(time.Millisecond * 50)

	count := rib6.RouteCount()
	if count != 1 {
		t.Errorf("Unexpected IPv6 route count. Expected: 1 Got: %d", count)
	}

	count = rib4.RouteCount()
	if count != 0 {
		t.Errorf("Unexpected IPv4 route count. Expected: 0 Got: %d", count)
	}

	conA.Close()
}

func TestIntegrationIncompleteBMPMsg(t *testing.T) {
	addr := net.IP{10, 20, 30, 40}
	port := uint16(12346)

	rib4 := locRIB.New("inet.0")
	rib6 := locRIB.New("inet6.0")

	r := newRouter(addr, port, rib4, rib6)
	con := biotesting.NewMockConn()

	// Peer Up Notification with invalid version number (4)
	con.Write([]byte{
		// Common Header
		4,            // Version
		0, 0, 0, 126, // Message Length
		3, // Msg Type = Peer Up Notification

		// Per Peer Header
		0,                      // Peer Type
		0,                      // Peer Flags
		0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
		0, 0, 0, 100, // Peer AS
		0, 0, 0, 255, // Peer BGP ID
		0, 0, 0, 0, // Timestamp s
		0, 0, 0, 0, // Timestamp µs

		// Peer Up Notification
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 41, // 10.20.30.41 local address
		0, 123, // Local Port
		0, 234, // Remote Port

		// Sent OPEN message
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 29, // Length
		1,      // Open message type
		4,      // BGP Version
		0, 200, // AS
		0, 180, // Hold Time
		1, 0, 0, 1, // BGP Identifier
		0, // Opt param length

		// Received OPEN message
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 29, // Length
		1,      // Open message type
		4,      // BGP Version
		0, 100, // AS
		0, 180, // Hold Time
		1, 0, 0, 255, // BGP Identifier
		0, // Opt param length

		// SECOND MESSAGE:

		// Common Header
		3,            // Version
		0, 0, 0, 116, // Message Length
		0, // Msg Type = Route Monitoring Message

		// Per Peer Header
		0,                      // Peer Type
		0,                      // Peer Flags
		0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
		0, 0, 0, 100, // Peer AS
		0, 0, 0, 255, // Peer BGP ID
		0, 0, 0, 0, // Timestamp s
		0, 0, 0, 0, // Timestamp µs

	})

	r.serve(con)

	if !con.Closed {
		t.Errorf("Connection not closed although failure should have happened")
	}
}

func TestBMPFullRunWithWithdraw(t *testing.T) {
	addr := net.IP{10, 20, 30, 40}
	port := uint16(12346)

	rib4 := locRIB.New("inet.0")
	rib6 := locRIB.New("inet6.0")

	r := newRouter(addr, port, rib4, rib6)
	con := biotesting.NewMockConn()

	go r.serve(con)

	con.Write([]byte{
		// #####################################################################
		// Initiation Message
		// #####################################################################
		// Common Header
		3,
		0, 0, 0, 23,
		4,
		0, 1, // sysDescr
		0, 4, // Length
		42, 42, 42, 42, // AAAA
		0, 2, //sysName
		0, 5, // Length
		43, 43, 43, 43, 43, // BBBBB

		// #####################################################################
		// Peer UP Notification
		// #####################################################################
		// Common Header
		3,            // Version
		0, 0, 0, 126, // Message Length
		3, // Msg Type = Peer Up Notification
		// Per Peer Header
		0,                      // Peer Type
		0,                      // Peer Flags
		0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
		0, 0, 0, 100, // Peer AS
		0, 0, 0, 255, // Peer BGP ID
		0, 0, 0, 0, // Timestamp s
		0, 0, 0, 0, // Timestamp µs
		// Peer Up Notification
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 41, // 10.20.30.41 local address
		0, 123, // Local Port
		0, 234, // Remote Port
		// Sent OPEN message
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 29, // Length
		1,      // Open message type
		4,      // BGP Version
		0, 200, // AS
		0, 180, // Hold Time
		1, 0, 0, 1, // BGP Identifier
		0, // Opt param length
		// Received OPEN message
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 29, // Length
		1,      // Open message type
		4,      // BGP Version
		0, 100, // AS
		0, 180, // Hold Time
		1, 0, 0, 255, // BGP Identifier
		0, // Opt param length

		// #####################################################################
		// Route Monitoring Message #1
		// #####################################################################
		// Common Header
		3,            // Version
		0, 0, 0, 116, // Message Length
		0, // Msg Type = Route Monitoring Message

		// Per Peer Header
		0,                      // Peer Type
		0,                      // Peer Flags
		0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
		0, 0, 0, 100, // Peer AS
		0, 0, 0, 255, // Peer BGP ID
		0, 0, 0, 0, // Timestamp s
		0, 0, 0, 0, // Timestamp µs

		// BGP Update
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 68, // Length
		2, // Update

		0, 0, // Withdrawn Routes Length
		0, 41, // Total Path Attribute Length

		255,  // Attribute flags
		1,    // Attribute Type code (ORIGIN)
		0, 1, // Length
		2, // INCOMPLETE

		0,      // Attribute flags
		2,      // Attribute Type code (AS Path)
		12,     // Length
		2,      // Type = AS_SEQUENCE
		2,      // Path Segment Length
		59, 65, // AS15169
		12, 248, // AS3320
		1,      // Type = AS_SET
		2,      // Path Segment Length
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

		// NLRI
		24,
		192, 168, 0,

		// #####################################################################
		// Route Monitoring Message #2 (withdraw of prefix from #1)
		// #####################################################################
		// Common Header
		3,           // Version
		0, 0, 0, 75, // Message Length
		0, // Msg Type = Route Monitoring Message

		// Per Peer Header
		0,                      // Peer Type
		0,                      // Peer Flags
		0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
		0, 0, 0, 100, // Peer AS
		0, 0, 0, 255, // Peer BGP ID
		0, 0, 0, 0, // Timestamp s
		0, 0, 0, 0, // Timestamp µs

		// BGP Update
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 27, // Length
		2, // Update

		0, 4, // Withdrawn Routes Length
		24,
		192, 168, 0,
		0, 0, // Total Path Attribute Length
	})

	time.Sleep(time.Millisecond * 50)
	assert.NotEmpty(t, r.neighbors)

	time.Sleep(time.Millisecond * 50)

	count := rib4.RouteCount()
	if count != 0 {
		t.Errorf("Unexpected route count. Expected: 0 Got: %d", count)
	}

}

func TestBMPFullRunWithPeerDownNotification(t *testing.T) {
	addr := net.IP{10, 20, 30, 40}
	port := uint16(12346)

	rib4 := locRIB.New("inet.0")
	rib6 := locRIB.New("inet6.0")

	r := newRouter(addr, port, rib4, rib6)
	con := biotesting.NewMockConn()

	go r.serve(con)

	con.Write([]byte{
		// #####################################################################
		// Initiation Message
		// #####################################################################
		// Common Header
		3,
		0, 0, 0, 23,
		4,
		0, 1, // sysDescr
		0, 4, // Length
		42, 42, 42, 42, // AAAA
		0, 2, //sysName
		0, 5, // Length
		43, 43, 43, 43, 43, // BBBBB

		// #####################################################################
		// Peer UP Notification
		// #####################################################################
		// Common Header
		3,            // Version
		0, 0, 0, 126, // Message Length
		3, // Msg Type = Peer Up Notification
		// Per Peer Header
		0,                      // Peer Type
		0,                      // Peer Flags
		0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
		0, 0, 0, 100, // Peer AS
		0, 0, 0, 255, // Peer BGP ID
		0, 0, 0, 0, // Timestamp s
		0, 0, 0, 0, // Timestamp µs
		// Peer Up Notification
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 41, // 10.20.30.41 local address
		0, 123, // Local Port
		0, 234, // Remote Port
		// Sent OPEN message
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 29, // Length
		1,      // Open message type
		4,      // BGP Version
		0, 200, // AS
		0, 180, // Hold Time
		1, 0, 0, 1, // BGP Identifier
		0, // Opt param length
		// Received OPEN message
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 29, // Length
		1,      // Open message type
		4,      // BGP Version
		0, 100, // AS
		0, 180, // Hold Time
		1, 0, 0, 255, // BGP Identifier
		0, // Opt param length

		// #####################################################################
		// Route Monitoring Message #1
		// #####################################################################
		// Common Header
		3,            // Version
		0, 0, 0, 116, // Message Length
		0, // Msg Type = Route Monitoring Message

		// Per Peer Header
		0,                      // Peer Type
		0,                      // Peer Flags
		0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
		0, 0, 0, 100, // Peer AS
		0, 0, 0, 255, // Peer BGP ID
		0, 0, 0, 0, // Timestamp s
		0, 0, 0, 0, // Timestamp µs

		// BGP Update
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 68, // Length
		2, // Update

		0, 0, // Withdrawn Routes Length
		0, 41, // Total Path Attribute Length

		255,  // Attribute flags
		1,    // Attribute Type code (ORIGIN)
		0, 1, // Length
		2, // INCOMPLETE

		0,      // Attribute flags
		2,      // Attribute Type code (AS Path)
		12,     // Length
		2,      // Type = AS_SEQUENCE
		2,      // Path Segment Length
		59, 65, // AS15169
		12, 248, // AS3320
		1,      // Type = AS_SET
		2,      // Path Segment Length
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

		// NLRI
		24,
		192, 168, 0,

		// #####################################################################
		// Peer Down Notification
		// #####################################################################
		// Common Header
		3,           // Version
		0, 0, 0, 49, // Message Length
		2, // Msg Type = Peer Down Notification

		// Per Peer Header
		0,                      // Peer Type
		0,                      // Peer Flags
		0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
		0, 0, 0, 100, // Peer AS
		0, 0, 0, 255, // Peer BGP ID
		0, 0, 0, 0, // Timestamp s
		0, 0, 0, 0, // Timestamp µs

		4, // Reason
	})

	time.Sleep(time.Millisecond * 50)
	assert.Empty(t, r.neighbors)

	time.Sleep(time.Millisecond * 50)

	count := rib4.RouteCount()
	if count != 0 {
		t.Errorf("Unexpected route count. Expected: 0 Got: %d", count)
	}
}

func TestBMPFullRunWithTerminationMessage(t *testing.T) {
	addr := net.IP{10, 20, 30, 40}
	port := uint16(12346)

	rib4 := locRIB.New("inet.0")
	rib6 := locRIB.New("inet6.0")

	r := newRouter(addr, port, rib4, rib6)
	con := biotesting.NewMockConn()

	go r.serve(con)

	con.Write([]byte{
		// #####################################################################
		// Initiation Message
		// #####################################################################
		// Common Header
		3,
		0, 0, 0, 23,
		4,
		0, 1, // sysDescr
		0, 4, // Length
		42, 42, 42, 42, // AAAA
		0, 2, //sysName
		0, 5, // Length
		43, 43, 43, 43, 43, // BBBBB

		// #####################################################################
		// Peer UP Notification
		// #####################################################################
		// Common Header
		3,            // Version
		0, 0, 0, 126, // Message Length
		3, // Msg Type = Peer Up Notification
		// Per Peer Header
		0,                      // Peer Type
		0,                      // Peer Flags
		0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
		0, 0, 0, 100, // Peer AS
		0, 0, 0, 255, // Peer BGP ID
		0, 0, 0, 0, // Timestamp s
		0, 0, 0, 0, // Timestamp µs
		// Peer Up Notification
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 41, // 10.20.30.41 local address
		0, 123, // Local Port
		0, 234, // Remote Port
		// Sent OPEN message
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 29, // Length
		1,      // Open message type
		4,      // BGP Version
		0, 200, // AS
		0, 180, // Hold Time
		1, 0, 0, 1, // BGP Identifier
		0, // Opt param length
		// Received OPEN message
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 29, // Length
		1,      // Open message type
		4,      // BGP Version
		0, 100, // AS
		0, 180, // Hold Time
		1, 0, 0, 255, // BGP Identifier
		0, // Opt param length

		// #####################################################################
		// Route Monitoring Message #1
		// #####################################################################
		// Common Header
		3,            // Version
		0, 0, 0, 116, // Message Length
		0, // Msg Type = Route Monitoring Message

		// Per Peer Header
		0,                      // Peer Type
		0,                      // Peer Flags
		0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
		0, 0, 0, 100, // Peer AS
		0, 0, 0, 255, // Peer BGP ID
		0, 0, 0, 0, // Timestamp s
		0, 0, 0, 0, // Timestamp µs

		// BGP Update
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 68, // Length
		2, // Update

		0, 0, // Withdrawn Routes Length
		0, 41, // Total Path Attribute Length

		255,  // Attribute flags
		1,    // Attribute Type code (ORIGIN)
		0, 1, // Length
		2, // INCOMPLETE

		0,      // Attribute flags
		2,      // Attribute Type code (AS Path)
		12,     // Length
		2,      // Type = AS_SEQUENCE
		2,      // Path Segment Length
		59, 65, // AS15169
		12, 248, // AS3320
		1,      // Type = AS_SET
		2,      // Path Segment Length
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

		// NLRI
		24,
		192, 168, 0,

		// #####################################################################
		// Termination Message
		// #####################################################################
		// Common Header
		3,           // Version
		0, 0, 0, 12, // Message Length
		5, // Msg Type = Termination Message

		// TLV
		0, 0, 0, 2,
		42, 42,
	})

	time.Sleep(time.Millisecond * 50)
	assert.Empty(t, r.neighbors)

	time.Sleep(time.Millisecond * 50)

	count := rib4.RouteCount()
	if count != 0 {
		t.Errorf("Unexpected route count. Expected: 0 Got: %d", count)
	}
}

func TestIntegrationPeerUpRouteMonitorIPv6WithClientAtEnd(t *testing.T) {
	addr := net.IP{10, 20, 30, 40}
	port := uint16(12346)

	rib4 := locRIB.New("inet.0")
	rib6 := locRIB.New("inet6.0")

	r := newRouter(addr, port, rib4, rib6)
	conA, conB := net.Pipe()

	go r.serve(conB)

	// Peer Up Notification
	_, err := conA.Write([]byte{
		// Common Header
		3,            // Version
		0, 0, 0, 142, // Message Length
		3, // Msg Type = Peer Up Notification

		// Per Peer Header
		0,                      // Peer Type
		0,                      // Peer Flags
		0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
		0, 0, 0, 100, // Peer AS
		0, 0, 0, 255, // Peer BGP ID
		0, 0, 0, 0, // Timestamp s
		0, 0, 0, 0, // Timestamp µs

		// Peer Up Notification
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 41, // 10.20.30.41 local address
		0, 123, // Local Port
		0, 234, // Remote Port

		// Sent OPEN message
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 37, // Length
		1,      // Open message type
		4,      // BGP Version
		0, 200, // AS
		0, 180, // Hold Time
		1, 0, 0, 1, // BGP Identifier
		8, // Opt param length
		2, 6,
		1, 4, // MP BGP
		0, 2, // IPv6
		0, 1, // Unicast

		// Received OPEN message
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 37, // Length
		1,      // Open message type
		4,      // BGP Version
		0, 100, // AS
		0, 180, // Hold Time
		1, 0, 0, 255, // BGP Identifier
		8, // Opt param length
		2, 6,
		1, 4, // MP BGP
		0, 2, // IPv6
		0, 1, // Unicast

		// SECOND MESSAGE:

		// Common Header
		3,            // Version
		0, 0, 0, 138, // Message Length
		0, // Msg Type = Route Monitoring Message

		// Per Peer Header
		0,                      // Peer Type
		0,                      // Peer Flags
		0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
		0, 0, 0, 100, // Peer AS
		0, 0, 0, 255, // Peer BGP ID
		0, 0, 0, 0, // Timestamp s
		0, 0, 0, 0, // Timestamp µs

		// BGP Update
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 90, // Length
		2, // Update

		0, 0, // Withdrawn Routes Length
		0, 67, // Total Path Attribute Length

		255,
		14,    // MP REACH NLRI
		0, 22, // Length
		0, 2, // IPv6
		1,  // Unicast
		16, // IPv6 Next Hop
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0, // Reserved
		0, // Pfxlen /0

		255,  // Attribute flags
		1,    // Attribute Type code (ORIGIN)
		0, 1, // Length
		2, // INCOMPLETE

		0,      // Attribute flags
		2,      // Attribute Type code (AS Path)
		12,     // Length
		2,      // Type = AS_SEQUENCE
		2,      // Path Segment Length
		59, 65, // AS15169
		12, 248, // AS3320
		1,      // Type = AS_SET
		2,      // Path Segment Length
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
	})

	if err != nil {
		panic("write failed")
	}

	time.Sleep(time.Millisecond * 50)
	assert.NotEmpty(t, r.neighbors)

	time.Sleep(time.Millisecond * 50)

	count := rib6.RouteCount()
	if count != 1 {
		t.Errorf("Unexpected IPv6 route count. Expected: 1 Got: %d", count)
	}

	count = rib4.RouteCount()
	if count != 0 {
		t.Errorf("Unexpected IPv4 route count. Expected: 0 Got: %d", count)
	}

	client6 := locRIB.New("client6")
	r.subscribeRIBs(client6, packet.IPv6AFI)

	count = client6.RouteCount()
	if count != 1 {
		t.Errorf("Unexpected IPv6 route count. Expected: 1 Got: %d", count)
	}

	conA.Close()
}

func TestIntegrationPeerUpRouteMonitorIPv6WithClientBeforeBMPPeer(t *testing.T) {
	tests := []struct {
		name               string
		afi                uint8
		unregister         bool
		doubleSubscribe    bool
		doubleUnsubscribe  bool
		input              []byte
		expectedRouteCount int
	}{
		{
			name:       "IPv4 without unregister",
			afi:        packet.IPv4AFI,
			unregister: false,
			input: []byte{
				// Common Header
				3,            // Version
				0, 0, 0, 142, // Message Length
				3, // Msg Type = Peer Up Notification

				// Per Peer Header
				0,                      // Peer Type
				0,                      // Peer Flags
				0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
				0, 0, 0, 100, // Peer AS
				0, 0, 0, 255, // Peer BGP ID
				0, 0, 0, 0, // Timestamp s
				0, 0, 0, 0, // Timestamp µs

				// Peer Up Notification
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 41, // 10.20.30.41 local address
				0, 123, // Local Port
				0, 234, // Remote Port

				// Sent OPEN message
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
				0, 37, // Length
				1,      // Open message type
				4,      // BGP Version
				0, 200, // AS
				0, 180, // Hold Time
				1, 0, 0, 1, // BGP Identifier
				8, // Opt param length
				2, 6,
				1, 4, // MP BGP
				0, 2, // IPv6
				0, 1, // Unicast

				// Received OPEN message
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
				0, 37, // Length
				1,      // Open message type
				4,      // BGP Version
				0, 100, // AS
				0, 180, // Hold Time
				1, 0, 0, 255, // BGP Identifier
				8, // Opt param length
				2, 6,
				1, 4, // MP BGP
				0, 2, // IPv6
				0, 1, // Unicast

				// SECOND MESSAGE:

				// Common Header
				3,            // Version
				0, 0, 0, 116, // Message Length
				0, // Msg Type = Route Monitoring Message

				// Per Peer Header
				0,                      // Peer Type
				0,                      // Peer Flags
				0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
				0, 0, 0, 100, // Peer AS
				0, 0, 0, 255, // Peer BGP ID
				0, 0, 0, 0, // Timestamp s
				0, 0, 0, 0, // Timestamp µs

				// BGP Update
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
				0, 68, // Length
				2, // Update

				0, 0, // Withdrawn Routes Length
				0, 41, // Total Path Attribute Length

				255,  // Attribute flags
				1,    // Attribute Type code (ORIGIN)
				0, 1, // Length
				2, // INCOMPLETE

				0,      // Attribute flags
				2,      // Attribute Type code (AS Path)
				12,     // Length
				2,      // Type = AS_SEQUENCE
				2,      // Path Segment Length
				59, 65, // AS15169
				12, 248, // AS3320
				1,      // Type = AS_SET
				2,      // Path Segment Length
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

				// NLRI
				24,
				192, 168, 0,
			},
			expectedRouteCount: 1,
		},
		{
			name:       "IPv4 with unregister",
			afi:        packet.IPv4AFI,
			unregister: true,
			input: []byte{
				// Common Header
				3,            // Version
				0, 0, 0, 142, // Message Length
				3, // Msg Type = Peer Up Notification

				// Per Peer Header
				0,                      // Peer Type
				0,                      // Peer Flags
				0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
				0, 0, 0, 100, // Peer AS
				0, 0, 0, 255, // Peer BGP ID
				0, 0, 0, 0, // Timestamp s
				0, 0, 0, 0, // Timestamp µs

				// Peer Up Notification
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 41, // 10.20.30.41 local address
				0, 123, // Local Port
				0, 234, // Remote Port

				// Sent OPEN message
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
				0, 37, // Length
				1,      // Open message type
				4,      // BGP Version
				0, 200, // AS
				0, 180, // Hold Time
				1, 0, 0, 1, // BGP Identifier
				8, // Opt param length
				2, 6,
				1, 4, // MP BGP
				0, 2, // IPv6
				0, 1, // Unicast

				// Received OPEN message
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
				0, 37, // Length
				1,      // Open message type
				4,      // BGP Version
				0, 100, // AS
				0, 180, // Hold Time
				1, 0, 0, 255, // BGP Identifier
				8, // Opt param length
				2, 6,
				1, 4, // MP BGP
				0, 2, // IPv6
				0, 1, // Unicast

				// SECOND MESSAGE:

				// Common Header
				3,            // Version
				0, 0, 0, 116, // Message Length
				0, // Msg Type = Route Monitoring Message

				// Per Peer Header
				0,                      // Peer Type
				0,                      // Peer Flags
				0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
				0, 0, 0, 100, // Peer AS
				0, 0, 0, 255, // Peer BGP ID
				0, 0, 0, 0, // Timestamp s
				0, 0, 0, 0, // Timestamp µs

				// BGP Update
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
				0, 68, // Length
				2, // Update

				0, 0, // Withdrawn Routes Length
				0, 41, // Total Path Attribute Length

				255,  // Attribute flags
				1,    // Attribute Type code (ORIGIN)
				0, 1, // Length
				2, // INCOMPLETE

				0,      // Attribute flags
				2,      // Attribute Type code (AS Path)
				12,     // Length
				2,      // Type = AS_SEQUENCE
				2,      // Path Segment Length
				59, 65, // AS15169
				12, 248, // AS3320
				1,      // Type = AS_SET
				2,      // Path Segment Length
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

				// NLRI
				24,
				192, 168, 0,
			},
			expectedRouteCount: 1,
		},
		{
			name:              "IPv4 with double unregister",
			afi:               packet.IPv4AFI,
			doubleUnsubscribe: true,
			unregister:        true,
			input: []byte{
				// Common Header
				3,            // Version
				0, 0, 0, 142, // Message Length
				3, // Msg Type = Peer Up Notification

				// Per Peer Header
				0,                      // Peer Type
				0,                      // Peer Flags
				0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
				0, 0, 0, 100, // Peer AS
				0, 0, 0, 255, // Peer BGP ID
				0, 0, 0, 0, // Timestamp s
				0, 0, 0, 0, // Timestamp µs

				// Peer Up Notification
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 41, // 10.20.30.41 local address
				0, 123, // Local Port
				0, 234, // Remote Port

				// Sent OPEN message
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
				0, 37, // Length
				1,      // Open message type
				4,      // BGP Version
				0, 200, // AS
				0, 180, // Hold Time
				1, 0, 0, 1, // BGP Identifier
				8, // Opt param length
				2, 6,
				1, 4, // MP BGP
				0, 2, // IPv6
				0, 1, // Unicast

				// Received OPEN message
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
				0, 37, // Length
				1,      // Open message type
				4,      // BGP Version
				0, 100, // AS
				0, 180, // Hold Time
				1, 0, 0, 255, // BGP Identifier
				8, // Opt param length
				2, 6,
				1, 4, // MP BGP
				0, 2, // IPv6
				0, 1, // Unicast

				// SECOND MESSAGE:

				// Common Header
				3,            // Version
				0, 0, 0, 116, // Message Length
				0, // Msg Type = Route Monitoring Message

				// Per Peer Header
				0,                      // Peer Type
				0,                      // Peer Flags
				0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
				0, 0, 0, 100, // Peer AS
				0, 0, 0, 255, // Peer BGP ID
				0, 0, 0, 0, // Timestamp s
				0, 0, 0, 0, // Timestamp µs

				// BGP Update
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
				0, 68, // Length
				2, // Update

				0, 0, // Withdrawn Routes Length
				0, 41, // Total Path Attribute Length

				255,  // Attribute flags
				1,    // Attribute Type code (ORIGIN)
				0, 1, // Length
				2, // INCOMPLETE

				0,      // Attribute flags
				2,      // Attribute Type code (AS Path)
				12,     // Length
				2,      // Type = AS_SEQUENCE
				2,      // Path Segment Length
				59, 65, // AS15169
				12, 248, // AS3320
				1,      // Type = AS_SET
				2,      // Path Segment Length
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

				// NLRI
				24,
				192, 168, 0,
			},
			expectedRouteCount: 1,
		},
		{
			name:            "IPv4 with double register",
			afi:             packet.IPv4AFI,
			doubleSubscribe: true,
			unregister:      true,
			input: []byte{
				// Common Header
				3,            // Version
				0, 0, 0, 142, // Message Length
				3, // Msg Type = Peer Up Notification

				// Per Peer Header
				0,                      // Peer Type
				0,                      // Peer Flags
				0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
				0, 0, 0, 100, // Peer AS
				0, 0, 0, 255, // Peer BGP ID
				0, 0, 0, 0, // Timestamp s
				0, 0, 0, 0, // Timestamp µs

				// Peer Up Notification
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 41, // 10.20.30.41 local address
				0, 123, // Local Port
				0, 234, // Remote Port

				// Sent OPEN message
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
				0, 37, // Length
				1,      // Open message type
				4,      // BGP Version
				0, 200, // AS
				0, 180, // Hold Time
				1, 0, 0, 1, // BGP Identifier
				8, // Opt param length
				2, 6,
				1, 4, // MP BGP
				0, 2, // IPv6
				0, 1, // Unicast

				// Received OPEN message
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
				0, 37, // Length
				1,      // Open message type
				4,      // BGP Version
				0, 100, // AS
				0, 180, // Hold Time
				1, 0, 0, 255, // BGP Identifier
				8, // Opt param length
				2, 6,
				1, 4, // MP BGP
				0, 2, // IPv6
				0, 1, // Unicast

				// SECOND MESSAGE:

				// Common Header
				3,            // Version
				0, 0, 0, 116, // Message Length
				0, // Msg Type = Route Monitoring Message

				// Per Peer Header
				0,                      // Peer Type
				0,                      // Peer Flags
				0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
				0, 0, 0, 100, // Peer AS
				0, 0, 0, 255, // Peer BGP ID
				0, 0, 0, 0, // Timestamp s
				0, 0, 0, 0, // Timestamp µs

				// BGP Update
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
				0, 68, // Length
				2, // Update

				0, 0, // Withdrawn Routes Length
				0, 41, // Total Path Attribute Length

				255,  // Attribute flags
				1,    // Attribute Type code (ORIGIN)
				0, 1, // Length
				2, // INCOMPLETE

				0,      // Attribute flags
				2,      // Attribute Type code (AS Path)
				12,     // Length
				2,      // Type = AS_SEQUENCE
				2,      // Path Segment Length
				59, 65, // AS15169
				12, 248, // AS3320
				1,      // Type = AS_SET
				2,      // Path Segment Length
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

				// NLRI
				24,
				192, 168, 0,
			},
			expectedRouteCount: 1,
		},
		{
			name:       "IPv6 without unregister",
			afi:        packet.IPv6AFI,
			unregister: false,
			input: []byte{
				// Common Header
				3,            // Version
				0, 0, 0, 142, // Message Length
				3, // Msg Type = Peer Up Notification

				// Per Peer Header
				0,                      // Peer Type
				0,                      // Peer Flags
				0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
				0, 0, 0, 100, // Peer AS
				0, 0, 0, 255, // Peer BGP ID
				0, 0, 0, 0, // Timestamp s
				0, 0, 0, 0, // Timestamp µs

				// Peer Up Notification
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 41, // 10.20.30.41 local address
				0, 123, // Local Port
				0, 234, // Remote Port

				// Sent OPEN message
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
				0, 37, // Length
				1,      // Open message type
				4,      // BGP Version
				0, 200, // AS
				0, 180, // Hold Time
				1, 0, 0, 1, // BGP Identifier
				8, // Opt param length
				2, 6,
				1, 4, // MP BGP
				0, 2, // IPv6
				0, 1, // Unicast

				// Received OPEN message
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
				0, 37, // Length
				1,      // Open message type
				4,      // BGP Version
				0, 100, // AS
				0, 180, // Hold Time
				1, 0, 0, 255, // BGP Identifier
				8, // Opt param length
				2, 6,
				1, 4, // MP BGP
				0, 2, // IPv6
				0, 1, // Unicast

				// SECOND MESSAGE:

				// Common Header
				3,            // Version
				0, 0, 0, 138, // Message Length
				0, // Msg Type = Route Monitoring Message

				// Per Peer Header
				0,                      // Peer Type
				0,                      // Peer Flags
				0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
				0, 0, 0, 100, // Peer AS
				0, 0, 0, 255, // Peer BGP ID
				0, 0, 0, 0, // Timestamp s
				0, 0, 0, 0, // Timestamp µs

				// BGP Update
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
				0, 90, // Length
				2, // Update

				0, 0, // Withdrawn Routes Length
				0, 67, // Total Path Attribute Length

				255,
				14,    // MP REACH NLRI
				0, 22, // Length
				0, 2, // IPv6
				1,  // Unicast
				16, // IPv6 Next Hop
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0, // Reserved
				0, // Pfxlen /0

				255,  // Attribute flags
				1,    // Attribute Type code (ORIGIN)
				0, 1, // Length
				2, // INCOMPLETE

				0,      // Attribute flags
				2,      // Attribute Type code (AS Path)
				12,     // Length
				2,      // Type = AS_SEQUENCE
				2,      // Path Segment Length
				59, 65, // AS15169
				12, 248, // AS3320
				1,      // Type = AS_SET
				2,      // Path Segment Length
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
			expectedRouteCount: 1,
		},
	}

	for _, test := range tests {
		addr := net.IP{10, 20, 30, 40}
		port := uint16(12346)

		rib4 := locRIB.New("inet.0")
		rib6 := locRIB.New("inet6.0")

		r := newRouter(addr, port, rib4, rib6)
		conA, conB := net.Pipe()

		client := locRIB.New("client")
		r.subscribeRIBs(client, test.afi)
		if test.doubleSubscribe {
			r.subscribeRIBs(client, test.afi)
		}

		if test.unregister {
			r.unsubscribeRIBs(client, test.afi)
			if test.doubleUnsubscribe {
				r.unsubscribeRIBs(client, test.afi)
			}
		}

		go r.serve(conB)

		_, err := conA.Write(test.input)

		if err != nil {
			panic("write failed")
		}

		time.Sleep(time.Millisecond * 50)

		expectedCount := int64(1)
		if test.unregister {
			expectedCount = 0
		}

		count := client.RouteCount()
		if count != expectedCount {
			t.Errorf("Unexpected route count for test %q. Expected: %d Got: %d", test.name, expectedCount, count)
		}

		conA.Close()
	}
}
*/
