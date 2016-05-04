/**
 * @file   queue_ms_lb.go
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
 * A simple lock-based queue.
**/

package dataset

import (
    "sync"
    "tools/share"
)

// -----------------------------------------------------------------------------

type node struct {
    key share.Key
    val share.Val
    next *node
}

type DataSet struct {
    head *node
    tail *node
    head_lock sync.Mutex
    tail_lock sync.Mutex
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
    set := new(DataSet)
    node := new_node(0, 0, nil)
    set.head = node
    set.tail = node
    return set
}

func (set *DataSet) Destroy() {
}

func (set *DataSet) Size() uint {
    size := uint(0)
    node := set.head
    for node.next != nil {
        size++
        node = node.next
    }
    return size
}

func (set *DataSet) Find(key share.Key) (share.Val, bool) {
    return 0, true
}

func (set *DataSet) Insert(key share.Key, val share.Val) bool {
    node := new_node(key, val, nil)
    set.tail_lock.Lock()
    defer set.tail_lock.Unlock()
    set.tail.next = node
    set.tail = node
    return true
}

func (set *DataSet) Delete(key share.Key) (share.Val, bool) {
    set.head_lock.Lock()
    defer set.head_lock.Unlock()
    node := set.head
    head_new := node.next
    if head_new == nil {
        return 0, false
    }
    set.head = head_new
    return head_new.val, true
}
