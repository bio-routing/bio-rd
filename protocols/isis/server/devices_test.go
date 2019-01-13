package server

import (
	"testing"

	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/protocols/device"
	"github.com/stretchr/testify/assert"
)

func TestRemoveDevice(t *testing.T) {
	tests := []struct {
		name           string
		db             *devices
		removeName     string
		wantFail       bool
		expected       *devices
		wantUnregister bool
	}{
		{
			name: "Remove existing",
			db: &devices{
				srv: &Server{
					ds: &device.MockServer{},
				},
				db: map[string]*dev{
					"foobar": &dev{
						done: make(chan struct{}),
						srv: &Server{
							sys: &mockSys{},
						},
						name: "foobar",
					},
				},
			},
			removeName: "foobar",
			expected: &devices{
				srv: &Server{
					ds: &device.MockServer{
						UnsubscribeCalled: true,
						UnsubscribeName:   "foobar",
					},
				},
				db: map[string]*dev{},
			},
			wantUnregister: true,
		},
		{
			name: "Remove non-existing",
			db: &devices{
				srv: &Server{
					ds: &device.MockServer{},
				},
				db: map[string]*dev{
					"foobar": &dev{
						done: make(chan struct{}),
						srv: &Server{
							sys: &mockSys{},
						},
						name: "foobar",
					},
				},
			},
			removeName: "baz",
			expected: &devices{
				srv: &Server{
					ds: &device.MockServer{},
				},
				db: map[string]*dev{
					"foobar": &dev{
						done: make(chan struct{}),
						srv: &Server{
							sys: &mockSys{},
						},
						name: "foobar",
					},
				},
			},
			wantUnregister: false,
			wantFail:       true,
		},
		{
			name: "Remove existing - disable fails",
			db: &devices{
				srv: &Server{
					ds: &device.MockServer{},
				},
				db: map[string]*dev{
					"foobar": &dev{
						done: make(chan struct{}),
						srv: &Server{
							sys: &mockSys{
								wantFailClosedPacketSocket: true,
							},
						},
						name: "foobar",
					},
				},
			},
			removeName: "foobar",
			expected: &devices{
				srv: &Server{
					ds: &device.MockServer{
						UnsubscribeCalled: true,
						UnsubscribeName:   "foobar",
					},
				},
				db: map[string]*dev{},
			},
			wantUnregister: true,
			wantFail:       true,
		},
	}

	for _, test := range tests {
		err := test.db.removeDevice(test.removeName)

		assert.Equal(t, test.wantUnregister, test.db.srv.ds.(*device.MockServer).UnsubscribeCalled, test.name)

		if err != nil {
			if test.wantFail {
				continue
			}

			t.Errorf("Unexpected failure for test %q: %v", test.name, err)
			continue
		}

		if test.wantFail {
			t.Errorf("Unexpected success for test %q", test.name)
		}

		// Ignore some attributes
		for i := range test.db.db {
			test.db.db[i].srv = nil
			test.db.db[i].helloMethod = nil
			test.db.db[i].receiverMethod = nil
			test.db.db[i].done = nil
		}

		assert.Equal(t, test.expected, test.db, test.name)
	}
}

func TestDeviceAddDevice(t *testing.T) {
	tests := []struct {
		name         string
		db           *devices
		addIfCfg     *config.ISISInterfaceConfig
		wantFail     bool
		expected     *devices
		wantRegister bool
	}{
		{
			name: "Test #1",
			db: &devices{
				srv: &Server{
					ds: &device.MockServer{},
				},
				db: map[string]*dev{
					"foobar": &dev{
						name: "foobar",
					},
				},
			},
			addIfCfg: &config.ISISInterfaceConfig{
				Name:    "baz",
				Passive: true,
			},
			expected: &devices{
				srv: &Server{
					ds: &device.MockServer{
						Called: true,
						Name:   "baz",
					},
				},
				db: map[string]*dev{
					"foobar": &dev{
						name: "foobar",
					},
					"baz": &dev{
						name:               "baz",
						passive:            true,
						supportedProtocols: []uint8{0xcc, 0x8e},
					},
				},
			},
			wantRegister: true,
		},
		{
			name: "Test #2",
			db: &devices{
				srv: &Server{
					ds: &device.MockServer{},
				},
				db: map[string]*dev{
					"foobar": &dev{
						name: "foobar",
					},
				},
			},
			addIfCfg: &config.ISISInterfaceConfig{
				Name:    "foobar",
				Passive: true,
			},
			expected: &devices{
				srv: &Server{
					ds: &device.MockServer{
						Called: true,
						Name:   "baz",
					},
				},
				db: map[string]*dev{
					"foobar": &dev{
						name: "foobar",
					},
				},
			},
			wantRegister: false,
			wantFail:     true,
		},
	}

	for _, test := range tests {
		err := test.db.addDevice(test.addIfCfg)

		assert.Equal(t, test.wantRegister, test.db.srv.ds.(*device.MockServer).Called, test.name)

		if err != nil {
			if test.wantFail {
				continue
			}

			t.Errorf("Unexpected failure for test %q: %v", test.name, err)
			continue
		}

		if test.wantFail {
			t.Errorf("Unexpected success for test %q", test.name)
		}

		// Ignore some attributes
		for i := range test.db.db {
			test.db.db[i].srv = nil
			test.db.db[i].helloMethod = nil
			test.db.db[i].receiverMethod = nil
			test.db.db[i].done = nil
		}

		assert.Equal(t, test.expected, test.db, test.name)
	}
}
