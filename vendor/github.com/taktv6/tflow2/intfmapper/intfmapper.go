package intfmapper

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/taktv6/tflow2/config"

	g "github.com/soniah/gosnmp"
)

const (
	ifNameOID = "1.3.6.1.2.1.31.1.1.1.1"
)

type IntfMapperInterface interface {
	GetInterfaceIDByName(agent string) InterfaceIDByName
	GetInterfaceNameByID(agent string) InterfaceNameByID
}

// Mapper is a service that maps agents interface IDs to names
type Mapper struct {
	agents                     []config.Agent
	renewInterval              int64
	interfaceIDByNameByAgent   map[string]InterfaceIDByName
	interfaceNameByIDByAgent   map[string]InterfaceNameByID
	interfaceIDByNameByAgentMu sync.RWMutex
	interfaceNameByIDByAgentMu sync.RWMutex
}

// InterfaceIDByName maps interface names to IDs
type InterfaceIDByName map[string]uint16

// InterfaceNameByID maps IDs to interface names
type InterfaceNameByID map[uint16]string

// New creates a new Mapper and starts workers for all agents that periodicly renew interface mappings
func New(agents []config.Agent, renewInterval int64) (*Mapper, error) {
	m := &Mapper{
		agents:                   agents,
		renewInterval:            renewInterval,
		interfaceIDByNameByAgent: make(map[string]InterfaceIDByName),
		interfaceNameByIDByAgent: make(map[string]InterfaceNameByID),
	}

	for _, agent := range m.agents {
		m.interfaceIDByNameByAgent[*agent.Name] = make(InterfaceIDByName)
		if err := m.renewMapping(agent); err != nil {
			return nil, fmt.Errorf("Unable to get interface mapping for %s: %v", *agent.Name, err)
		}
	}

	m.startRenewWorkers()

	return m, nil
}

func (m *Mapper) startRenewWorkers() {
	for _, agent := range m.agents {
		go func(agent config.Agent) {
			for {
				time.Sleep(time.Second * time.Duration(m.renewInterval))
				err := m.renewMapping(agent)
				if err != nil {
					glog.Infof("Unable to renew interface mapping for %s: %v", *agent.Name, err)
				}
			}
		}(agent)
	}
}

func (m *Mapper) renewMapping(a config.Agent) error {
	snmpClient := g.Default
	snmpClient.Target = *a.IPAddress
	snmpClient.Community = *a.SNMPCommunity

	if err := snmpClient.Connect(); err != nil {
		return fmt.Errorf("SNMP client unable to connect: %v", err)
	}
	defer snmpClient.Conn.Close()

	newMapByName := make(InterfaceIDByName)
	err := snmpClient.BulkWalk(ifNameOID, newMapByName.update)
	if err != nil {
		return fmt.Errorf("Walk error: %v", err)
	}

	newMapByID := make(InterfaceNameByID)
	for name, id := range newMapByName {
		newMapByID[id] = name
	}

	m.interfaceIDByNameByAgentMu.Lock()
	defer m.interfaceIDByNameByAgentMu.Unlock()

	m.interfaceIDByNameByAgent[*a.Name] = newMapByName
	m.interfaceNameByIDByAgent[*a.Name] = newMapByID

	return nil
}

func (im InterfaceIDByName) update(pdu g.SnmpPDU) error {
	oid := strings.Split(pdu.Name, ".")
	id, err := strconv.Atoi(oid[len(oid)-1])
	if err != nil {
		return fmt.Errorf("Unable to convert interface id: %v", err)
	}

	if pdu.Type != g.OctetString {
		return fmt.Errorf("Unexpected PDU type: %d", pdu.Type)
	}

	im[string(pdu.Value.([]byte))] = uint16(id)

	return nil
}

// GetInterfaceIDByName gets the InterfaceIDByName
func (m *Mapper) GetInterfaceIDByName(agent string) InterfaceIDByName {
	m.interfaceIDByNameByAgentMu.RLock()
	defer m.interfaceIDByNameByAgentMu.RUnlock()

	ret := make(InterfaceIDByName)
	for key, value := range m.interfaceIDByNameByAgent[agent] {
		ret[key] = value
	}

	return ret
}

// GetInterfaceNameByID gets the InterfaceNameByID
func (m *Mapper) GetInterfaceNameByID(agent string) InterfaceNameByID {
	m.interfaceNameByIDByAgentMu.RLock()
	defer m.interfaceNameByIDByAgentMu.RUnlock()

	ret := make(InterfaceNameByID)
	for key, value := range m.interfaceNameByIDByAgent[agent] {
		ret[key] = value
	}

	return ret
}
