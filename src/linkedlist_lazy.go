/**
 * @file   linkedlist_lazy.go
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
 * A Lazy Concurrent List-Based Set Algorithm,
 * S. Heller, M. Herlihy, V. Luchangco, M. Moir, W.N. Scherer III, N. Shavit
 * p.3-16, OPODIS 2005
**/

package dataset

import (
    "sync"
    "sync/atomic"
    "tools/share"
    "unsafe"
)

const (
    FindIsDef bool = true
    lazy_ro_fail = true
)

// -----------------------------------------------------------------------------

type node struct {
    key share.Key
    val share.Val
    next *node
    marked bool
    mutex sync.Mutex
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
    node := new(node) // No allocation failure test to do, and we cannot recover from an "OOM panic" (see http://stackoverflow.com/questions/30577308/golang-cannot-recover-from-out-of-memory-crash)
    node.key = key
    node.val = val
    node.next = next
    node.marked = false
    return node
}

func validate(pred *node, curr *node) bool {
    return !pred.marked && !curr.marked && pred.next == curr
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
    var size uint
    var node *node
    /* We have at least 2 elements */
    node = set.head.next
    for node.next != nil {
        size++
        node = node.next
    }
    return size
}

func (set *DataSet) Find(key share.Key) (share.Val, bool) {
    curr := set.head
    for curr.key < key {
        curr = curr.next
    }
    if curr.key == key && !curr.marked {
        return curr.val, true
    }
    return 0, false
}

func (set *DataSet) Insert(key share.Key, val share.Val) bool {
    var curr *node
    var pred *node
    var newnode *node
    for {
        // PARSE_TRY()
        pred = set.head
        curr = pred.next
        for curr.key < key {
            pred = curr
            curr = curr.next
        }
        // UPDATE_TRY()
        if lazy_ro_fail {
            if curr.key == key {
                if curr.marked {
                    continue
                }
                return false
            }
        }
        {
            pred.lock()
            if validate(pred, curr) {
                if curr.key == key {
                    pred.unlock()
                    return false
                } else {
                    newnode = new_node(key, val, curr)
                    atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&pred.next)), unsafe.Pointer(newnode))
                    pred.unlock()
                    return true
                }
            }
            pred.unlock()
        }
    }
}

func (set *DataSet) Delete(key share.Key) (result share.Val, ok bool) {
    var pred *node
    var curr *node
    var done bool = false
    for !done {
        pred = set.head
        curr = pred.next
        for curr.key < key {
            pred = curr
            curr = curr.next
        }
        if lazy_ro_fail {
            if curr.key != key {
                return
            }
        }
        {
            pred.lock()
            curr.lock()
            if validate(pred, curr) {
                if key == curr.key {
                    result, ok = curr.val, true
                    var c_nxt *node = curr.next
                    curr.marked = true
                    pred.next = c_nxt
                }
                done = true
            }
            curr.unlock()
            pred.unlock()
        }
    }
    return
}
