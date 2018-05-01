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

package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIPByteSlice(t *testing.T) {
	tests := []struct {
		address string
		wanted  []byte
	}{
		{
			address: "192.168.0.1",
			wanted:  []byte{192, 168, 0, 1},
		},
		{
			address: "255.255.255.255",
			wanted:  []byte{255, 255, 255, 255},
		},
		{
			address: "ffff::ff",
			wanted:  []byte{255, 255, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255},
		},
	}

	for _, test := range tests {
		res := IPByteSlice(test.address)
		if !sliceEq(res, test.wanted) {
			t.Errorf("Expected: %d, got: %d", test.wanted, res)
		}
	}
}

func TestUint16b(t *testing.T) {
	tests := []struct {
		input  []byte
		wanted uint16
	}{
		{
			input:  []byte{2, 4},
			wanted: 516,
		},
		{
			input:  []byte{0, 22},
			wanted: 22,
		},
	}

	for _, test := range tests {
		res := Uint16b(test.input)
		if res != test.wanted {
			t.Errorf("Expected: %d, got: %d", test.wanted, res)
		}
	}
}

func TestUint32b(t *testing.T) {
	tests := []struct {
		input  []byte
		wanted uint32
	}{
		{
			input:  []byte{2, 3, 4, 0},
			wanted: 33752064,
		},
		{
			input:  []byte{0, 1, 0, 0},
			wanted: 65536,
		},
	}

	for _, test := range tests {
		res := Uint32b(test.input)
		if res != test.wanted {
			t.Errorf("Expected: %d, got: %d", test.wanted, res)
		}
	}
}

func TestUint64b(t *testing.T) {
	tests := []struct {
		input  []byte
		wanted uint64
	}{
		{
			input:  []byte{0, 0, 0, 0, 2, 3, 4, 0},
			wanted: 33752064,
		},
		{
			input:  []byte{0, 0, 0, 0, 0, 1, 0, 0},
			wanted: 65536,
		},
		{
			input:  []byte{0, 0, 0, 1, 0, 0, 0, 0},
			wanted: 4294967296,
		},
	}

	for _, test := range tests {
		res := Uint64b(test.input)
		if res != test.wanted {
			t.Errorf("Expected: %d, got: %d", test.wanted, res)
		}
	}
}

func TestUintX(t *testing.T) {
	tests := []struct {
		input  []byte
		wanted uint64
	}{
		{
			input:  []byte{0, 0, 0, 0, 2, 3, 4, 0},
			wanted: 1129207031660544,
		},
		{
			input:  []byte{0, 0, 0, 0, 0, 1, 0, 0},
			wanted: 1099511627776,
		},
		{
			input:  []byte{0, 0, 0, 1, 0, 0, 0, 0},
			wanted: 16777216,
		},
	}

	for _, test := range tests {
		res := UintX(test.input)
		if res != test.wanted {
			t.Errorf("Expected: %d, got: %d", test.wanted, res)
		}
	}
}

func TestReverse(t *testing.T) {
	tests := []struct {
		input  []byte
		wanted []byte
	}{
		{
			input:  []byte{1, 2, 3, 4},
			wanted: []byte{4, 3, 2, 1},
		},
	}

	for _, test := range tests {
		res := Reverse(test.input)
		if !sliceEq(res, test.wanted) {
			t.Errorf("Expected: %d, got: %d", test.wanted, res)
		}
	}
}

func sliceEq(a []byte, b []byte) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestUint16Byte(t *testing.T) {
	tests := []struct {
		name     string
		input    uint16
		expected []byte
	}{
		{
			name:     "Test #1",
			input:    23,
			expected: []byte{0, 23},
		},
		{
			name:     "Test #1",
			input:    256,
			expected: []byte{1, 0},
		},
	}

	for _, test := range tests {
		res := Uint16Byte(test.input)
		assert.Equal(t, test.expected, res)
	}
}
