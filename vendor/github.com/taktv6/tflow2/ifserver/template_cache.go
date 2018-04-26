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

package ifserver

import (
	"sync"

	"github.com/taktv6/tflow2/ipfix"
)

type templateCache struct {
	cache map[uint32]map[uint32]map[uint16]ipfix.TemplateRecords
	lock  sync.RWMutex
}

// newTemplateCache creates and initializes a new `templateCache` instance
func newTemplateCache() *templateCache {
	return &templateCache{cache: make(map[uint32]map[uint32]map[uint16]ipfix.TemplateRecords)}
}

func (c *templateCache) set(rtr uint32, domainID uint32, templateID uint16, records ipfix.TemplateRecords) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if _, ok := c.cache[rtr]; !ok {
		c.cache[rtr] = make(map[uint32]map[uint16]ipfix.TemplateRecords)
	}
	if _, ok := c.cache[rtr][domainID]; !ok {
		c.cache[rtr][domainID] = make(map[uint16]ipfix.TemplateRecords)
	}
	c.cache[rtr][domainID][templateID] = records
}

func (c *templateCache) get(rtr uint32, domainID uint32, templateID uint16) *ipfix.TemplateRecords {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if _, ok := c.cache[rtr]; !ok {
		return nil
	}
	if _, ok := c.cache[rtr][domainID]; !ok {
		return nil
	}
	if _, ok := c.cache[rtr][domainID][templateID]; !ok {
		return nil
	}
	ret := c.cache[rtr][domainID][templateID]
	return &ret
}
