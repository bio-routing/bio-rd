package routingtable

import (
	"sync"
)

type contributingASN struct {
	asn   uint32
	count uint32
}

// ContributingASNs contains a list of contributing ASN to a LocRIB to check ASPaths for possible routing loops.
type ContributingASNs struct {
	contributingASNs []*contributingASN
	mu               sync.RWMutex
}

// NewContributingASNs creates a list of contributing ASNs to a LocRIB for routing loop prevention.
func NewContributingASNs() *ContributingASNs {
	c := &ContributingASNs{
		contributingASNs: []*contributingASN{},
	}

	return c
}

// Add a new ASN to the list of contributing ASNs or add the ref count of an existing one.
func (c *ContributingASNs) Add(asn uint32) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, cASN := range c.contributingASNs {
		if cASN.asn == asn {
			cASN.count++
			return
		}
	}

	c.contributingASNs = append(c.contributingASNs, &contributingASN{
		asn:   asn,
		count: 1,
	})
}

// Remove a ASN to the list of contributing ASNs or decrement the ref count of an existing one.
func (c *ContributingASNs) Remove(asn uint32) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	asnList := c.contributingASNs

	for i, cASN := range asnList {
		if cASN.asn == asn {
			cASN.count--

			if cASN.count == 0 {
				copy(asnList[i:], asnList[i+1:])
				asnList = asnList[:len(asnList)]
				c.contributingASNs = asnList[:len(asnList)-1]
			}
			return
		}
	}
}

// IsContributingASN checks if  a given ASN is part of the contributing ASNs
func (c *ContributingASNs) IsContributingASN(asn uint32) bool {
	for _, cASN := range c.contributingASNs {
		if asn == cASN.asn {
			return true
		}
	}

	return false
}
