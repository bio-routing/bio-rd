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

// Package avltree provides an universal AVL tree
package avltree

import (
	"fmt"
	"sync"

	"github.com/golang/glog"
)

// Comparable is an interface used to pass compare functions to this avltree
type Comparable func(c1 interface{}, c2 interface{}) bool

// EachFunc is an interface used to pass a function to the each() method
type EachFunc func(node *TreeNode, vals ...interface{})

// Tree represents a tree
type Tree struct {
	root  *TreeNode
	lock  sync.RWMutex
	Count int
}

// TreeNode represents a node in a tree
type TreeNode struct {
	left      *TreeNode
	right     *TreeNode
	key       interface{}
	Values    []interface{}
	height    int64
	issmaller Comparable
}

// max gives returns the maximum of a and b
func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// getHeight return the height of tree with root `root`
func (root *TreeNode) getHeight() int64 {
	if root != nil {
		return root.height
	}
	return -1
}

// TreeMinValueNode returns the node with the minimal key in the tree
func (root *TreeNode) minValueNode() *TreeNode {
	for root.left != nil {
		return root.left.minValueNode()
	}
	return nil
}

// search searches element with key `key` in tree with root `root`
// in case the searched element doesn't exist nil is returned
func (root *TreeNode) search(key interface{}) *TreeNode {
	if root.key == key {
		return root
	}

	if root.issmaller(key, root.key) {
		if root.left == nil {
			return nil
		}
		return root.left.search(key)
	}

	if root.right == nil {
		return nil
	}
	return root.right.search(key)
}

// getBalance return difference of height of left and right
// subtrees of tree with root `root`
func (root *TreeNode) getBalance() int64 {
	if root == nil {
		return 0
	}
	return root.left.getHeight() - root.right.getHeight()
}

// leftRotate rotates tree with root `root` to the left and
// returns it's new root
func (root *TreeNode) leftRotate() *TreeNode {
	node := root.right
	root.right = node.left
	node.left = root

	root.height = max(root.left.getHeight(), root.right.getHeight()) + 1
	node.height = max(node.right.getHeight(), node.left.getHeight()) + 1
	return node
}

// leftRightRotate performs a left-right rotation of tree with root
// `root` and returns it's new root
func (root *TreeNode) leftRightRotate() *TreeNode {
	root.left = root.left.leftRotate()
	root = root.rightRotate()
	return root
}

// rightRotate performs a right rotation of tree with root `root`
// and returns it's new root
func (root *TreeNode) rightRotate() *TreeNode {
	node := root.left
	root.left = node.right
	node.right = root
	root.height = max(root.left.getHeight(), root.right.getHeight()) + 1
	node.height = max(node.left.getHeight(), node.right.getHeight()) + 1
	return node
}

// rightLeftRotate preforms a right-left rotation of tree with root
// `root` and returns it's new root
func (root *TreeNode) rightLeftRotate() *TreeNode {
	root.right = root.right.rightRotate()
	root = root.leftRotate()
	return root
}

// delete deletes node with key `key` from tree. If necessary rebalancing
// is done and new root is returned
func (root *TreeNode) delete(key interface{}) *TreeNode {
	if root == nil {
		return nil
	}

	if root.issmaller(key, root.key) {
		root.left = root.left.delete(key)
	} else if key == root.key {
		if root.left == nil && root.right == nil {
			return nil
		} else if root.left == nil {
			return root.left
		} else if root.right == nil {
			return root.right
		}

		tmp := root.minValueNode()
		root.key = tmp.key
		root.Values = tmp.Values
		root.right = root.right.delete(tmp.key)

		root.height = max(root.left.getHeight(), root.right.getHeight()) + 1
		balance := root.getBalance()
		if balance > 1 {
			if root.left.getBalance() >= 0 {
				return root.rightRotate()
			}
			return root.leftRightRotate()
		} else if balance < -1 {
			if root.right.getBalance() <= 0 {
				return root.leftRotate()
			}
			return root.rightLeftRotate()
		}
	} else {
		root.right = root.right.delete(key)
	}

	return root
}

// isEqual is a generic function that compares a and b of any comparable type
// return true if a and b are equal, otherwise false
func isEqual(a interface{}, b interface{}) bool {
	return a == b
}

// New simply returns a new (empty) tree
func New() *Tree {
	return &Tree{}
}

// Insert inserts an element to tree with root `t`
func (t *Tree) Insert(key interface{}, value interface{}, issmaller Comparable) (new *TreeNode, err error) {
	if t == nil {
		return nil, fmt.Errorf("unable to insert into nil tree")
	}
	t.lock.Lock()
	defer t.lock.Unlock()
	t.root, new = t.root.insert(key, value, issmaller)
	t.Count++
	return new, nil
}

// insert inserts an element into tree with root `root`
func (root *TreeNode) insert(key interface{}, value interface{}, issmaller Comparable) (*TreeNode, *TreeNode) {
	if root == nil {
		root = &TreeNode{
			left:  nil,
			right: nil,
			key:   key,
			Values: []interface{}{
				value,
			},
			height:    0,
			issmaller: issmaller,
		}
		return root, root
	}

	if isEqual(key, root.key) {
		root.Values = append(root.Values, value)
		return root, root
	}

	var new *TreeNode
	if root.issmaller(key, root.key) {
		root.left, new = root.left.insert(key, value, issmaller)
		if root.left.getHeight()-root.right.getHeight() == 2 {
			if root.issmaller(key, root.left.key) {
				root = root.rightRotate()
			} else {
				root = root.leftRightRotate()
			}
		}
	} else {
		root.right, new = root.right.insert(key, value, issmaller)
		if root.right.getHeight()-root.left.getHeight() == 2 {
			if (!root.issmaller(key, root.right.key)) && !isEqual(key, root.right.key) {
				root = root.leftRotate()
			} else {
				root = root.rightLeftRotate()
			}
		}
	}

	root.height = max(root.left.getHeight(), root.right.getHeight()) + 1
	return root, new
}

