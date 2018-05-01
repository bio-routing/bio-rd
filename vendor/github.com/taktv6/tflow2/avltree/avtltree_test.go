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

package avltree

import "testing"

func testIsSmaller(c1 interface{}, c2 interface{}) bool {
	if c1.(int) < c2.(int) {
		return true
	}
	return false
}

func TestInsert(t *testing.T) {
	values := [...]int{100, 50, 150, 160, 170, 180, 90, 80, 70, 50, 60, 54, 32, 12, 5, 1}
	var tree *TreeNode

	for val := range values {
		tree.insert(val, val, testIsSmaller)
	}
}

func TestIntersection(t *testing.T) {
	candidates := make([]*Tree, 0)
	valuesA := [...]int{20, 100, 50, 150, 160, 170, 180, 90, 15, 80, 70, 50, 60, 54, 32, 12, 5, 1}
	valuesB := [...]int{101, 51, 150, 160, 171, 182, 89, 80, 75, 15, 53, 20, 1}
	valuesC := [...]int{20, 100, 50, 150, 15, 160, 170, 180, 90, 80, 70, 50, 60, 54, 32, 12, 5, 1}
	valuesD := [...]int{20, 101, 51, 150, 171, 182, 89, 80, 75, 15, 53, 20, 1}
	valuesE := [...]int{15, 20, 100, 50, 150, 160, 170, 180, 90, 80, 70, 50, 60, 54, 32, 12, 5, 1}
	valuesCommon := [...]int{150, 80, 1, 20, 15}

	treeA := New()
	treeB := New()
	treeC := New()
	treeD := New()
	treeE := New()

	for _, val := range valuesA {
		treeA.Insert(val, val, testIsSmaller)
	}

	for _, val := range valuesB {
		treeB.Insert(val, val, testIsSmaller)
	}

	for _, val := range valuesC {
		treeC.Insert(val, val, testIsSmaller)
	}

	for _, val := range valuesD {
		treeD.Insert(val, val, testIsSmaller)
	}

	for _, val := range valuesE {
		treeE.Insert(val, val, testIsSmaller)
	}

	candidates = append(candidates, treeA)
	candidates = append(candidates, treeB)
	candidates = append(candidates, treeC)
	candidates = append(candidates, treeD)
	candidates = append(candidates, treeE)

	res := Intersection(candidates)
	for _, val := range valuesCommon {
		if !res.Exists(val) {
			t.Errorf("Element %d not found in common elements tree\n", val)
		}
	}

}

func TestNodeExists(t *testing.T) {
	tests := []struct {
		input int
		want  bool
	}{
		{
			input: 90,
			want:  true,
		},
		{
			input: 50,
			want:  true,
		},
		{
			input: 54,
			want:  true,
		},
		{
			input: 111,
			want:  false,
		},
	}

	values := [...]int{100, 50, 150, 160, 170, 180, 90, 80, 70, 50, 60, 54, 32, 12, 5, 1}
	tree := New()
	for _, val := range values {
		tree.Insert(val, val, testIsSmaller)
	}

	for _, test := range tests {
		if ret := tree.Exists(test.input); ret != test.want {
			t.Errorf("Test for %d was %t expected to be %t", test.input, ret, test.want)
		}
	}

}

func TestCommon(t *testing.T) {
	valuesA := [...]int{20, 100, 50, 150, 160, 170, 180, 90, 80, 70, 50, 60, 54, 32, 12, 5, 1}
	valuesB := [...]int{20, 101, 51, 150, 160, 171, 182, 89, 80, 75, 53, 20, 1}
	valuesCommon := [...]int{20, 150, 160, 80, 1}
	treeA := New()
	treeB := New()

	for _, val := range valuesA {
		treeA.Insert(val, val, testIsSmaller)
	}

	for _, val := range valuesB {
		treeB.Insert(val, val, testIsSmaller)
	}

	common := treeA.Intersection(treeB)

	for _, val := range valuesCommon {
		if !common.Exists(val) {
			t.Errorf("Element %d not found in common elements tree\n", val)
		}
	}
}

func sliceEq(a []interface{}, b []int) bool {
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

func TestTopN(t *testing.T) {
	tests := []struct {
		values    [20]int
		topValues [6]int
		want      bool
	}{
		{
			values:    [...]int{1000, 20, 100, 5555, 50, 150, 2000, 160, 170, 180, 90, 80, 70, 50, 60, 54, 32, 12, 5, 1},
			topValues: [...]int{5555, 2000, 1000, 180, 170, 160},
			want:      true,
		},
		{
			values:    [...]int{57489, 2541, 5214, 2254, 2, 588, 98, 2874, 544, 98, 74, 22, 556, 14, 12, 23, 500, 532, 12, 15},
			topValues: [...]int{57489, 5214, 2874, 2541, 2254, 588},
			want:      true,
		},
	}

	for _, test := range tests {
		tree := New()
		for _, val := range test.values {
			tree.Insert(val, val, testIsSmaller)
		}

		res := tree.TopN(6)
		if sliceEq(res,
			test.topValues[:]) != test.want {
			t.Errorf("Tested: %v, got %v, wanted %v\n", test.values, res, test.topValues)
		}
	}
}
