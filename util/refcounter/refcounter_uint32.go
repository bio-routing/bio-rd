package refcounter

import (
	"fmt"
	"math"
	"sync"
)

type item struct {
	value uint32
	count uint32
}

// RefcounterUint32 contains a list of items to keep refcounts on
type RefcounterUint32 struct {
	items   []*item
	itemsMu sync.RWMutex
}

// NewRefCounterUint32 creates a list of items to keep refcounts on
func NewRefCounterUint32() *RefcounterUint32 {
	c := &RefcounterUint32{
		items: []*item{},
	}

	return c
}

// Add adds new item to the list of items or add the ref count of an existing one.
func (r *RefcounterUint32) Add(value uint32) {
	r.itemsMu.Lock()
	defer r.itemsMu.Unlock()

	for _, iterItem := range r.items {
		if iterItem.value == value {
			iterItem.count++

			if iterItem.count == math.MaxUint32 {
				panic(fmt.Sprintf("Counter overflow triggered for item %d. Dying of shame.", value))
			}

			return
		}
	}

	r.items = append(r.items, &item{
		value: value,
		count: 1,
	})
}

// Remove a value from the list of items or decrement the ref count of an existing one.
func (r *RefcounterUint32) Remove(value uint32) {
	r.itemsMu.Lock()
	defer r.itemsMu.Unlock()

	itemList := r.items

	for i, iterItem := range itemList {
		if iterItem.value != value {
			continue
		}

		iterItem.count--

		if iterItem.count == 0 {
			copy(itemList[i:], itemList[i+1:])
			itemList = itemList[:]
			r.items = itemList[:len(itemList)-1]
		}

		return
	}
}

// IsPresent checks if a given value is part of the known items
func (r *RefcounterUint32) IsPresent(value uint32) bool {
	r.itemsMu.RLock()
	defer r.itemsMu.RUnlock()

	for _, iterItem := range r.items {
		if value == iterItem.value {
			return true
		}
	}

	return false
}
