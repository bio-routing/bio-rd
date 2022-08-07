package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommunityStringFromUint32(t *testing.T) {
	tests := []struct {
		name     string
		value    uint32
		expected string
	}{
		{
			name:     "both elements",
			value:    131080,
			expected: "(2,8)",
		},
		{
			name:     "right element only",
			value:    250,
			expected: "(0,250)",
		},
		{
			name:     "left element only",
			value:    131072,
			expected: "(2,0)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(te *testing.T) {
			assert.Equal(te, test.expected, CommunityStringForUint32(test.value))
		})
	}
}

func TestParseCommunityString(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected uint32
		wantFail bool
	}{
		{
			name:     "both elements",
			expected: 131080,
			value:    "(2,8)",
		},
		{
			name:     "right element only",
			expected: 250,
			value:    "(0,250)",
		},
		{
			name:     "left element only",
			expected: 131072,
			value:    "(2,0)",
		},
		{
			name:     "too big",
			value:    "(131072,256)",
			wantFail: true,
		},
		{
			name:     "bad element in brackets",
			value:    "(131072,256a)",
			wantFail: true,
		},
		{
			name:     "empty string",
			value:    "",
			wantFail: true,
		},
		{
			name:     "random string",
			value:    "foo-bar",
			wantFail: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(te *testing.T) {
			c, err := ParseCommunityString(test.value)

			if test.wantFail {
				if err == nil {
					te.Fatal("test was expected to fail, but did not")
				}

				return
			}

			assert.Equal(te, test.expected, c)
		})
	}
}

func TestCommunityToString(t *testing.T) {
	tests := []struct {
		name     string
		value    *Communities
		expected string
	}{
		{
			name:     "nil",
			expected: "",
			value:    nil,
		},
		{
			name:     "one elememt",
			expected: "250",
			value:    &Communities{250},
		},
		{
			name:     "two elements",
			expected: "131080 2342",
			value:    &Communities{131080, 2342},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(te *testing.T) {
			assert.Equal(te, test.expected, test.value.String())
		})
	}
}
