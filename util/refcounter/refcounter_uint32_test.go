package refcounter

import (
	"fmt"
	"testing"
)

func TestFancy(t *testing.T) {
	r := NewRefCounterUint32()

	tests := []struct {
		runCmd func()
		expect func() bool
		msg    string
	}{
		// Empty list
		{
			runCmd: func() {},
			expect: func() bool { return !r.IsPresent(41981) },
			msg:    "41981 shouldn't be present yet.",
		},

		// Add and remove one item
		{
			runCmd: func() { r.Add(41981) },
			expect: func() bool { return r.IsPresent(41981) },
			msg:    "41981 should be contributing.",
		},
		{
			runCmd: func() { r.Remove(41981) },
			expect: func() bool { return !r.IsPresent(41981) },
			msg:    "41981 shouldn't be contributing no more.",
		},

		// Two items present
		{
			runCmd: func() { r.Add(41981) },
			expect: func() bool { return r.IsPresent(41981) },
			msg:    "41981 should be contributing.",
		},
		{
			runCmd: func() { r.Add(201701) },
			expect: func() bool { return r.IsPresent(41981) },
			msg:    "201701 should be contributing.",
		},

		// Add 41981 2nd time
		{
			runCmd: func() { r.Add(41981) },
			expect: func() bool { return r.IsPresent(41981) },
			msg:    "41981 should be still contributing.",
		},
		{
			runCmd: func() {},
			expect: func() bool { return r.items[0].value == 41981 },
			msg:    "41981 is first item in list.",
		},
		{
			runCmd: func() { fmt.Printf("%+v", r.items) },
			expect: func() bool { return r.items[0].count == 2 },
			msg:    "41981 should be present twice.",
		},

		// Remove 2nd 41981
		{
			runCmd: func() { r.Remove(41981) },
			expect: func() bool { return r.IsPresent(41981) },
			msg:    "41981 should still be contributing.",
		},
		{
			runCmd: func() {},
			expect: func() bool { return r.items[0].count == 1 },
			msg:    "41981 should be present once.",
		},

		// Remove 201701
		{
			runCmd: func() { r.Remove(201701) },
			expect: func() bool { return !r.IsPresent(201701) },
			msg:    "201701 shouldn't be contributing no more.",
		},
	}

	for i, test := range tests {
		test.runCmd()
		if !test.expect() {
			t.Errorf("Test %d failed: %v", i, test.msg)
		}
	}
}
