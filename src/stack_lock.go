/**
 * @file   stack_lock.go
 * @author Sébastien Rouault <sebastien.rouault@epfl.ch>
 *
 * @section LICENSE
 *
 * Copyright (c) 2014 Vasileios Trigonakis <vasileios.trigonakis@epfl.ch>,
 *                    Tudor David <tudor.david@epfl.ch>
 *                    Distributed Programming Lab (LPD), EPFL
 *
 * ASCYLIB is free software: you can redistribute it and/or
 * modify it under the terms of the GNU General Public License
 * as published by the Free Software Foundation, version 2
 * of the License.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * @section DESCRIPTION
 *
 * A very simple stack implementation (lock-based).
**/

package dataset

import (
    "sync"
    "tools/share"
)

const (
    FindIsDef bool = false
)

// -----------------------------------------------------------------------------

type node struct {
  key share.Key
  val share.Val
  next *node
}

type DataSet struct {
    top *node
    lock sync.Mutex
}

// -----------------------------------------------------------------------------

func new_node(key share.Key, val share.Val, next *node) *node {
    elem := new(node)
    elem.key = key
    elem.val = val
    elem.next = next
    return elem
}

// -----------------------------------------------------------------------------

func New() *DataSet {
    return new(DataSet) // 0 initialized by default
}

func (set *DataSet) Destroy() {
}

func (set *DataSet) Size() uint {
    size := uint(0)
    node := set.top
    for node != nil {
        size++
        node = node.next
    }
    return size
}

func (set *DataSet) Find(key share.Key) (share.Val, bool) {
    return 0, true // Not supposed to use Find with a stack...
}

func (set *DataSet) Insert(key share.Key, val share.Val) bool {
    elem := new_node(key, val, nil)
    set.lock.Lock()
    defer set.lock.Unlock()
    elem.next = set.top
    set.top = elem
    return true
}

func (set *DataSet) Delete(key share.Key) (share.Val, bool) {
    set.lock.Lock()
    defer set.lock.Unlock()
    top := set.top
    if top == nil {
        return 0, false
    }
    set.top = top.next
    return top.val, true
}
