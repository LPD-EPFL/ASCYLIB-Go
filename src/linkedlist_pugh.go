/**
 * @file   linkedlist_pugh.go
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
 * William Pugh.
 * Concurrent Maintenance of Skip Lists.
 * Technical report, 1990.
**/

package dataset

import (
    "tools/share"
    "tools/ttas"
)

const (
    FindIsDef bool = true
    pugh_ro_fail = true
)

// -----------------------------------------------------------------------------

type node struct {
    key share.Key
    val share.Val
    next *node
    mutex ttas.Mutex
}

type DataSet struct {
    head *node
}

// -----------------------------------------------------------------------------

func (n *node) lock() {
    n.mutex.Lock()
}

func (n *node) unlock() {
    n.mutex.Unlock()
}

func new_node(key share.Key, val share.Val, next *node) *node {
    node := new(node)
    node.key = key
    node.val = val
    node.next = next
    return node
}

func (set *DataSet) search_weak_left(key share.Key) *node {
    pred := set.head
    succ := pred.next
    for succ.key < key {
        pred = succ
        succ = succ.next
    }
    return pred
}

func (set *DataSet) search_weak_right(key share.Key) *node {
    succ := set.head.next
    for succ.key < key {
        succ = succ.next
    }
    return succ
}

func (set *DataSet) search_strong(key share.Key) (pred *node, succ *node) {
    pred = set.search_weak_left(key)
    pred.lock()
    succ = pred.next
    for succ.key < key {
        pred.unlock()
        pred = succ
        pred.lock()
        succ = pred.next
    }
    return
}

func (set *DataSet) search_strong_cond(key share.Key, equal bool) (pred *node, succ *node, ok bool) {
    pred = set.search_weak_left(key)
    succ = pred.next
    if (succ.key == key) == equal {
        return nil, nil, false
    }
    pred.lock()
    succ = pred.next
    for succ.key < key {
        pred.unlock()
        pred = succ
        pred.lock()
        succ = pred.next
    }
    ok = true
    return
}

// -----------------------------------------------------------------------------

func New() *DataSet {
    set := new(DataSet)
    max := new_node(share.KEY_MAX, 0, nil)
    min := new_node(share.KEY_MIN, 0, max)
    set.head = min
    return set
}

func (set *DataSet) Destroy() {
}

func (set *DataSet) Size() uint {
    var size uint = 0
    node := set.head.next // We have at least 2 elements
    for node.next != nil {
        size++
        node = node.next
    }
    return size
}

func (set *DataSet) Find(key share.Key) (share.Val, bool) {
    right := set.search_weak_right(key)
    if right.key == key {
        return right.val, true
    }
    return 0, false
}

func (set *DataSet) Insert(key share.Key, val share.Val) bool {
    result := true
    var left, right *node
    // Optimize for step-wise strong search: if found, return before locking!
    if pugh_ro_fail {
        var ok bool
        left, right, ok = set.search_strong_cond(key, true)
        if !ok {
            return false
        }
    } else {
        left, right = set.search_strong(key)
    }
    if right.key == key {
        result = false
    } else {
        left.next = new_node(key, val, left.next)
    }
    left.unlock()
    return result
}

func (set *DataSet) Delete(key share.Key) (result share.Val, ok bool) {
    var left, right *node
    if pugh_ro_fail {
        left, right, ok = set.search_strong_cond(key, false)
        if !ok {
            return
        }
    } else {
        left, right = set.search_strong(key)
    }
    if right.key == key {
      right.lock()
      result = right.val
      left.next = right.next
      right.next = left
      right.unlock()
    }
    left.unlock()
    ok = true
    return
}
