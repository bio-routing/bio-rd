package routingtable

import (
	"fmt"
	"testing"
)

func TestFancy(t *testing.T) {
	c := NewContributingASNs()

	tests := []struct {
		runCmd func()
		expect func() bool
		msg    string
	}{
		// Empty list
		{
			runCmd: func() {},
			expect: func() bool { return !c.IsContributingASN(41981) },
			msg:    "AS41981 shouldn't be contributing yet.",
		},

		// Add and remove one ASN
		{
			runCmd: func() { c.Add(41981) },
			expect: func() bool { return c.IsContributingASN(41981) },
			msg:    "AS41981 should be contributing.",
		},
		{
			runCmd: func() { c.Remove(41981) },
			expect: func() bool { return !c.IsContributingASN(41981) },
			msg:    "AS41981 shouldn't be contributing no more.",
		},

		// Two ASNs present
		{
			runCmd: func() { c.Add(41981) },
			expect: func() bool { return c.IsContributingASN(41981) },
			msg:    "AS41981 should be contributing.",
		},
		{
			runCmd: func() { c.Add(201701) },
			expect: func() bool { return c.IsContributingASN(41981) },
			msg:    "AS201701 should be contributing.",
		},

		// Add AS41981 2nd time
		{
			runCmd: func() { c.Add(41981) },
			expect: func() bool { return c.IsContributingASN(41981) },
			msg:    "AS41981 should be still contributing.",
		},
		{
			runCmd: func() {},
			expect: func() bool { return c.contributingASNs[0].asn == 41981 },
			msg:    "AS41981 is first ASN in list.",
		},
		{
			runCmd: func() { fmt.Printf("%+v", c.contributingASNs) },
			expect: func() bool { return c.contributingASNs[0].count == 2 },
			msg:    "AS41981 should be present twice.",
		},

		// Remove 2nd AS41981
		{
			runCmd: func() { c.Remove(41981) },
			expect: func() bool { return c.IsContributingASN(41981) },
			msg:    "AS41981 should still be contributing.",
		},
		{
			runCmd: func() {},
			expect: func() bool { return c.contributingASNs[0].count == 1 },
			msg:    "S41981 should be present once.",
		},

		// Remove AS201701
		{
			runCmd: func() { c.Remove(201701) },
			expect: func() bool { return !c.IsContributingASN(201701) },
			msg:    "AS201701 shouldn't be contributing no more.",
		},
	}

	for i, test := range tests {
		test.runCmd()
		if !test.expect() {
			t.Errorf("Test %d failed: %v", i, test.msg)
		}
	}
}