// Exists checks if a node with key `key` exists in tree `t`
func (t *Tree) Exists(key interface{}) bool {
	if t == nil {
		return false
	}
	return t.root.exists(key)
}

// exists recursively searches through tree with root `root` for element with
// key `key`
func (root *TreeNode) exists(key interface{}) bool {
	if root == nil {
		return false
	}

	if isEqual(key, root.key) {
		return true
	}

	if root.issmaller(key, root.key) {
		if root.left == nil {
			return false
		}
		return root.left.exists(key)
	}
	if root.right == nil {
		return false
	}
	return root.right.exists(key)
}

// Intersection finds common elements in trees `t` and `x` and returns them in a new tree
func (t *Tree) Intersection(x *Tree) (res *Tree) {
	if t == nil || x == nil {
		return nil
	}
	res = New()
	t.lock.RLock()
	x.lock.RLock()
	defer t.lock.RUnlock()
	defer x.lock.RUnlock()

	n := 0
	newRoot := t.root.intersection(x.root, res.root, &n)

	return &Tree{
		root:  newRoot,
		Count: n,
	}
}

// Intersection builds a tree of common elements of all trees in `candidates`
func Intersection(candidates []*Tree) (res *Tree) {
	n := len(candidates)
	if n == 0 {
		return nil
	}

	if n == 1 {
		return candidates[0]
	}

	chA := make([]chan *Tree, n/2)
	chB := make([]chan *Tree, n/2)
	chRet := make([]chan *Tree, n/2)

	// Start a go routine that builds intersection of each pair of candidates
	for i := 0; i < n/2; i++ {
		chA[i] = make(chan *Tree)
		chB[i] = make(chan *Tree)
		chRet[i] = make(chan *Tree)
		go func(chA chan *Tree, chB chan *Tree, chRes chan *Tree) {
			a := <-chA
			b := <-chB
			if a == nil || b == nil {
				chRes <- nil
				return
			}

			glog.Infof("finding common elements in %d and %d elements", a.Count, b.Count)
			chRes <- a.Intersection(b)
		}(chA[i], chB[i], chRet[i])
		chA[i] <- candidates[i*2]
		chB[i] <- candidates[i*2+1]
	}

	results := make([]*Tree, 0)

	// If amount of candidate trees is uneven we have to add last tree to results
	if n%2 == 1 {
		results = append(results, candidates[n-1])
	}

	// Fetch results
	for i := 0; i < n/2; i++ {
		results = append(results, <-chRet[i])
	}

	// If we only have one tree left over, we're done
	if len(results) != 1 {
		return Intersection(results)
	}

	return results[0]
}

// intersection recursively finds common elements in tree with roots `root` and `b`
// and returns the result in a new tree
func (root *TreeNode) intersection(b *TreeNode, res *TreeNode, n *int) *TreeNode {
	if root == nil || b == nil {
		return res
	}

	if root.left != nil {
		res = root.left.intersection(b, res, n)
	}
	if root.right != nil {
		res = root.right.intersection(b, res, n)
	}
	if b.exists(root.key) {
		res, _ = res.insert(root.key, root.key, root.issmaller)
		*n++
	}

	return res
}

// Each can be used to traverse tree `t` and call function f with params vals...
// for each node in the tree
func (t *Tree) Each(f EachFunc, vals ...interface{}) {
	if t == nil {
		return
	}

	t.lock.RLock()
	defer t.lock.RUnlock()
	t.root.Each(f, vals...)
}

// Each recursively traverses tree `tree` and calls functions f with params vals...
// for each node in the tree
func (root *TreeNode) Each(f EachFunc, vals ...interface{}) {
	if root == nil {
		return
	}
	f(root, vals...)
	if root.left != nil {
		root.left.Each(f, vals...)
	}
	if root.right != nil {
		root.right.Each(f, vals...)
	}
}

// Dump dumps tree `t` into a slice and returns it
func (t *Tree) Dump() (res []interface{}) {
	if t == nil {
		return
	}

	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.root.dump()
}

// dump recursively dumps all nodes of tree with root `t` into a slice and returns it
func (root *TreeNode) dump() (res []interface{}) {
	if root == nil {
		return res
	}
	if root.left != nil {
		tmp := root.left.dump()
		res = append(res, tmp...)
	}
	res = append(res, root.Values...)
	if root.right != nil {
		tmp := root.right.dump()
		res = append(res, tmp...)
	}
	return res
}

// TopN finds the the `n` biggest elements in tree `t` and returns them in a slice
func (t *Tree) TopN(n int) (res []interface{}) {
	if t == nil {
		return
	}

	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.root.topN(n)
}

// topN recursively traverses tree with root `t` to find biggest `n` elements.
// Top elements are returned as a slice
func (root *TreeNode) topN(n int) (res []interface{}) {
	if root == nil {
		return res
	}

	if root.right != nil {
		tmp := root.right.topN(n)
		for _, k := range tmp {
			if len(res) == n {
				return res
			}

			res = append(res, k)
		}
	}

	if len(res) < n {
		res = append(res, root.Values...)
	}

	if len(res) == n {
		return res
	}

	if root.left != nil {
		tmp := root.left.topN(n - len(res))
		for _, k := range tmp {
			if len(res) == n {
				return res
			}
			res = append(res, k)
		}
	}
	return res
}
