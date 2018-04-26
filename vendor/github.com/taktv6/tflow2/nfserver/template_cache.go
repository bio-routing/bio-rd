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

package nfserver

import (
	"sync"

	"github.com/taktv6/tflow2/nf9"
)

type templateCache struct {
	cache map[uint32]map[uint32]map[uint16]nf9.TemplateRecords
	lock  sync.RWMutex
}

// newTemplateCache creates and initializes a new `templateCache` instance
func newTemplateCache() *templateCache {
	return &templateCache{cache: make(map[uint32]map[uint32]map[uint16]nf9.TemplateRecords)}
}

func (c *templateCache) set(rtr uint32, sourceID uint32, templateID uint16, records nf9.TemplateRecords) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if _, ok := c.cache[rtr]; !ok {
		c.cache[rtr] = make(map[uint32]map[uint16]nf9.TemplateRecords)
	}
	if _, ok := c.cache[rtr][sourceID]; !ok {
		c.cache[rtr][sourceID] = make(map[uint16]nf9.TemplateRecords)
	}
	c.cache[rtr][sourceID][templateID] = records
}

func (c *templateCache) get(rtr uint32, sourceID uint32, templateID uint16) *nf9.TemplateRecords {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if _, ok := c.cache[rtr]; !ok {
		return nil
	}
	if _, ok := c.cache[rtr][sourceID]; !ok {
		return nil
	}
	if _, ok := c.cache[rtr][sourceID][templateID]; !ok {
		return nil
	}
	ret := c.cache[rtr][sourceID][templateID]
	return &ret
}
